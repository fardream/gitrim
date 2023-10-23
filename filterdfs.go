package gitrim

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// FilterDFSPath filters a slice of [object.Commit] that comes from a Depth First Search from a commit - this means the earlier commits
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
func FilterDFSPath(
	ctx context.Context,
	dfspath []*object.Commit,
	s storer.Storer,
	filter Filter,
) ([]*object.Commit, error) {
	r, err := NewFilteredDFS(ctx, dfspath, nil, s, filter)
	if err != nil {
		return nil, err
	}

	return r.todfs, nil
}
