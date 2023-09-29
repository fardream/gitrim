package gitrim

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
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
func FilterDFSPath(ctx context.Context, dfspath []*object.Commit, s storer.Storer, filter Filter) ([]*object.Commit, error) {
	newpath := make([]*object.Commit, 0, len(dfspath))

	fromorigtonew := make(map[plumbing.Hash]*object.Commit)

	n := len(dfspath)

	for i, c := range dfspath {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if c == nil {
			continue
		}

		parents := make([]*object.Commit, 0, c.NumParents())
		seen := make(map[plumbing.Hash]empty)
	addparentloop:
		for j := 0; j < c.NumParents(); j++ {
			if newparent, found := fromorigtonew[c.ParentHashes[j]]; !found {
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
			return nil, errorf(err, "failed to generate commit at %d for commit %s: %w ", i, c.Hash, err)
		}

		fromorigtonew[c.Hash] = newcommit

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
			newpath = append(newpath, newcommit)
		}
	}

	return newpath, nil
}
