package svc

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// workspace contains the repo, the branch, and the memory storage for this repo.
type workspace struct {
	storage *memory.Storage
	repoId  *GitRepoIdentifier
	branch  string
	repo    *git.Repository
	isempty bool
	auth    *http.BasicAuth
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
	refSpecSingleBranch = "+refs/heads/%s:refs/remotes/%s/%[1]s"
	remotename          = "origin"
)

// newWorkspace creates a new workspace.
//
// The process first validate the configurations, then
// from the configuration construct the remote URL.
// It initializes an empty git repo in in-memory storage,
// adds the remote as origin, and fetches the remote branch.
// fetch may fail if the remote is empty.
// If the repo is not empty, a local branch will be set up with branch name
// with the remote if the remote branch exists.
func (s *Svc) newWorkspace(
	ctx context.Context,
	id *GitRepoIdentifier,
	branch string,
) (*workspace, error) {
	// check the config
	if s.config.Remotes == nil {
		return nil, ErrEmptyRemoteConfig
	}
	remoteConfig, found := s.config.Remotes[id.RemoteName]
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
					fmt.Sprintf(refSpecSingleBranch, branch, remotename),
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

func (w *workspace) fetch(ctx context.Context) error {
	err := w.repo.FetchContext(ctx, &git.FetchOptions{
		Auth:       w.auth,
		RemoteName: remotename,
	})

	// check if the remote is empty.
	if err != nil && errors.Is(err, transport.ErrEmptyRemoteRepository) {
		logger.Warn("empty remote")
		w.isempty = true
	} else if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	// if the repo is not empty, try set the branch, since no local branch yet.
	if !w.isempty {
		remotebranch := plumbing.NewRemoteReferenceName(remotename, w.branch)
		r, err := w.storage.Reference(remotebranch)
		if err == nil && !r.Hash().IsZero() {
			if err := w.storage.SetReference(
				plumbing.NewHashReference(
					plumbing.NewBranchReferenceName(w.branch),
					r.Hash())); err != nil {
				logger.Warn("cannot set local branch", "branch", w.branch)
			}
		}
	}

	return nil
}

func (w *workspace) updateToLatest(ctx context.Context) error {
	err := w.repo.PushContext(
		ctx,
		&git.PushOptions{
			Auth: w.auth,
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
