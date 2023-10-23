package gitrim

import (
	"context"
	"fmt"
	"math"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type dfsBuilderNode struct {
	data       *object.Commit
	nparent    int
	nextvisit  int
	generation int
}

type dfsBuilder struct {
	seen  map[plumbing.Hash]empty
	stack []*dfsBuilderNode
}

func newDFSBuilder() *dfsBuilder {
	return &dfsBuilder{
		stack: make([]*dfsBuilderNode, 0),
		seen:  make(map[plumbing.Hash]empty),
	}
}

func (gb *dfsBuilder) add(v *object.Commit, generation int) {
	hash := v.Hash
	if _, seen := gb.seen[hash]; seen {
		return
	}

	gb.seen[hash] = empty{}
	gb.stack = append(gb.stack, &dfsBuilderNode{
		data:       v,
		nparent:    v.NumParents(),
		nextvisit:  0,
		generation: generation,
	})
}

func (gb *dfsBuilder) pop() error {
	if len(gb.stack) == 0 {
		return fmt.Errorf("failed to pop empty stack")
	}

	gb.stack = gb.stack[:len(gb.stack)-1]

	return nil
}

func (gb *dfsBuilder) top() *dfsBuilderNode {
	if len(gb.stack) == 0 {
		return nil
	}

	return gb.stack[len(gb.stack)-1]
}

// GetDFSPath gets a deterministic depth first search path from a head commit,
// the returned slice has the head commit as the last one in the slice,
// and one of the root commits as the first of the slice.
// The search always search the first parent, then second, and so-on, therefore the commits first returned
// are history from git command with "--first-parent" parameter.
//
// rootcommits can be optionally set so the search will stop for that path if one of those commits is seen.
// Max generation can be turned off by setting it to any value that is 0 or negative.
func GetDFSPath(
	ctx context.Context,
	head *object.Commit,
	roots HashSet,
	maxGeneration int,
) ([]*object.Commit, error) {
	result := make([]*object.Commit, 0)
	gb := newDFSBuilder()

	gb.add(head, 0)

	if roots == nil {
		roots = make(map[plumbing.Hash]empty)
	}

	if maxGeneration <= 0 {
		maxGeneration = math.MaxInt
	}

addloop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		current := gb.top()

		if current == nil {
			break addloop
		}

		_, isroot := roots[current.data.Hash]
		switch {
		case current.nextvisit == current.nparent:
			result = append(result, current.data)
			if err := gb.pop(); err != nil {
				return nil, err
			}
		case isroot:
			result = append(result, current.data)
			if err := gb.pop(); err != nil {
				return nil, err
			}
		case current.generation >= maxGeneration-1:
			result = append(result, current.data)
			if err := gb.pop(); err != nil {
				return nil, err
			}
		default:
			p, err := current.data.Parent(current.nextvisit)
			if err != nil {
				return nil, fmt.Errorf(
					"cannot get parent %d for %s: %w",
					current.nextvisit,
					current.data.Hash.String(),
					err)
			}
			current.nextvisit += 1
			gb.add(p, current.generation+1)
		}
	}

	return result, nil
}
