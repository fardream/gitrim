package svc

import (
	"context"
	"encoding/hex"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var ErrStatusNotFound = status.Error(codes.NotFound, "repo sync not found for the provided id")

// getRepoSync returns the repo sync, the id, and the error
func getRepoSync(db *bbolt.DB, idHex string, requireExist bool) (*DbRepoSync, []byte, error) {
	id, err := hex.DecodeString(idHex)
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "failed to parse id: %s", err.Error())
	}

	r, err := getRepoSyncFromDb(db, id)
	if err != nil {
		return nil, id, err
	}
	if requireExist && r == nil {
		return nil, id, ErrStatusNotFound
	}

	return r, id, nil
}

func (svc *Svc) GetRepoSync(
	ctx context.Context,
	request *GetRepoSyncRequest,
) (*GetRepoSyncResponse, error) {
	idHex := request.Id
	id, err := hex.DecodeString(idHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse id: %s", err.Error())
	}

	rs, err := getRepoSyncFromDb(svc.db, id)
	if err != nil {
		return nil, err
	}
	if rs == nil {
		return nil, ErrStatusNotFound
	}
	secret, err := getSecretForId(svc.db, id)
	if err != nil {
		return nil, err
	}

	result := &GetRepoSyncResponse{
		RepoSync: rs.SyncData,
		Secret:   hex.EncodeToString(secret),
		SyncStat: rs.Stat,
	}

	return result, nil
}

func getRepoSyncFromDb(
	db *bbolt.DB,
	id []byte,
) (*DbRepoSync, error) {
	return getFromDb(
		db,
		[]byte(REPO_SYNC_BUCKET),
		id,
		func(d []byte, v *DbRepoSync) error {
			return proto.Unmarshal(d, v)
		})
}
