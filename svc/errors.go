// errors

package svc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Below codes are for non-grpc
var (
	ErrNilRepo                = errors.New("nil repo")
	ErrEmptyParentName        = errors.New("empty owner name")
	ErrEmptyRepoName          = errors.New("empty repo name")
	ErrEmptyBranchName        = errors.New("empty branch name")
	ErrEmptyRemoteName        = errors.New("empty remote name")
	ErrDuplicateRepoSync      = errors.New("duplicate repo sync")
	ErrEmptyFilter            = errors.New("empty filter")
	ErrEmptyRemoteConfig      = errors.New("empty remote config")
	ErrEmptyFromRepo          = errors.New("empty from repo")
	ErrInvalidCommitSHALength = errors.New("invalid commit sha length")
	ErrInvalidBranch          = errors.New("invalid branch")
	ErrNilDB                  = errors.New("nil db")
	ErrSaltGenError           = errors.New("failed to generate salt")
	ErrFilteredRepoEmpty      = errors.New("filtered repo is empty")
	ErrSecretNotFoundForId    = errors.New("secret not found for id")
	ErrRepoSyncNotFound       = errors.New("repo sync not found for the provided id")
)

// Below codes are for gRPC
var (
	ErrStatusNotFound      = status.Error(codes.NotFound, "repo sync not found for the provided id")
	ErrStatusEmptyFromRepo = status.Error(codes.InvalidArgument, "empty from repo")
	ErrStatusDBFailure     = status.Error(codes.Internal, "DB failure")
)
