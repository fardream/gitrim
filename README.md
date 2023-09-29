# gitrim

Git Trim is a deterministic tool to manipulate trees contained in git commits,
and it is based on the excellent [go-git](https://github.com/go-git/go-git) library.

[![Go Reference](https://pkg.go.dev/badge/github.com/fardream/gitrim.svg)](https://pkg.go.dev/github.com/fardream/gitrim)

## Trim/Filter Git History

Often, the history and files contained in a git repository need to be filtered in some way.
One simple case will be a contributor is only allowed to access part of the repo.
This can be done through `git-filter-branch`, although no so user-friendly.

`gitrim` does just that:

1. read git commit history.
1. from start, filter the tree contained in the commit and copy over author,
   committor, commit message. The parents are replaced with the newly created commits,
   and GPG signatures are omitted.

As long as the filters don't change, the generated git history is deterministic and can be one-to-one mapped back to the original repo.

Modifications made in the trimmed/filtered repo can be recreated by

1. filter the changes, and reject any changes that don't pass the filter.
1. apply them back to the original repo, copying over author, committor, commit message, and add the original commits as parents.
   GPG signatures are again omitted.

The commits in the filtered/trimmed repo will match the commit reproduced from original repo if they are without GPG signatures.

## Filters

The filter all implements the [Filter](https://pkg.go.dev/github.com/fardream/gitrim#Filter) interface

The pattern used is a more restricted version of the pattern used by [.gitignore](https://git-scm.com/docs/gitignore).

- `**` is for multi level directories, and it can only appear once in the match.
- `*` is for match one level of names.
- `!` and escapes are unsupported.
- paths are always relative to the root (the leading `/` is implicit).
  For example, `LICENSE` will only match `LICENSE` in the root of the repo.
  To match `LICENSE` at all directory levels, use `**/LICENSE`.

Refer to documentation on [`PatternFilter`](https://pkg.go.dev/github.com/fardream/gitrim#PatternFilter)

## DotGit

`gitrim`, through [go-git](https://github.com/go-git/go-git), operates on the contents of `.git` (or dotgit) folder (the commit,
blob, and tree objects).

## Example

See [Examples](https://pkg.go.dev/github.com/fardream/gitrim#pkg-examples)

## CLI

- [filter-git-hist](cmd/filter-git-hist) filters the history of a git repo and output it to another git repo.
- [expand-git-commit](cmd/expand-git-commit) expands the new commit back to the original repo.
- [dump-git-tree](cmd/dump-git-tree) prints the files of a branch/tree/commit/head. Optionally filters can be applied.
- [remve-git-gpg](cmd/remove-git-gpg) removes gpg signatures for commits.
