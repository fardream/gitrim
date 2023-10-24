package gitrim

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// LazyCommit records the Hash, and optionally the [object.Commit].
// Use GetCommit to lazy load the commit.
type LazyCommit struct {
	c    *object.Commit
	Hash plumbing.Hash
	s    storer.Storer
}

func NewLazyCommitWithHash(h plumbing.Hash, s storer.Storer) *LazyCommit {
	return &LazyCommit{
		Hash: h,
		s:    s,
	}
}

func NewLazyCommit(c *object.Commit) *LazyCommit {
	return &LazyCommit{
		c:    c,
		Hash: c.Hash,
	}
}

func (l *LazyCommit) GetCommit() (*object.Commit, error) {
	if l.c != nil {
		return l.c, nil
	}

	c, err := object.GetCommit(l.s, l.Hash)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type KeyedDFSPath struct {
	Path         []*LazyCommit
	HashToCommit map[plumbing.Hash]*LazyCommit
	s            storer.Storer
}

func NewKeyedDFSPath(s storer.Storer) *KeyedDFSPath {
	return &KeyedDFSPath{
		s:            s,
		HashToCommit: make(map[plumbing.Hash]*LazyCommit),
	}
}

func (k *KeyedDFSPath) Add(lc *LazyCommit) {
	if lc == nil {
		return
	}
	k.Path = append(k.Path, lc)
	k.HashToCommit[lc.Hash] = lc
}

func (k *KeyedDFSPath) AddCommit(c *object.Commit) {
	if c == nil {
		return
	}

	if v, found := k.HashToCommit[c.Hash]; found {
		v.c = c
	} else {
		lz := &LazyCommit{c: c, Hash: c.Hash, s: k.s}
		k.Path = append(k.Path, lz)
		k.HashToCommit[lz.Hash] = lz
	}
}

func (k *KeyedDFSPath) AddHash(h plumbing.Hash) {
	if h.IsZero() {
		return
	}

	if _, found := k.HashToCommit[h]; found {
		return
	}
	lz := &LazyCommit{
		Hash: h,
		s:    k.s,
	}
	k.HashToCommit[h] = lz
	k.Path = append(k.Path, lz)
}

func (k *KeyedDFSPath) GetCommit(h plumbing.Hash) (*object.Commit, error) {
	v, found := k.HashToCommit[h]

	if !found {
		return nil, fmt.Errorf("%s is not found in keyed dfs", h.String())
	}

	return v.GetCommit()
}

func (k *KeyedDFSPath) GetPath() ([]*object.Commit, error) {
	result := make([]*object.Commit, 0, len(k.Path))
	for _, p := range k.Path {
		c, err := p.GetCommit()
		if err != nil {
			return nil, fmt.Errorf("failed to get commit %s: %w", p.Hash, err)
		}
		result = append(result, c)
	}

	return result, nil
}

// FilteredDFS contains the mapping between a DFS path from the original repo and the filtered repo.
type FilteredDFS struct {
	FromDFS     KeyedDFSPath
	fromStorage storer.Storer
	ToDFS       KeyedDFSPath
	toStorage   storer.Storer

	filter Filter

	FromToTo map[plumbing.Hash]plumbing.Hash
	ToToFrom map[plumbing.Hash]plumbing.Hash
}

// NewEmptyFilteredDFS creates an empty [FilteredDFS] with the given storage and filter.
func NewEmptyFilteredDFS(
	fromstorage storer.Storer,
	tostorage storer.Storer,
	filter Filter,
) *FilteredDFS {
	result := &FilteredDFS{
		fromStorage: fromstorage,
		FromDFS:     *NewKeyedDFSPath(fromstorage),
		toStorage:   tostorage,
		ToDFS:       *NewKeyedDFSPath(tostorage),
		filter:      filter,
		FromToTo:    make(map[plumbing.Hash]plumbing.Hash),
		ToToFrom:    make(map[plumbing.Hash]plumbing.Hash),
	}

	return result
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
	result := NewEmptyFilteredDFS(fromStorage, toStorage, filter)

	err := result.AppendCommits(ctx, dfspath)
	if err != nil {
		return nil, err
	}

	return result, nil
}

var (
	ErrNilToStorage = errors.New("nil to storage")
	ErrEmptyFilter  = errors.New("empty filter")
)

// AppendCommits adds more commits to the filtered dfs path dfs.
// Input morecommits must also conform to the assumption that earlier commits
// come before the later commits. If a commit in the input is already processed and stored
// in dfs, it will be skipped.
func (dfs *FilteredDFS) AppendCommits(
	ctx context.Context,
	morecommits []*object.Commit,
) error {
	if dfs.toStorage == nil {
		return ErrNilToStorage
	}
	if dfs.filter == nil {
		return ErrEmptyFilter
	}
	if dfs.FromToTo == nil {
		dfs.FromToTo = make(map[plumbing.Hash]plumbing.Hash)
	}
	if dfs.ToToFrom == nil {
		dfs.ToToFrom = make(map[plumbing.Hash]plumbing.Hash)
	}

	s := dfs.toStorage
	filter := dfs.filter

	n := len(morecommits)

	for i, vc := range morecommits {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// technically, this check should not be here. just in case some one accidentally passed in
		if vc == nil {
			logger.Warn("nil commit", "id", i, "total", n)
			continue
		}

		// skip the commit if it already processed
		if _, found := dfs.FromToTo[vc.Hash]; found {
			continue
		}

		lc := NewLazyCommit(vc)
		c, _ := lc.GetCommit()
		// this commit is not processed, add it to the path.
		dfs.FromDFS.Add(lc)

		// now creates parent list.
		// the order of parents must be maintained, and we only keep the first occurence if same commit appears multiple times.
		parents := make([]*object.Commit, 0, c.NumParents())
		parentsSeen := make(map[plumbing.Hash]empty)

	addparentloop:
		for j := 0; j < c.NumParents(); j++ {
			newparent, found := dfs.FromToTo[c.ParentHashes[j]]
			if !found {
				logger.Warn("parent to commit not found", "parent from commit", c.ParentHashes[j].String())
				continue addparentloop
			}
			if newparent.IsZero() {
				continue addparentloop
			}
			if _, found := parentsSeen[newparent]; found {
				continue addparentloop
			}
			np, err := dfs.ToDFS.GetCommit(newparent)
			if err != nil {
				return fmt.Errorf("failed to obtain parent commit %s due to: %w", newparent, err)
			}
			parents = append(parents, np)
			parentsSeen[newparent] = empty{}
		}

		newcommit, isparent, err := FilterCommit(ctx, c, parents, s, filter)
		if err != nil {
			return errorf(err, "failed to generate commit at %d for commit %s: %w ", i, c.Hash, err)
		}
		commitinfo := "empty"
		if newcommit != nil {
			commitinfo = fmt.Sprintf("%s by %s <%s>", newcommit.Hash, newcommit.Author.Name, newcommit.Author.Email)
		}
		if isparent {
			logger.Debug("reuse parent commit", "id", i, "total", n, "hash", c.Hash, "commit", commitinfo)
		} else {
			logger.Debug("processing commit", "id", i, "total", n, "hash", c.Hash, "newcommit", commitinfo)
		}
		if newcommit == nil {
			dfs.FromToTo[c.Hash] = plumbing.ZeroHash
		} else {
			dfs.FromToTo[c.Hash] = newcommit.Hash
		}
		if newcommit != nil && !isparent {
			dfs.ToDFS.AddCommit(newcommit)
			dfs.ToToFrom[newcommit.Hash] = c.Hash
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
			fromparenthash, found := dfs.ToToFrom[toparenthash]
			if !found {
				return nil, fmt.Errorf("parent commit %s has no correponding commit in unfiltered path", toparenthash.String())
			}
			fromparent, err := dfs.FromDFS.GetCommit(fromparenthash)
			if err != nil {
				return nil, fmt.Errorf("cannot get parent commit %s: %w", fromparenthash, err)
			}
			parents = append(parents, fromparent)
		}

		newcommit, err := ExpandCommitMultiParents(ctx, dfs.toStorage, firstparent, c, parents, dfs.toStorage, dfs.filter)
		if err != nil {
			return nil, fmt.Errorf("failed to expand commit %s: %w", c.Hash.String(), err)
		}
		logger.Info("processing filtered commit", "id", i, "total", n, "hash", c.Hash, "new unfiltered", newcommit.Hash)

		dfs.FromDFS.AddCommit(newcommit)
		dfs.ToToFrom[c.Hash] = newcommit.Hash
		dfs.FromToTo[newcommit.Hash] = c.Hash
	}

	return result, nil
}
