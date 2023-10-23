package gitrim

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type empty = struct{}

// HashSet is simply map from [plumbing.Hash] to [empty]
type HashSet = map[plumbing.Hash]empty

// NewHashSet creates a new set of Hash
func NewHashSet(hashes ...plumbing.Hash) HashSet {
	result := make(map[plumbing.Hash]empty)

	for _, v := range hashes {
		result[v] = empty{}
	}

	return result
}

// NewHashSetFromStrings decodes the input strings and creates a new [HashSet]
func NewHashSetFromStrings(strs ...string) (HashSet, error) {
	hashes, err := DecodeHashHexes(strs...)
	if err != nil {
		return nil, err
	}

	return NewHashSet(hashes...), nil
}

// MustNewHashSetFromStrings decodes and input strings and creates a new [HashSet], or panics
// if any error is encountered.
func MustNewHashSetFromStrings(strs ...string) HashSet {
	set, err := NewHashSetFromStrings(strs...)
	if err != nil {
		panic(err)
	}

	return set
}

// NewHashSetFromCommits collects the hashes of the commits into a [HashSet]
func NewHashSetFromCommits(commits []*object.Commit) HashSet {
	result := make(HashSet)
	for _, c := range commits {
		if c == nil {
			continue
		}

		result[c.Hash] = empty{}
	}
	return result
}
