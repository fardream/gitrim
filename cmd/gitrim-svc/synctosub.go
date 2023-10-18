package main

import "github.com/spf13/cobra"

type syncToSubCmd struct {
	*cobra.Command

	id    string
	force bool
}

func newSyncToSubCmd(torun func(*cobra.Command, []string)) *syncToSubCmd {
	r := &syncToSubCmd{
		Command: &cobra.Command{
			Use:   "sync-to-sub",
			Short: "sync from original repo to sub repo",
			Args:  cobra.NoArgs,
		},
	}

	r.Flags().StringVarP(&r.id, "id", "i", r.id, "id of the sync")
	r.MarkFlagRequired("id")
	r.Flags().BoolVarP(&r.force, "force", "f", r.force, "force push")

	r.Run = torun

	return r
}
