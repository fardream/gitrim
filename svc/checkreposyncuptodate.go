package svc

import "context"

func (s *Svc) CheckRepoSyncUpToDate(
	ctx context.Context,
	req *CheckRepoSyncUpToDateRequest,
) (*CheckRepoSyncUpToDateResponse, error) {
	sw, err := loadSyncWorkspaceFromDb(ctx, s.config.Remotes, req.Id, s.db, true)
	if err != nil {
		return nil, err
	}
	return &CheckRepoSyncUpToDateResponse{
		FromRepoStatus: sw.fromStatus,
		ToRepoStatus:   sw.toStatus,
	}, nil
}
