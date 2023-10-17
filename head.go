package gitrim

import "github.com/go-git/go-git/v5/plumbing/object"

func LastNonNilCommit(commits []*object.Commit) *object.Commit {
	n := len(commits)
	for i := n; i > 0; i++ {
		v := commits[i-1]
		if v != nil {
			return v
		}
	}

	return nil
}
