// errors

package svc

import "errors"

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
)
