package svc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newRepoSyncId(fromRepo *GitRepoIdentifier, fromBranch string, toRepo *GitRepoIdentifier, toBranch string) (string, error) {
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s-%s-%s-%s",
		fromRepo.RemoteName, fromRepo.Owner, fromRepo.Repo,
		fromBranch,
		toRepo.RemoteName, toRepo.Owner, toRepo.Repo,
		toBranch), nil
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
		return nil, status.Errorf(codes.Internal, "failed to obtain from repo: %s", err.Error())
	}
	towskp, err := s.newWorkspace(ctx, req.ToRepo, req.ToBranch)
	if err != nil {
		return nil, err
	}

	info, err := syncWksp(ctx, fromwksp, req.FromBranch, towskp, req.ToBranch, reposync.Filter.CanonicalFilters, req.RootCommits, req.MaxDepth)
	if err != nil {
		return nil, err
	}

	if info != nil {
		reposync.RootCommits = append(reposync.RootCommits, info.RootCommits...)
		reposync.InitHeadCommit = info.InitHeadCommit
		reposync.LastSyncFromCommit = info.LastSyncFromCommit
		reposync.LastSyncToCommit = info.LastSyncToCommit
	}

	if err := s.db.Update(func(tx *bbolt.Tx) error {
		if err := putSecretFunc(id[:], secret)(tx); err != nil {
			return err
		}

		return putRepoSyncFunc(id[:], reposync)(tx)
	}); err != nil {
		return nil, err
	}

	if err := s.db.Sync(); err != nil {
		return nil, ErrStatusDBFailure
	}

	return resp, nil
}
