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

func RemoveGPGForDFSPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ExampleRemoveGPGForDFSPath() {
	// URL for the repo
	url := "https://github.com/go-git/go-git"
	// commit to start from
	headcommithash := plumbing.NewHash("7d047a9f8a43bca9d137d8787278265dd3415219")

	// Clone repo
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})
	RemoveGPGForDFSPanic(err)

	// find the commit
	headcommit, err := r.CommitObject(headcommithash)
	RemoveGPGForDFSPanic(err)

	graph, err := gitrim.GetDFSPath(context.Background(), headcommit, gitrim.MustNewHashSetFromStrings("99e2f85843878671b028d4d01bd4668676226dd1"), 90)

	RemoveGPGForDFSPanic(err)

	// output storer
	outputfs := memory.NewStorage()

	newgraph, err := gitrim.RemoveGPGForDFSPath(context.Background(), graph, outputfs)
	RemoveGPGForDFSPanic(err)

	// Note the result is deterministic
	fmt.Printf("From %d commits, generated %d commits.\n", len(graph), len(newgraph))

	lastcommit := newgraph[240]
	fmt.Println("last commit hash:")
	fmt.Println(lastcommit.Hash)
	fmt.Println("parents:")
	fmt.Println(lastcommit.ParentHashes[0])
	fmt.Println(lastcommit.ParentHashes[1])

	// Output:
	// From 241 commits, generated 241 commits.
	// last commit hash:
	// dc860bcd4bf62d0f90c518022e75621ecbe62885
	// parents:
	// 4a2c8f269c2f122b814f767dab8f579bea6466cd
	// bc0c0692b987229363edb4f591a6eb3318e3ae67
}
