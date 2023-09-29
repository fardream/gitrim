package gitrim

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// FilterCommit creates a new [object.Commit] in the given [storer.Storer]
// by applying filters to the tree in the input [object.Commit].
// Optionally parent commits can set on the generated commit.
// The author info, committor info, commit message will be copied from the input commit.
// Howver, GPG sign information will be dropped.
// The function returns three values, the new commit, a boolean indicating if the returned commit is actually parent containing the same tree, or an error.
//
//   - If after filtering, the tree is empty, a nil will be returned, isparent will be set to false, and error will also be nil.
//   - If the generated tree is exactly the same as the parent's, the parent commit will be returned, isparent bool will be set to true.
//
// Submodules will be silently ignored.
func FilterCommit(
	ctx context.Context,
	c *object.Commit,
	parents []*object.Commit,
	s storer.Storer,
	filters Filter,
) (*object.Commit, bool, error) {
	t, err := c.Tree()
	if err != nil {
		return nil, false, fmt.Errorf("failed to obtain tree for commit %s: %w", c.Hash.String(), err)
	}

	newtree, err := FilterTree(ctx, t, nil, s, filters)
	if err != nil {
		return nil, false, errorf(err, "failed to filter tree: %w", err)
	}

	if newtree == nil {
		return nil, false, nil
	}

	var parenthashes []plumbing.Hash

	for _, parent := range parents {
		if parent == nil {
			continue
		}
		if parent.TreeHash == newtree.Hash {
			return parent, true, nil
		}
		parenthashes = append(parenthashes, parent.Hash)
	}

	newcommit := &object.Commit{
		TreeHash:     newtree.Hash,
		Author:       c.Author,
		Committer:    c.Committer,
		Message:      c.Message,
		ParentHashes: parenthashes,
	}

	newhash, err := GetHash(newcommit)
	if err != nil {
		return nil, false, fmt.Errorf("failed to obtain new hash for commit: %w ", err)
	}

	newcommit.Hash = *newhash

	if err := updateHashAndSave(ctx, newcommit, s); err != nil {
		return nil, false, errorf(err, "failed to save commit: %w", err)
	}

	return newcommit, false, nil
}
