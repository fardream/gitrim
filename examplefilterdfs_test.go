package gitrim_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/fardream/gitrim"
)

func FilterDFSPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func removeEmptyLines(s string) string {
	lines := strings.Split(s, "\n")
	r := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			r = append(r, "")
		} else {
			r = append(r, line)
		}
	}

	return strings.Join(r, "\n")
}

// Example cloning a repo into in-memory store, select several commits from a specific commit, and filter it into another in-memory store.
func ExampleFilterDFSPath() {
	// URL for the repo
	url := "https://github.com/go-git/go-git"
	// commit to start from
	headcommithash := plumbing.NewHash("7d047a9f8a43bca9d137d8787278265dd3415219")

	// Clone repo
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})
	FilterDFSPanic(err)

	// find the commit
	headcommit, err := r.CommitObject(headcommithash)
	FilterDFSPanic(err)

	graph, err := gitrim.GetDFSPath(context.Background(), headcommit, []plumbing.Hash{plumbing.NewHash("99e2f85843878671b028d4d01bd4668676226dd1")}, 90)

	FilterDFSPanic(err)

	// select 3 files
	orfilter, err := gitrim.NewOrFilterForPatterns(
		"README.md",
		"LICENSE",
		"plumbing/**/*.go",
	)
	FilterDFSPanic(err)

	// output storer
	outputfs := memory.NewStorage()

	newgraph, err := gitrim.FilterDFSPath(context.Background(), graph, outputfs, orfilter)
	FilterDFSPanic(err)

	// Note the result is deterministic
	fmt.Printf("From %d commits, generated %d commits.\nHead commit is:\n", len(graph), len(newgraph))

	commitinfo := newgraph[5].String()
	commitinfo = removeEmptyLines(strings.ReplaceAll(commitinfo, "\r\n", "\n"))
	fmt.Println(commitinfo)

	lastcommit := newgraph[88]
	fmt.Println("parents:")
	fmt.Println(lastcommit.ParentHashes[0])
	fmt.Println(lastcommit.ParentHashes[1])

	// Output:
	// From 241 commits, generated 89 commits.
	// Head commit is:
	// commit d5f3d5523dcd0e977f555831385eae31ccd8a30d
	// Author: cui fliter <imcusg@gmail.com>
	// Date:   Thu Sep 22 16:27:41 2022 +0800
	//
	//     *: fix some typos (#567)
	//
	//     Signed-off-by: cui fliter <imcusg@gmail.com>
	//
	//     Signed-off-by: cui fliter <imcusg@gmail.com>
	//
	// parents:
	// a6fae4bd1c424c3e7da6bc5c4ac8397a9f28db92
	// f09651ec4e2589543cf3ce89167c46cc43f3c0cd
}
