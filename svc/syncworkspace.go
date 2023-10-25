package svc

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fardream/gitrim"
)

type syncWorkspace struct {
	db *DbRepoSync

	filter gitrim.Filter
	roots  gitrim.HashSet

	fromWksp       *workspace
	fromNewcommits []*object.Commit
	fromStatus     LastSyncCommitStatus_Enum

	toWksp       *workspace
	toNewcommits []*object.Commit
	toStatus     LastSyncCommitStatus_Enum
}

func loadSyncWorkspaceFromDb(ctx context.Context, remoeConfig map[string]*RemoteConfig, idhex string, db *bbolt.DB, requireexist bool) (*syncWorkspace, error) {
	reposync, _, err := getRepoSync(db, idhex, requireexist)
	if err != nil {
		return nil, err
	}

	return newSyncWorkspace(ctx, remoeConfig, reposync)
}

func newSyncWorkspace(ctx context.Context, remoteConfig map[string]*RemoteConfig, reposync *DbRepoSync) (*syncWorkspace, error) {
	filter, err := gitrim.NewOrFilterForPatterns(reposync.SyncData.Filter.CanonicalFilters...)
	if err != nil {
		return nil, err
	}

	roots, err := gitrim.NewHashSetFromStrings(reposync.SyncData.RootCommits...)
	if err != nil {
		return nil, err
	}

	fromwksp, err := newWorkspace(ctx, remoteConfig, reposync.SyncData.FromRepo, reposync.SyncData.FromBranch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain from repo: %s", err.Error())
	}
	towksp, err := newWorkspace(ctx, remoteConfig, reposync.SyncData.ToRepo, reposync.SyncData.ToBranch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to obtain to repo: %s", err.Error())
	}
	if reposync.Stat == nil {
		reposync.Stat = EmptySyncStat()
	}

	fromhead, frompast, tohead, topast, err := reposync.Stat.Hashes()
	if err != nil {
		return nil, err
	}
	fromstatus, fromcommits, err := getLastSyncCommitStatus(ctx, fromhead, gitrim.CombineHashSets(roots, frompast), false, fromwksp)
	if err != nil {
		return nil, err
	}
	tostatus, tocommits, err := getLastSyncCommitStatus(ctx, tohead, topast, true, towksp)
	if err != nil {
		return nil, err
	}

	return &syncWorkspace{
		db: reposync,

		filter: filter,
		roots:  roots,

		fromWksp:       fromwksp,
		fromNewcommits: fromcommits,
		fromStatus:     fromstatus,

		toWksp:       towksp,
		toNewcommits: tocommits,
		toStatus:     tostatus,
	}, nil
}

var ErrZeroRoots = errors.New("zero roots found for DFS path")

func getLastSyncCommitStatus(
	ctx context.Context,
	currentcommit plumbing.Hash,
	pastcommits gitrim.HashSet,
	islinear bool,
	wksp *workspace,
) (LastSyncCommitStatus_Enum, []*object.Commit, error) {
	if wksp.branchhead == nil {
		if currentcommit.IsZero() {
			return LastSyncCommitStatus_INSYNC, nil, nil
		} else {
			return LastSyncCommitStatus_DIVERGED, nil, nil
		}
	}

	// if currentcommit is empty, then the branch has advanced (it has commits now)
	if currentcommit.IsZero() {
		return LastSyncCommitStatus_ADVANCED, nil, nil
	}

	if wksp.branchhead.Hash == currentcommit {
		return LastSyncCommitStatus_INSYNC, nil, nil
	}

	historicalcommits, err := gitrim.GetDFSPath(ctx, wksp.branchhead, pastcommits, 0)
	if err != nil {
		return LastSyncCommitStatus_UNKNOWN, nil, fmt.Errorf("failed to obtain commits to inspect history: %w", err)
	}

	roots := gitrim.GetRoots(historicalcommits)

	if len(roots) == 0 {
		return LastSyncCommitStatus_UNKNOWN, historicalcommits, ErrZeroRoots
	}
	if len(roots) != 1 || roots[0].Hash != currentcommit {
		return LastSyncCommitStatus_DIVERGED, historicalcommits, nil
	}

	return LastSyncCommitStatus_ADVANCED, historicalcommits, nil
}

