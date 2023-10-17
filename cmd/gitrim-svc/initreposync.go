package main

import (
	"github.com/spf13/cobra"

	"github.com/fardream/gitrim/svc"
)

type initRepoSyncCmd struct {
	*cobra.Command

	filterFile string

	request *svc.InitRepoSyncRequest
}

func newInitRepoSyncCmd(torun func(*cobra.Command, []string)) *initRepoSyncCmd {
	r := &initRepoSyncCmd{
		Command: &cobra.Command{
			Use:   "init-repo-sync",
			Short: "initialize a repo sync",
			Args:  cobra.NoArgs,
		},
		request: &svc.InitRepoSyncRequest{
			FromRepo: &svc.GitRepoIdentifier{},
			ToRepo:   &svc.GitRepoIdentifier{},
		},
	}

	r.Flags().StringVarP(&r.filterFile, "filter", "f", r.filterFile, "file contains the filters for this repo sync")
	r.MarkFlagFilename("filter")
	r.Flags().StringVar(&r.request.FromRepo.RemoteName, "from-remote", r.request.FromRepo.RemoteName, "from remote")
	r.Flags().StringVar(&r.request.FromRepo.Owner, "from-owner", r.request.FromRepo.RemoteName, "from owner")
	r.Flags().StringVar(&r.request.FromRepo.Repo, "from-repo", r.request.FromRepo.RemoteName, "from repo")
	r.Flags().StringVar(&r.request.FromBranch, "from-branch", r.request.FromRepo.RemoteName, "from branch")
	r.MarkFlagRequired("from-remote")
	r.MarkFlagRequired("from-owner")
	r.MarkFlagRequired("from-repo")
	r.MarkFlagRequired("from-branch")

	r.Flags().StringVar(&r.request.ToRepo.RemoteName, "to-remote", r.request.FromRepo.RemoteName, "to remote")
	r.Flags().StringVar(&r.request.ToRepo.Owner, "to-owner", r.request.FromRepo.RemoteName, "to owner")
	r.Flags().StringVar(&r.request.ToRepo.Repo, "to-repo", r.request.FromRepo.RemoteName, "to repo")
	r.Flags().StringVar(&r.request.ToBranch, "to-branch", r.request.FromRepo.RemoteName, "to branch")
	r.MarkFlagRequired("to-remote")
	r.MarkFlagRequired("to-owner")
	r.MarkFlagRequired("to-repo")
	r.MarkFlagRequired("to-branch")

	r.Run = torun

	return r
}
