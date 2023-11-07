package svc

import (
	"context"
	"encoding/hex"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrStatusEmptyFromRepo = status.Error(codes.InvalidArgument, "empty from repo")
	ErrStatusDBFailure     = status.Error(codes.Internal, "DB failure")
)

func (s *Svc) SyncToSubRepo(ctx context.Context, request *SyncToSubRepoRequest) (*SyncToSubRepoResponse, error) {
	ws, err := loadSyncWorkspaceFroReq(ctx, s.config.Remotes, s.db, request, true)
	if err != nil {
		return nil, err
	}

	originalhead := ws.db.Stat.LastSyncToCommit
	id, err := hex.DecodeString(request.Id)
	if err != nil {
		return nil, err
	}

	// we need to lock the id if the repos will be updated.
	if !HasOverrides(request) {
		idwaiter, err := s.lockId(ctx, request.Id)
		if err != nil {
			return nil, err
		}
		defer s.unlockId(request.Id, idwaiter)
	}

	newcommits, err := ws.syncToTo(ctx, request.GetForce())
	if err != nil {
		return nil, err
	}

	if !HasOverrides(request) {
		if err := s.db.Update(func(tx *bbolt.Tx) error {
			return putRepoSyncFunc(id, ws.db)(tx)
		}); err != nil {
			return nil, err
		}

		if err := s.db.Sync(); err != nil {
			return nil, ErrStatusDBFailure
		}
	} else {
		logger.Info("not updating due to override", "id", request.Id)
	}

	return &SyncToSubRepoResponse{
		NumberOfNewCommits: int32(len(newcommits)),
		OriginalHead:       originalhead,
		NewHead:            ws.db.Stat.LastSyncToCommit,
	}, nil
}