func loadSyncWorkspaceFroReq(
	ctx context.Context,
	remoteConfig map[string]*RemoteConfig,
	db *bbolt.DB,
	req RequestWithPossibleOverride,
	mustExist bool,
) (*syncWorkspace, error) {
	reposync, _, err := getRepoSync(db, req.GetId(), true)
	if err != nil {
		return nil, err
	}

	if req.GetOverrideToBranch() != "" {
		reposync.SyncData.ToBranch = req.GetOverrideToBranch()
	}
	if req.GetOverrideFromBranch() != "" {
		reposync.SyncData.FromBranch = req.GetOverrideFromBranch()
	}

	return newSyncWorkspace(ctx, remoteConfig, reposync)
}

var ErrToNotInSync = errors.New("to branch not in sync")

func getHashStringPossibleNil(c *object.Commit) string {
	if c == nil {
		return "<empty>"
	}
	return c.Hash.String()
}

func (sw *syncWorkspace) getFilteredDFS() (*gitrim.FilteredDFS, error) {
	stat := sw.db.Stat
	return gitrim.NewFilteredDFSWithStat(stat.FromDfs, stat.ToDfs, stat.FromToTo, stat.ToToFrom, sw.fromWksp.storage, sw.toWksp.storage, sw.filter)
}

func (sw *syncWorkspace) syncToTo(ctx context.Context, force bool) ([]*object.Commit, error) {
	if sw.toStatus != LastSyncCommitStatus_INSYNC && !force {
		logger.Warn("to branch not in sync", "status", sw.toStatus.String(), "expecting", sw.db.Stat.LastSyncToCommit, "got", getHashStringPossibleNil(sw.toWksp.branchhead))
		return nil, ErrToNotInSync
	}
	// already up to date
	if sw.fromStatus == LastSyncCommitStatus_INSYNC && sw.toStatus == LastSyncCommitStatus_INSYNC {
		if !force {
			logger.Info("already in sync, skip", "from", sw.fromWksp.branch, "to", sw.toWksp.branch)
			return nil, nil
		}
		logger.Info("already in sync, force update", "from", sw.fromWksp.branch, "to", sw.toWksp.branch)
	}

	stat := sw.db.Stat

	if sw.toStatus != LastSyncCommitStatus_INSYNC || sw.fromStatus != LastSyncCommitStatus_ADVANCED {
		logger.Info("reset stat", "from-status", sw.fromStatus, "to-status", sw.toStatus)
		stat.SetToEmpty()
	}

	filtereddfs, err := sw.getFilteredDFS()
	if err != nil {
		return nil, fmt.Errorf("failed to get local status for commits: %w", err)
	}

	fromhead, fromcommits, _, _, err := stat.Hashes()
	if err != nil {
		return nil, fmt.Errorf("failed to get from head and from commits from stat: %w", err)
	}
	if len(sw.fromNewcommits) == 0 {
		if sw.toStatus == LastSyncCommitStatus_INSYNC {
			logger.Info("from-in-sync-add-fromcommits")
			sw.fromNewcommits, err = sw.fromWksp.getNewCommits(ctx, fromhead, gitrim.CombineHashSets(fromcommits, sw.roots), false)
		} else {
			sw.fromNewcommits, err = sw.fromWksp.getNewCommits(ctx, fromhead, sw.roots, false)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to obtain new commits for from repo: %w", err)
		}
	}

	if len(sw.fromNewcommits) == 0 {
		logger.Info("no new commits")
		return nil, nil
	}

	newcommits, err := filtereddfs.AppendCommits(ctx, sw.fromNewcommits)
	if err != nil {
		return nil, fmt.Errorf("failed to add new commits: %w", err)
	}

	if len(newcommits) == 0 {
		logger.Info("zero new commits")
		return nil, nil
	}

	fromc, toc, err := filtereddfs.LastCommits()
	if err != nil {
		return nil, fmt.Errorf("failed to get the last commits after filtering: %w", err)
	}

	sw.toWksp.isempty = false

	err = sw.toWksp.updateBranchHead(toc)
	if err != nil {
		return nil, fmt.Errorf("failed to update branch head after filtering: %w", err)
	}
	err = sw.toWksp.pushToRemote(ctx, force)
	if err != nil {
		return nil, fmt.Errorf("failed to push: %w", err)
	}

	stat.FromDfs, stat.ToDfs, stat.FromToTo, stat.ToToFrom = filtereddfs.DumpStat()
	stat.LastSyncFromCommit = fromc.Hash.String()
	stat.LastSyncToCommit = toc.Hash.String()

	return newcommits, nil
}

