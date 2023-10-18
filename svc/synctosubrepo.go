package svc

import (
	"context"
	"encoding/hex"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Svc) SyncToSubRepo(ctx context.Context, request *SyncToSubRepoRequest) (*SyncToSubRepoResponse, error) {
	idHex := request.Id
	id, err := hex.DecodeString(idHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse id: %s", err.Error())
	}

	reposync, err := getRepoSyncFromDb(s.db, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain repo sync from db: %s", err.Error())
	}
	if reposync == nil {
		return nil, ErrStatusNotFound
	}

	fromwksp, err := s.newWorkspace(ctx, reposync.FromRepo, reposync.FromBranch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain from repo: %s", err.Error())
	}
	towksp, err := s.newWorkspace(ctx, reposync.ToRepo, reposync.ToBranch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain to repo: %s", err.Error())
	}

	info, err := syncWksp(ctx, fromwksp, reposync.FromBranch, towksp, reposync.ToBranch, reposync.Filter.CanonicalFilters, reposync.RootCommits, 0)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, ErrStatusEmptyFromRepo
	}

	reposync.LastSyncFromCommit = info.LastSyncFromCommit
	reposync.LastSyncToCommit = info.LastSyncToCommit

	if err := s.db.Update(func(tx *bbolt.Tx) error {
		return putRepoSyncFunc(id[:], reposync)(tx)
	}); err != nil {
		return nil, err
	}

	if err := s.db.Sync(); err != nil {
		return nil, ErrStatusDBFailure
	}

	return &SyncToSubRepoResponse{}, nil
}
