package svc

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
