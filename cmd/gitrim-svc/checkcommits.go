package main

import "github.com/spf13/cobra"

type syncToFromCmd struct {
	*cobra.Command

	id       string
	noDryrun bool

	overrideFromBranch string
	overrideToBranch   string
	allowGpg           bool
}

func newSyncToFromCmd(torun func(*cobra.Command, []string)) *syncToFromCmd {
	r := &syncToFromCmd{
		Command: &cobra.Command{
			Use:   "sync-to-from",
			Short: "sync changes to original repo",
			Args:  cobra.NoArgs,
		},
	}

	r.Flags().StringVarP(&r.id, "id", "i", r.id, "id of the sync")
	r.MarkFlagRequired("id")
	r.Flags().BoolVarP(&r.noDryrun, "no-dryrun", "p", r.noDryrun, "push the changes, instead of dryrun/check the commits")
	r.Flags().StringVar(&r.overrideFromBranch, "from-branch", r.overrideFromBranch, "override from branch")
	r.Flags().StringVar(&r.overrideToBranch, "to-branch", r.overrideToBranch, "override to branch")
	r.Flags().BoolVar(&r.allowGpg, "allow-gpg", r.allowGpg, "allow gpg signatures in the commits")

	r.Run = torun

	return r
}
