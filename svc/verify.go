package svc

import "errors"

var (
	ErrNilRepo         = errors.New("nil repo")
	ErrEmptyParentName = errors.New("empty owner name")
	ErrEmptyRepoName   = errors.New("empty repo name")
	ErrEmptyBranchName = errors.New("empty branch name")
	ErrEmptyRemoteName = errors.New("empty remote name")
)

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
