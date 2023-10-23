package gitrim

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// FilteredDFS contains the mapping between a DFS path from the original repo and the filtered repo.
type FilteredDFS struct {
	FromDFS     []*object.Commit
	fromStorage storer.Storer
	ToDFS       []*object.Commit
	toStorage   storer.Storer

	filter Filter

	FromToTo map[plumbing.Hash]*object.Commit
	ToToFrom map[plumbing.Hash]*object.Commit
}

// NewFilteredDFS filters a slice of [object.Commit] that comes from a Depth First Search from a commit - this means the earlier commits
// should come first in the input slice dfspath, and the head/latest commit should come the last. dfspath can be obtained by [GetDFSPath].
// The result is saved into a [storer.Store].
//
//   - The commits without parents  will become the new roots of the filtered repo.
//   - Filtered commits containing empty trees will be dropped, and subsequent commits following that path will have next non-nil
//     commit as the new root.
//   - Filtered commits containing the exact same tree as its parent will also be dropped,
//     and commit after it will consider its parent its own parent.
//
// The newly created commits will have exact same author info, committor info, commit message,
// but will parent correctly linked and gpg sign information dropped.
//
// The above procedure means that an unfiltered commit will be mapped to a new commit if the commit contains changes from parent after filtering,
// or mapped to its parent if all changes are filtered out. A commit will be mapped to nil if its tree is empty after filter.
//
// If there is no needs to add back the filtered commits, the fromStorage can be left as nil.
//
// Also see [FilterDFSPath].
func NewFilteredDFS(
	ctx context.Context,
	dfspath []*object.Commit,
	fromStorage storer.Storer,
	toStorage storer.Storer,
	filter Filter,
) (*FilteredDFS, error) {
	n := len(dfspath)
	result := &FilteredDFS{
		FromDFS:     make([]*object.Commit, 0, n),
		fromStorage: fromStorage,
		ToDFS:       make([]*object.Commit, 0, n),
		toStorage:   toStorage,

		filter: filter,

		FromToTo: make(map[plumbing.Hash]*object.Commit),
		ToToFrom: make(map[plumbing.Hash]*object.Commit),
	}

	err := result.AppendCommits(ctx, dfspath)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// AppendCommits adds more commits to the filtered dfs path dfs.
// Input morecommits must also conform to the assumption that earlier commits
// come before the later commits. If a commit in the input is already processed and stored
// in dfs, it will be skipped.
func (dfs *FilteredDFS) AppendCommits(
	ctx context.Context,
	morecommits []*object.Commit,
) error {
	s := dfs.toStorage
	filter := dfs.filter

	n := len(morecommits)

	for i, c := range morecommits {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if c == nil {
			continue
		}

		if _, found := dfs.FromToTo[c.Hash]; found {
			continue
		}

		dfs.FromDFS = append(dfs.FromDFS, c)

		parents := make([]*object.Commit, 0, c.NumParents())
		seen := make(map[plumbing.Hash]empty)

	addparentloop:
		for j := 0; j < c.NumParents(); j++ {
			if newparent, found := dfs.FromToTo[c.ParentHashes[j]]; !found {
				continue addparentloop
			} else if newparent != nil {
				if _, found := seen[newparent.Hash]; !found {
					parents = append(parents, newparent)
					seen[newparent.Hash] = empty{}
				}
			}
		}

		newcommit, isparent, err := FilterCommit(ctx, c, parents, s, filter)
		if err != nil {
			return errorf(err, "failed to generate commit at %d for commit %s: %w ", i, c.Hash, err)
		}
		dfs.FromToTo[c.Hash] = newcommit
		commitinfo := "empty"
		if newcommit != nil {
			commitinfo = fmt.Sprintf("%s by %s <%s>", newcommit.Hash, newcommit.Author.Name, newcommit.Author.Email)
		}

		if isparent {
			logger.Info("reuse parent commit", "id", i, "total", n, "hash", c.Hash, "commit", commitinfo)
		} else {
			logger.Info("processing commit", "id", i, "total", n, "hash", c.Hash, "newcommit", commitinfo)
		}

		if newcommit != nil && !isparent {
			dfs.ToDFS = append(dfs.ToDFS, newcommit)
			dfs.ToToFrom[newcommit.Hash] = c
		}
	}

	return nil
}

// SetFromStorage
func (dfs *FilteredDFS) SetFromStorage(s storer.Storer) {
	dfs.fromStorage = s
}

var ErrNilFromStorage = errors.New("from storage is nil, use SetFromStorage to set it")

// ExpandFilteredCommits expands the commits from the filtered repo back to the unfiltered repo.
//
// The input unfiltered commits must have earlier commits before the later commits.
// The parents of those filtered commits must be either in the input commits or already added to the [FilteredDFS] dfs.
// The commits canno be root commits (because filtered repo are always from an unfiltered repo).
func (dfs *FilteredDFS) ExpandFilteredCommits(ctx context.Context, commits []*object.Commit) ([]*object.Commit, error) {
	n := len(commits)
	if dfs.fromStorage == nil {
		return nil, ErrNilFromStorage
	}

	result := make([]*object.Commit, 0, len(commits))

	for i, c := range commits {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if c == nil {
			continue
		}
		if _, found := dfs.ToToFrom[c.Hash]; found {
			continue
		}

		if c.NumParents() == 0 {
			return nil, fmt.Errorf("commit %s has no parents", c.Hash.String())
		}
		parents := make([]*object.Commit, 0, len(c.ParentHashes))
		firstparent, err := c.Parent(0)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain parent for commit %s: %w", c.Hash.String(), err)
		}

		for _, toparenthash := range c.ParentHashes {
			fromparent, found := dfs.ToToFrom[toparenthash]
			if !found {
				return nil, fmt.Errorf("parent commit %s has no correponding commit in unfiltered path", toparenthash.String())
			}
			parents = append(parents, fromparent)
		}

		newcommit, err := ExpandCommitMultiParents(ctx, dfs.toStorage, firstparent, c, parents, dfs.toStorage, dfs.filter)
		if err != nil {
			return nil, fmt.Errorf("failed to expand commit %s: %w", c.Hash.String(), err)
		}
		logger.Info("processing filtered commit", "id", i, "total", n, "hash", c.Hash, "new unfiltered", newcommit.Hash)

		dfs.FromDFS = append(dfs.FromDFS, newcommit)
		dfs.ToToFrom[c.Hash] = newcommit
		dfs.FromToTo[newcommit.Hash] = c
	}

	return result, nil
}
