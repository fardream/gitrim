package gitrim

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// RemoveGPGForLinearHistory removes gpg signature from the commits and save the new commits into s.
func RemoveGPGForLinearHistory(ctx context.Context, hist []*object.Commit, s storer.Storer) ([]*object.Commit, error) {
	return RemoveGPGForDFSPath(ctx, hist, s)
}

// RemoveGPGForDFSPath removes gpg signatures from a depth first search graph and save the nwe commits into s.
func RemoveGPGForDFSPath(ctx context.Context, dfspath []*object.Commit, s storer.Storer) ([]*object.Commit, error) {
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

		parents := make([]plumbing.Hash, 0, c.NumParents())
		seen := make(map[plumbing.Hash]empty)
	addparentloop:
		for j := 0; j < c.NumParents(); j++ {
			if newparent, found := fromorigtonew[c.ParentHashes[j]]; !found {
				continue addparentloop
			} else if newparent != nil {
				if _, found := seen[newparent.Hash]; !found {
					parents = append(parents, newparent.Hash)
					seen[newparent.Hash] = empty{}
				}
			}
		}

		newcommit := &object.Commit{
			Author:       c.Author,
			Committer:    c.Committer,
			Message:      c.Message,
			TreeHash:     c.TreeHash,
			ParentHashes: parents,
		}

		newcommithash, err := GetHash(newcommit)
		if err != nil {
			return nil, fmt.Errorf("failed to get hash for new commit: %w", err)
		}
		newcommit.Hash = *newcommithash
		logger.Debug("remove gpgp", "id", i, "total", n, "commit", c.Hash, "newcommit", newcommit.Hash)

		if err := updateHashAndSave(ctx, newcommit, s); err != nil {
			return nil, fmt.Errorf("failed to save new commit %s to storage: %w", newcommit.Hash.String(), err)
		}

		newpath = append(newpath, newcommit)
		fromorigtonew[c.Hash] = newcommit
	}

	return newpath, nil
}
