package gitrim

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GetRoots goes through the input commits and find all the ones that have zero parents
// parents not in the provivded list.
func GetRoots(commits []*object.Commit) []*object.Commit {
	result := make([]*object.Commit, 0, 1)
	all := make(map[plumbing.Hash]empty)
	for _, c := range commits {
		if c == nil || c.Hash.IsZero() {
			continue
		}
		all[c.Hash] = empty{}
	}

	for _, c := range commits {
		if c == nil {
			continue
		}

		n := 0
		for _, p := range c.ParentHashes {
			if _, in := all[p]; in {
				n += 1
			}
		}

		if n == 0 {
			result = append(result, c)
		}
	}

	return result
}
