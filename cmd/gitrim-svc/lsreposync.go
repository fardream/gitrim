package main

import "github.com/spf13/cobra"

type lsRepoSyncCmd struct {
	*cobra.Command

	id        []string
	showmap   bool
	checkstat bool

	overrideFromBranch string
	overrideToBranch   string
}

func newLsRepoSyncCmd(torun func(cmd *cobra.Command, args []string)) *lsRepoSyncCmd {
	r := &lsRepoSyncCmd{
		Command: &cobra.Command{
			Use:   "ls",
			Short: "list repo syncs",
			Args:  cobra.NoArgs,
		},
	}

	r.Flags().StringArrayVarP(&r.id, "id", "i", r.id, "ids to list")
	r.Flags().StringVar(&r.overrideFromBranch, "from-branch", r.overrideFromBranch, "override from branch")
	r.Flags().StringVar(&r.overrideToBranch, "to-branch", r.overrideToBranch, "override to branch")
	r.Flags().BoolVarP(&r.checkstat, "check-repo-stat", "s", r.checkstat, "check repo stat")
	r.Flags().BoolVar(&r.showmap, "show-commit-map", r.showmap, "show commit map")

	r.Run = torun

	return r
}
