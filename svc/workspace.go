package svc

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/fardream/gitrim"
)

// workspace contains the repo, the branch, and the memory storage for one repo.
type workspace struct {
	// storage
	storage *memory.Storage
	// repo id
	repoId *GitRepoIdentifier
	// branch
	branch string
	// repo
	repo *git.Repository
	// is empty indicates if the repo or branch is empty
	isempty bool

	branchhead *object.Commit

	auth *http.BasicAuth
}

func (c *RemoteConfig) auth() *http.BasicAuth {
	if c.Username == "" {
		return nil
	}

	return &http.BasicAuth{
		Username: c.Username,
		Password: c.Secret,
	}
}

func constructPullUrl(cfg *RemoteConfig, owner string, repo string) (string, error) {
	return fmt.Sprintf("%s/%s/%s", cfg.RemoteUrl, owner, repo), nil
}

const (
	refSpecSingleBranchRemote = "+refs/heads/%s:refs/remotes/%s/%[1]s"
	refSpecSingleBranchPush   = "+refs/heads/%s:refs/heads/%[1]s"
	remotename                = "origin"
)

var ErrEmptyRemoteConfig = errors.New("empty remote config")

// newWorkspace creates a new workspace.
//
// The process first validate the configurations, then
// from the configuration construct the remote URL.
// It initializes an empty git repo in in-memory storage,
// adds the remote as origin, and fetches the remote branch.
// fetch may fail if the remote is empty.
// If the repo is not empty, a local branch will be set up with branch name
// with the remote if the remote branch exists.
func newWorkspace(
	ctx context.Context,
	configmap map[string]*RemoteConfig,
	id *GitRepoIdentifier,
	branch string,
) (*workspace, error) {
	// check the config
	if configmap == nil {
		return nil, ErrEmptyRemoteConfig
	}
	remoteConfig, found := configmap[id.RemoteName]
	if !found {
		return nil, fmt.Errorf("unknown remote: %s", id.RemoteName)
	}

	// figure out the url
	url, err := constructPullUrl(remoteConfig, id.Owner, id.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to construct url for remote repo: %w", err)
	}

	// storage
	storage := memory.NewStorage()

	logger.Info("cloning repo", "remote", url, "branch", branch)

	// init a repo
	repo, err := git.InitWithOptions(
		storage,
		nil,
		git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName(branch),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain init: %w", err)
	}

	// add remote
	_, err = repo.CreateRemote(
		&config.RemoteConfig{
			Name: remotename,
			URLs: []string{url},
			Fetch: []config.RefSpec{
				config.RefSpec(
					fmt.Sprintf(refSpecSingleBranchRemote, branch, remotename),
				),
			},
			Mirror: true,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create remote for origin: %w", err)
	}

	w := &workspace{
		storage: storage,
		repoId:  id,
		branch:  branch,
		repo:    repo,
		auth:    remoteConfig.auth(),
	}

	if err := w.fetch(ctx); err != nil {
		return nil, err
	}

	return w, nil
}

// sethead set the head of the branch
func (w *workspace) sethead() bool {
	setemptyandreturnfalse := func() bool {
		w.isempty = true
		return false
	}
	if w.isempty {
		return setemptyandreturnfalse()
	}

	remotebranch := plumbing.NewRemoteReferenceName(remotename, w.branch)
	r, err := w.storage.Reference(remotebranch)
	if err != nil || r.Hash().IsZero() {
		return setemptyandreturnfalse()
	}
	err = w.storage.SetReference(
		plumbing.NewHashReference(
			plumbing.NewBranchReferenceName(w.branch),
			r.Hash()))

	if err != nil {
		logger.Warn("cannot set local branch", "branch", w.branch)
		return setemptyandreturnfalse()
	}

	head, err := object.GetCommit(w.storage, r.Hash())
	if err != nil {
		logger.Warn("failed to get the head commit", "branch", w.branch, "commit", r.Hash().String())
		return setemptyandreturnfalse()
	}

	w.branchhead = head

	return true
}

func isNoMatchingRef(err error) bool {
	var v git.NoMatchingRefSpecError
	return errors.As(err, &v)
}

// fetch the branch from remote and set the branch to
func (w *workspace) fetch(ctx context.Context) error {
	err := w.repo.FetchContext(ctx,
		&git.FetchOptions{
			Auth:       w.auth,
			RemoteName: remotename,
		})

	// check if the remote is empty.
	if err != nil && errors.Is(err, transport.ErrEmptyRemoteRepository) {
		logger.Warn("empty remote")
		w.isempty = true
	} else if err != nil && isNoMatchingRef(err) {
		logger.Warn("branch doesn't exist")
		w.isempty = true
	} else if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	// if the repo is not empty, try set the branch, since no local branch yet.
	w.sethead()

	return nil
}

// pushToRemote push the changes to the remote
func (w *workspace) pushToRemote(ctx context.Context, forcePush bool) error {
	refspec := config.RefSpec(fmt.Sprintf(refSpecSingleBranchPush, w.branch))
	err := w.repo.PushContext(
		ctx,
		&git.PushOptions{
			RemoteName: remotename,
			Auth:       w.auth,
			Force:      forcePush,
			RefSpecs:   []config.RefSpec{refspec},
		})
	isuptodate := errors.Is(err, git.NoErrAlreadyUpToDate)
	switch {
	case err != nil && !isuptodate:
		return fmt.Errorf("failed to update the remote: %w", err)
	case isuptodate:
		logger.Warn("remote already updated")
		fallthrough
	default:
		return nil
	}
}

func (wksp *workspace) updateBranchHead(toc *object.Commit) error {
	h := toc.Hash
	refname := plumbing.NewBranchReferenceName(wksp.branch)
	ref := plumbing.NewHashReference(refname, h)
	wksp.branchhead = toc
	if err := wksp.storage.SetReference(ref); err != nil {
		return fmt.Errorf("failed to set to branch in to repo: %w", err)
	}
	bheadref := plumbing.NewSymbolicReference(plumbing.HEAD, refname)
	if err := wksp.storage.SetReference(bheadref); err != nil {
		return fmt.Errorf("failed to set HEAD due to: %w", err)
	}
	return nil
}

func (wksp *workspace) getBranchHead() (*object.Commit, error) {
	if wksp.branchhead != nil {
		return wksp.branchhead, nil
	}
	branchref, err := wksp.storage.Reference(plumbing.NewBranchReferenceName(wksp.branch))
	if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to obtain branch %s head: %w", wksp.branch, err)
	}
	return object.GetCommit(wksp.storage, branchref.Hash())
}

func (wksp *workspace) getNewCommits(
	ctx context.Context,
	lastcommit plumbing.Hash,
	pastcommits gitrim.HashSet,
	islinear bool,
) ([]*object.Commit, error) {
	if wksp.isempty {
		return nil, nil
	}

	headcommit, err := wksp.getBranchHead()
	if err != nil {
		return nil, err
	}

	var hist []*object.Commit

	if islinear {
		hist, err = gitrim.GetLinearHistory(ctx, headcommit, lastcommit, 0)
	} else {
		hist, err = gitrim.GetDFSPath(ctx, headcommit, pastcommits, 0)
	}

	if err != nil {
		return nil, err
	}

	return hist, nil
}
