package gitrim

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// ExpandCommit added the changes contained in the filteredNew to filteredOrig and try to apply them to target, it will generate a new commit.
func ExpandCommit(
	ctx context.Context,
	sourceStorer storer.Storer,
	filteredOrig *object.Commit,
	filteredNew *object.Commit,
	target *object.Commit,
	targetStorer storer.Storer,
	filter Filter,
) (*object.Commit, error) {
	newtarget := &object.Commit{
		Committer:    filteredNew.Committer,
		Author:       filteredNew.Author,
		Message:      filteredNew.Message,
		ParentHashes: []plumbing.Hash{target.Hash},
	}

	err := expandCommitInner(ctx, newtarget, sourceStorer, filteredOrig, filteredNew, target, targetStorer, filter)
	if err != nil {
		return nil, err
	}

	return newtarget, nil
}

func expandCommitInner(
	ctx context.Context,
	newtarget *object.Commit,
	sourceStorer storer.Storer,
	filteredOrig *object.Commit,
	filteredNew *object.Commit,
	target *object.Commit,
	targetStorer storer.Storer,
	filter Filter,
) error {
	filteredOrigTree, err := filteredOrig.Tree()
	if err != nil {
		return fmt.Errorf("failed to obtain filtered parent tree: %w", err)
	}
	filteredNewTree, err := filteredNew.Tree()
	if err != nil {
		return fmt.Errorf("failed to obtain filtered new tree: %w", err)
	}
	targetOrigTree, err := target.Tree()
	if err != nil {
		return fmt.Errorf("failed to obtain target parent tree: %w", err)
	}

	newtree, err := ExpandTree(ctx, sourceStorer, filteredOrigTree, filteredNewTree, targetOrigTree, targetStorer, filter)
	if err != nil {
		return errorf(err, "failed to expand tree for target: %w", err)
	}
	if newtree != nil {
		newtarget.TreeHash = newtree.Hash
	} else {
		logger.Warn("empty tree", "filtered-new-commit", filteredNew.Hash, "filtered-orig-commit", filteredOrig.Hash, "target", target.Hash)
	}

	err = updateHashAndSave(ctx, newtarget, targetStorer)
	if err != nil {
		return errorf(err, "failed to update new tree into storage: %w", err)
	}

	return nil
}

var ErrEmptyToParents = errors.New("target commits is empty")

// ExpandCommitMultiParents is similar to [ExpandCommit] but with multiple parents. The first parent is used to identify changes.
func ExpandCommitMultiParents(ctx context.Context,
	sourceStorer storer.Storer,
	filteredOrig *object.Commit,
	filteredNew *object.Commit,
	parents []*object.Commit,
	targetStorer storer.Storer,
	filter Filter,
) (*object.Commit, error) {
	if len(parents) <= 0 {
		return nil, ErrEmptyToParents
	}

	newtarget := &object.Commit{
		Committer: filteredNew.Committer,
		Author:    filteredNew.Author,
		Message:   filteredNew.Message,
	}

	for _, p := range parents {
		newtarget.ParentHashes = append(newtarget.ParentHashes, p.Hash)
	}

	err := expandCommitInner(ctx, newtarget, sourceStorer, filteredOrig, filteredNew, parents[0], targetStorer, filter)
	if err != nil {
		return nil, err
	}

	return newtarget, nil
}
