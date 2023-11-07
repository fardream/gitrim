package svc

import (
	"context"
	"encoding/hex"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/fardream/gitrim"
)

func (s *Svc) CommitsFromSubRepo(ctx context.Context, req *CommitsFromSubRepoRequest) (*CommitsFromSubRepoResponse, error) {
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

	resp := &CommitsFromSubRepoResponse{
		Result:           status,
		FromRepoStatus:   sw.fromStatus,
		ToRepoStatus:     sw.toStatus,
		HasGpgSignatures: isgpg,
		RejectedFiles:    rejectedfiles,
	}

	if status != SubRepoCommitsCheck_CHECK_PASSED {
		return resp, nil
	}

	dopush := !HasOverrides(req) && req.DoPush
	if dopush {
		idwaiter, err := s.lockId(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		defer s.unlockId(req.Id, idwaiter)
	}
	newcommits, err := sw.syncToFrom(ctx, dopush, req.AllowPgpSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to push: %w", err)
	}

	if dopush {
		id, err := hex.DecodeString(req.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to decode id: %w", err)
		}
		if err := s.db.Update(func(tx *bbolt.Tx) error {
			return putRepoSyncFunc(id, sw.db)(tx)
		}); err != nil {
			return nil, err
		}
		if err := s.db.Sync(); err != nil {
			return nil, ErrStatusDBFailure
		}
	} else {
		logger.Info("not updating due to has override or not pushing", "has-override", HasOverrides(req), "do-push", req.DoPush)
	}

	for _, nc := range newcommits {
		resp.NewCommits = append(resp.NewCommits, nc.Hash.String())
	}

	return resp, nil
}
