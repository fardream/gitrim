package svc

import (
	"context"
	"slices"

	"github.com/fardream/gitrim"
)

func getRejectedFiles(results []*gitrim.FilePatchCheckResult) []string {
	var rejected []string

	for _, r := range results {
		if r == nil {
			continue
		}
		for _, e := range r.Errors {
			rejected = append(rejected, e.ErrorFiles()...)
		}
	}
	slices.Sort(rejected)
	return slices.Compact(rejected)
}

func (s *Svc) CheckCommitsFromSubRepo(
	ctx context.Context,
	req *CheckCommitsFromSubRepoRequest,
) (*CheckCommitsFromSubRepoResponse, error) {
	sw, err := loadSyncWorkspaceFroReq(ctx, s.config.Remotes, s.db, req, true)
	if err != nil {
		return nil, err
	}

	status := checkSubHistoryForSyncToFrom(sw.fromStatus, sw.toStatus)
	var rejectedfiles []string
	var fileerrors []*gitrim.FilePatchCheckResult
	var isgpg bool

	if status == SubRepoCommitsCheck_CHECK_PASSED {
		fileerrors, isgpg, err = sw.checkCommits(ctx)
		if err != nil {
			return nil, err
		}
		if isgpg && !req.AllowPgpSignature {
			status = SubRepoCommitsCheck_COMMITS_REJECTED
		}
		rejectedfiles = getRejectedFiles(fileerrors)
		if len(rejectedfiles) > 0 {
			status = SubRepoCommitsCheck_COMMITS_REJECTED
		}
	}

	return &CheckCommitsFromSubRepoResponse{
		Result:           status,
		FromRepoStatus:   sw.fromStatus,
		ToRepoStatus:     sw.toStatus,
		HasGpgSignatures: isgpg,
		RejectedFiles:    rejectedfiles,
	}, nil
}
