package svc

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *Svc) GetRepoSync(ctx context.Context, request *GetRepoSyncRequest) (*GetRepoSyncResponse, error) {
	idHex := request.Id
	id, err := hex.DecodeString(idHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse id: %s", err.Error())
	}

	rs, err := getRepoSyncFromDb(svc.db, id)
	if err != nil {
		return nil, err
	}
	secret, err := getSecretForId(svc.db, id)
	if err != nil {
		return nil, err
	}

	result := &GetRepoSyncResponse{
		RepoSync: rs,
		Secret:   hex.EncodeToString(secret),
	}

	return result, nil
}
