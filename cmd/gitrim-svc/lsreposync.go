package main

import "github.com/spf13/cobra"

type lsRepoSyncCmd struct {
	*cobra.Command

	id []string
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

	r.Run = torun

	return r
}