func checkSubHistoryForSyncToFrom(fromstatus, tostatus LastSyncCommitStatus_Enum) SubRepoCommitsCheck_Status {
	switch {
	case fromstatus != LastSyncCommitStatus_INSYNC: // if from is not in sync, return
		return SubRepoCommitsCheck_FROM_NOT_IN_SYNC
	case tostatus == LastSyncCommitStatus_INSYNC:
		return SubRepoCommitsCheck_TO_NO_NEW_COMMITS
	case tostatus != LastSyncCommitStatus_ADVANCED:
		return SubRepoCommitsCheck_TO_DIVERGED
	default:
		return SubRepoCommitsCheck_CHECK_PASSED
	}
}

var (
	ErrToHasNoNewCommits               = errors.New("to has no new commits")
	ErrSubCommitCannotHavePGPSignature = errors.New("sub commit cannot have PGP signature")
)

func (sw *syncWorkspace) checkCommits(ctx context.Context) ([]*gitrim.FilePatchCheckResult, bool, error) {
	status := checkSubHistoryForSyncToFrom(sw.fromStatus, sw.toStatus)
	if status != SubRepoCommitsCheck_CHECK_PASSED {
		return nil, false, fmt.Errorf("repos are not in good status to sync: from repo status %s, to repo status %s", sw.fromStatus.String(), sw.toStatus.String())
	}

	if len(sw.toNewcommits) == 0 {
		_, _, toheadhash, topastcommits, err := sw.db.Stat.Hashes()
		tonew, err := sw.toWksp.getNewCommits(ctx, toheadhash, topastcommits, true)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get new commits for to repo: %w", err)
		}
		sw.toNewcommits = tonew
	}

	if len(sw.toNewcommits) == 0 {
		return nil, false, ErrToHasNoNewCommits
	}

	var hasgpg bool

	for _, nc := range sw.toNewcommits {
		if nc.PGPSignature != "" {
			hasgpg = true
		}
	}

	filtereddfs, err := sw.getFilteredDFS()
	if err != nil {
		return nil, false, err
	}

	checkresults, err := filtereddfs.CheckCommitsAgainstFilter(ctx, sw.toNewcommits)
	if err != nil {
		return nil, false, err
	}

	return checkresults, hasgpg, nil
}

func (sw *syncWorkspace) syncToFrom(ctx context.Context, dopush bool, allowpgp bool) ([]*object.Commit, error) {
	status := checkSubHistoryForSyncToFrom(sw.fromStatus, sw.toStatus)
	if status != SubRepoCommitsCheck_CHECK_PASSED {
		return nil, fmt.Errorf("repos are not in good status to sync: from repo status %s, to repo status %s", sw.fromStatus.String(), sw.toStatus.String())
	}

	if len(sw.toNewcommits) == 0 {
		_, _, toheadhash, topastcommits, err := sw.db.Stat.Hashes()
		tonew, err := sw.toWksp.getNewCommits(ctx, toheadhash, topastcommits, true)
		if err != nil {
			return nil, fmt.Errorf("failed to get new commits for to repo: %w", err)
		}
		sw.toNewcommits = tonew
	}

	if len(sw.toNewcommits) == 0 {
		return nil, ErrToHasNoNewCommits
	}

	if !allowpgp {
		for _, nc := range sw.toNewcommits {
			if nc.PGPSignature != "" {
				return nil, ErrSubCommitCannotHavePGPSignature
			}
		}
	}
	stat := sw.db.Stat

	filtereddfs, err := sw.getFilteredDFS()
	if err != nil {
		return nil, fmt.Errorf("failed to get local status for commits: %w", err)
	}

	newcommits, err := filtereddfs.ExpandFilteredCommits(ctx, sw.toNewcommits)
	if err != nil {
		return nil, fmt.Errorf("failed to expand commits: %w", err)
	}

	fromc, toc, err := filtereddfs.LastCommits()
	if err != nil {
		return nil, fmt.Errorf("failed to get the last commits after filtering: %w", err)
	}

	err = sw.fromWksp.updateBranchHead(fromc)
	if err != nil {
		return nil, fmt.Errorf("failed to update branch head after expanding: %w", err)
	}
	if dopush {
		err = sw.fromWksp.pushToRemote(ctx, false)
		if err != nil {
			return nil, fmt.Errorf("failed to update from repo: %w", err)
		}

	}

	stat.FromDfs, stat.ToDfs, stat.FromToTo, stat.ToToFrom = filtereddfs.DumpStat()
	stat.LastSyncFromCommit = fromc.Hash.String()
	stat.LastSyncToCommit = toc.Hash.String()

	return newcommits, nil
}
