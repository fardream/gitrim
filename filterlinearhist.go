package gitrim

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// FilterLinearHistory performs filters on a sequence of commits of a linear history and
// produces new commits in the provided [storer.Store].
// Similar to [FilterDFSPath]:
//
//   - The first commit will become the new root of the filtered repo.
//   - Filtered commits containing empty trees cause all previous commits to be dropped.
//     The next commit with non-empty tree will become the new root.
//   - Filtered commits containing the exact same tree as its parent will also be dropped,
//     and commit after it will consider its parent its own parent.
//
// The input commits can be obtained from [GetLinearHistory].
func FilterLinearHistory(
	ctx context.Context,
	hist []*object.Commit,
	s storer.Storer,
	filter Filter,
) ([]*object.Commit, error) {
	// this is implemented before FilterDFSPath is done.
	return FilterDFSPath(ctx, hist, s, filter)
}
