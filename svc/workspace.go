package svc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
func (s *svc) newWorkspace(
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

	slog.Info("cloning repo", "remote", url, "branch", branch)

	// init a repo and fetch the remotes
	repo, err := git.InitWithOptions(
		storage,
		nil,
		git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName(branch),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain init: %w", err)
	}

	spec := fmt.Sprintf(refSpecSingleBranch, branch, "origin")
	slog.Info("fetch-spec", "spec", spec)

	// add remote
	_, err = repo.CreateRemote(
		&config.RemoteConfig{
			Name:   remotename,
			URLs:   []string{url},
			Fetch:  []config.RefSpec{config.RefSpec(spec)},
			Mirror: true,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create remote for origin: %w", err)
	}

	// fetch
	auth := remoteConfig.auth()

	isempty := false

	err = repo.FetchContext(ctx, &git.FetchOptions{
		Auth:       auth,
		RemoteName: remotename,
		RefSpecs:   []config.RefSpec{config.RefSpec(spec)},
	})
	// check if the remote is empty.
	if err != nil && errors.Is(err, transport.ErrEmptyRemoteRepository) {
		slog.Warn("empty remote")
		isempty = true
	} else if err != nil {
		return nil, fmt.Errorf("failed to clone: %w", err)
	}

	// if the repo is not empty, try set the branch, since no local branch yet.
	if !isempty {
		remotebranch := plumbing.NewRemoteReferenceName(remotename, branch)
		r, err := storage.Reference(remotebranch)
		if err == nil && !r.Hash().IsZero() {
			if err := storage.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName(branch), r.Hash())); err != nil {
				slog.Warn("cannot set local branch", "branch", branch)
			}
		}
	}

	return &workspace{
		storage: storage,
		repoId:  id,
		branch:  branch,
		repo:    repo,
		isempty: isempty,
		auth:    auth,
	}, nil
}

func (w *workspace) push(ctx context.Context) error {
	return w.repo.PushContext(
		ctx,
		&git.PushOptions{
			Auth: w.auth,
		})
}
