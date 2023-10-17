package svc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.etcd.io/bbolt"

	"github.com/fardream/gitrim"
)

func newRepoSyncId(fromRepo *GitRepoIdentifier, fromBranch string, toRepo *GitRepoIdentifier, toBranch string) (string, error) {
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s-%s-%s-%s",
		fromRepo.RemoteName, fromRepo.Owner, fromRepo.Repo,
		fromBranch,
		toRepo.RemoteName, toRepo.Owner, toRepo.Repo,
		toBranch), nil
}

func verifyGitRepoIdentifier(repo *GitRepoIdentifier) error {
	if repo == nil {
		return ErrNilRepo
	}
	if repo.Owner == "" {
		return ErrEmptyParentName
	}
	if repo.Repo == "" {
		return ErrEmptyRepoName
	}
	if repo.RemoteName == "" {
		return ErrEmptyRemoteName
	}
	return nil
}

func (s *Svc) verifyInitRepoSyncRequest(req *InitRepoSyncRequest) error {
	var err error
	err = verifyGitRepoIdentifier(req.FromRepo)
	if err != nil {
		return err
	}
	err = verifyGitRepoIdentifier(req.ToRepo)
	if err != nil {
		return err
	}
	if req.FromBranch == "" || req.ToBranch == "" {
		return ErrEmptyBranchName
	}

	return nil
}

func (s *Svc) InitRepoSync(ctx context.Context, req *InitRepoSyncRequest) (*InitRepoSyncResponse, error) {
	err := s.verifyInitRepoSyncRequest(req)
	if err != nil {
		return nil, err
	}
	idraw, err := newRepoSyncId(req.FromRepo, req.FromBranch, req.ToRepo, req.ToBranch)
	if err != nil {
		return nil, err
	}
	id := sha256.Sum256([]byte(idraw))

	reposync, err := getRepoSyncFromDb(s.mustGetDb(), id[:])
	if err != nil {
		return nil, fmt.Errorf("failed to check if the database is already created: %w", err)
	}

	if reposync != nil {
		return nil, fmt.Errorf("sync-ing between the two repos already exists")
	}

	idstr := hex.EncodeToString(id[:])
	secret, err := newSecret(s.encryptor, id[:])

	resp := &InitRepoSyncResponse{
		Id:     hex.EncodeToString(id[:]),
		Secret: hex.EncodeToString(secret),
	}

	// create the canonical filter
	filter, err := NewCanonicalFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	if len(filter.CanonicalFilters) == 0 {
		return nil, ErrEmptyFilter
	}

	reposync = &RepoSync{
		Id:         idstr,
		FromRepo:   req.FromRepo,
		FromBranch: req.FromBranch,
		ToRepo:     req.ToRepo,
		ToBranch:   req.ToBranch,
		Filter:     filter,
	}

	fromwksp, err := s.newWorkspace(ctx, req.FromRepo, req.FromBranch)
	if err != nil {
		return nil, err
	}
	if fromwksp.isempty {
		return nil, ErrEmptyFromRepo
	}
	towskp, err := s.newWorkspace(ctx, req.ToRepo, req.ToBranch)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get from branch: %s", reposync.FromBranch)
	}

	frombranchref := plumbing.NewBranchReferenceName(fromwksp.branch)
	branch, err := fromwksp.storage.Reference(frombranchref)
	if err != nil {
		if iter, err := fromwksp.storage.IterReferences(); err == nil {
			iter.ForEach(func(r *plumbing.Reference) error {
				fmt.Printf("we have reference: %s\n", r)
				return nil
			})
		}
		return nil, fmt.Errorf("failed to find from branch %s in from repo: %w", frombranchref, err)
	}

	headcommit, err := object.GetCommit(fromwksp.storage, branch.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get head commit for from branch %s: %w", branch.Hash(), err)
	}

	rootcommits, err := gitrim.DecodeHashHexes(req.RootCommits...)
	if err != nil {
		return nil, err
	}

	commits, err := gitrim.GetDFSPath(ctx, headcommit, rootcommits, int(req.MaxDepth))
	if err != nil {
		return nil, fmt.Errorf("failed to obtain commit graph from from repo: %w", err)
	}

	roots := gitrim.GetRoots(commits)

	for _, r := range roots {
		reposync.RootCommits = append(reposync.RootCommits, r.String())
	}

	f, err := gitrim.NewOrFilterForPatterns(filter.CanonicalFilters...)
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}
	newcommits, err := gitrim.FilterDFSPath(ctx, commits, towskp.storage, f)
	if err != nil {
		return nil, fmt.Errorf("failed to filter commits according to filter: %w", err)
	}

	newhead := gitrim.LastNonNilCommit(newcommits)
	if newhead == nil {
		return nil, ErrFilteredRepoEmpty
	}

	h := newhead.Hash
	refname := plumbing.NewBranchReferenceName(towskp.branch)
	ref := plumbing.NewHashReference(refname, h)
	if err := towskp.storage.SetReference(ref); err != nil {
		return nil, fmt.Errorf("failed to set to branch in to repo: %w", err)
	}

	reposync.LastSyncFromCommit = commits[len(commits)-1].Hash.String()
	reposync.LastSyncToCommit = newhead.Hash.String()

	err = towskp.updateToLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to push: %w", err)
	}

	if err := s.db.Update(func(tx *bbolt.Tx) error {
		if err := putSecretFunc(id[:], secret)(tx); err != nil {
			return err
		}

		return putRepoSyncFunc(id[:], reposync)(tx)
	}); err != nil {
		return nil, err
	}

	s.db.Sync()

	return resp, nil
}
