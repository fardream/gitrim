package svc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewRepoSyncId creates a new id for the repo sync.
func NewRepoSyncId(
	fromRepo *GitRepoIdentifier,
	fromBranch string,
	toRepo *GitRepoIdentifier,
	toBranch string,
) []byte {
	r := sha256.Sum256(
		([]byte)(fmt.Sprintf(
			"%s-%s-%s-%s-%s-%s-%s-%s",
			fromRepo.RemoteName, fromRepo.Owner, fromRepo.Repo,
			fromBranch,
			toRepo.RemoteName, toRepo.Owner, toRepo.Repo,
			toBranch)))
	return r[:]
}

var ErrEmptyFilter = errors.New("empty filter")

func (s *Svc) InitRepoSync(
	ctx context.Context,
	req *InitRepoSyncRequest,
) (*InitRepoSyncResponse, error) {
	// check if the repo sync request is valid
	err := s.verifyInitRepoSyncRequest(req)
	if err != nil {
		return nil, err
	}
	// create id
	id := NewRepoSyncId(req.FromRepo, req.FromBranch, req.ToRepo, req.ToBranch)

	reposync, err := getRepoSyncFromDb(s.mustGetDb(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to check if the database is already created: %w", err)
	}

	if reposync != nil {
		return nil, fmt.Errorf("sync-ing between the two repos already exists")
	}

	idstr := hex.EncodeToString(id[:])
	secret, err := newSecret(s.encryptor, id[:])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate secret: %s", err.Error())
	}

	resp := &InitRepoSyncResponse{
		Id:     hex.EncodeToString(id[:]),
		Secret: hex.EncodeToString(secret),
	}

	// create the canonical filter
	filter, err := NewCanonicalFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	if len(filter.CanonicalFilters) == 0 {
		return nil, ErrEmptyFilter
	}

	reposync = &DbRepoSync{
		SyncData: &RepoSync{
			Id:         idstr,
			FromRepo:   req.FromRepo,
			FromBranch: req.FromBranch,
			ToRepo:     req.ToRepo,
			ToBranch:   req.ToBranch,
			Filter:     filter,
		},
		Stat: EmptySyncStat(),
	}

	ws, err := newSyncWorkspace(ctx, s.config.Remotes, reposync)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain from repo: %s", err.Error())
	}

	if _, err := ws.syncToTo(ctx, true); err != nil {
		return nil, err
	}

	if err := s.db.Update(func(tx *bbolt.Tx) error {
		if err := putSecretFunc(id[:], secret)(tx); err != nil {
			return err
		}

		return putRepoSyncFunc(id[:], reposync)(tx)
	}); err != nil {
		return nil, err
	}

	if err := s.db.Sync(); err != nil {
		return nil, ErrStatusDBFailure
	}

	return resp, nil
}
