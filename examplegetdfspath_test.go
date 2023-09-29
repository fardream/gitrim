package gitrim_test

import (
	"context"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/fardream/gitrim"
)

func GetDFSPathPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ExampleGetDFSPath() {
	// URL for the repo
	url := "https://github.com/go-git/go-git"
	// commit to start from
	headcommithash := plumbing.NewHash("7d047a9f8a43bca9d137d8787278265dd3415219")

	// Clone repo
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})
	GetDFSPathPanic(err)

	// find the commit
	headcommit, err := r.CommitObject(headcommithash)
	GetDFSPathPanic(err)

	graph, err := gitrim.GetDFSPath(context.Background(), headcommit, nil, 0)

	GetDFSPathPanic(err)

	fmt.Println(len(graph))
	fmt.Println(graph[0].Hash.String())
	fmt.Println(graph[len(graph)-1].Hash.String())

	// Output:
	// 1986
	// 5d7303c49ac984a9fec60523f2d5297682e16646
	// 7d047a9f8a43bca9d137d8787278265dd3415219
}
