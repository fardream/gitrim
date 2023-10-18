package svc

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/fardream/gitrim"
)

type syncInfo struct {
	RootCommits    []string
	InitHeadCommit string

	LastSyncFromCommit string
	LastSyncToCommit   string
}

// syncWksp syncs the from workspace to the to workspace.
//
// if the from workspace is empty, the returned [syncInfo] will be nil.
func syncWksp(
	ctx context.Context,
	fromwksp *workspace,
	frombranch string,
	towksp *workspace,
	tobranch string,
	filters []string,
	reqrootcommits []string,
	maxdepth int64,
	forcepush bool,
) (*syncInfo, error) {
	result := &syncInfo{}

	if fromwksp.isempty {
		return nil, nil
	}

	frombranchref := plumbing.NewBranchReferenceName(fromwksp.branch)
	branch, err := fromwksp.storage.Reference(frombranchref)
	if err != nil {
		if iter, err := fromwksp.storage.IterReferences(); err == nil {
			iter.ForEach(func(r *plumbing.Reference) error {
				fmt.Printf("we have reference: %s\n", r)
				return nil
			})
		}
		return nil, fmt.Errorf("failed to find from branch %s in from repo: %w", frombranchref, err)
	}

	headcommit, err := object.GetCommit(fromwksp.storage, branch.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get head commit for from branch %s: %w", branch.Hash(), err)
	}
	result.InitHeadCommit = headcommit.String()

	rootcommits, err := gitrim.DecodeHashHexes(reqrootcommits...)
	if err != nil {
		return nil, fmt.Errorf("failed to decode root commits: %w", err)
	}

	commits, err := gitrim.GetDFSPath(ctx, headcommit, rootcommits, int(maxdepth))
	if err != nil {
		return nil, fmt.Errorf("failed to obtain commit graph from from repo: %w", err)
	}

	roots := gitrim.GetRoots(commits)

	for _, r := range roots {
		result.RootCommits = append(result.RootCommits, r.Hash.String())
	}

	f, err := gitrim.NewOrFilterForPatterns(filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}
	newcommits, err := gitrim.FilterDFSPath(ctx, commits, towksp.storage, f)
	if err != nil {
		return nil, fmt.Errorf("failed to filter commits according to filter: %w", err)
	}

	newhead := gitrim.LastNonNilCommit(newcommits)
	if newhead == nil {
		return nil, ErrFilteredRepoEmpty
	}

	h := newhead.Hash
	refname := plumbing.NewBranchReferenceName(towksp.branch)
	ref := plumbing.NewHashReference(refname, h)
	if err := towksp.storage.SetReference(ref); err != nil {
		return nil, fmt.Errorf("failed to set to branch in to repo: %w", err)
	}

	result.LastSyncFromCommit = commits[len(commits)-1].Hash.String()
	result.LastSyncToCommit = newhead.Hash.String()

	err = towksp.updateToLatest(ctx, forcepush)
	if err != nil {
		return nil, fmt.Errorf("failed to push: %w", err)
	}

	return result, nil
}
