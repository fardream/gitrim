package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/fardream/gitrim/cmd"
	"github.com/fardream/gitrim/svc"
)

func main() {
	newRootCmd().Execute()
}

type rootCmd struct {
	*cobra.Command

	configPath string

	initRepoSyncCmd *initRepoSyncCmd
	lsRepoSyncCmd   *lsRepoSyncCmd
	syncToSubCmd    *syncToSubCmd
	syncToFromCmd   *syncToFromCmd
}

func newRootCmd() *rootCmd {
	c := &rootCmd{
		Command: &cobra.Command{
			Use:   "gitrim-svc",
			Short: "gitrim webhook service",
			Args:  cobra.NoArgs,
		},
	}

	c.PersistentFlags().StringVarP(&c.configPath, "config", "c", c.configPath, "path to the configuration")
	c.MarkPersistentFlagFilename("config")

	c.initRepoSyncCmd = newInitRepoSyncCmd(func(*cobra.Command, []string) {
		c.runInitRepoSync()
	})
	c.syncToSubCmd = newSyncToSubCmd(func(*cobra.Command, []string) {
		c.runSyncToSub()
	})
	c.syncToFromCmd = newSyncToFromCmd(func(*cobra.Command, []string) {
		c.runSyncToFrom()
	})
	c.lsRepoSyncCmd = newLsRepoSyncCmd(func(*cobra.Command, []string) {
		c.runLs()
	})

	c.AddCommand(c.initRepoSyncCmd.Command, c.syncToSubCmd.Command, c.lsRepoSyncCmd.Command, c.syncToFromCmd.Command)

	return c
}

func (c *rootCmd) runInitRepoSync() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	s := cmd.GetOrPanic(svc.New(config))
	defer s.Close()

	filter := cmd.GetOrPanic(os.ReadFile(c.initRepoSyncCmd.filterFile))

	c.initRepoSyncCmd.request.Filter = string(filter)

	resp := cmd.GetOrPanic(s.InitRepoSync(ctx, c.initRepoSyncCmd.request))
	fmt.Println(PrintProtoText(resp))
}

func (c *rootCmd) runSyncToSub() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	s := cmd.GetOrPanic(svc.New(config))
	defer s.Close()

	resp := cmd.GetOrPanic(s.SyncToSubRepo(ctx,
		&svc.SyncToSubRepoRequest{
			Id:                 c.syncToSubCmd.id,
			Force:              c.syncToSubCmd.force,
			OverrideFromBranch: c.syncToSubCmd.overrideFromBranch,
			OverrideToBranch:   c.syncToSubCmd.overrideToBranch,
		}))

	fmt.Println(PrintProtoText(resp))
}

func (c *rootCmd) runLs() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	s := cmd.GetOrPanic(svc.New(config))
	defer s.Close()

printloop:
	for _, id := range c.lsRepoSyncCmd.id {
		resp, err := s.GetRepoSync(ctx, &svc.GetRepoSyncRequest{Id: id})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list id: %s\nerror:\n%s\n", id, err.Error())
			continue printloop
		}
		fmt.Println((PrintProtoText(resp.RepoSync)))
		if c.lsRepoSyncCmd.showmap {
			fmt.Println((PrintProtoText(resp.SyncStat)))
		} else {
			resp.SyncStat.FromDfs = nil
			clear(resp.SyncStat.FromToTo)
			resp.SyncStat.ToDfs = nil
			clear(resp.SyncStat.ToToFrom)
			fmt.Println((PrintProtoText(resp.SyncStat)))
		}
		if !c.lsRepoSyncCmd.checkstat {
			continue printloop
		}
		stat, err := s.CheckCommitsFromSubRepo(ctx,
			&svc.CheckCommitsFromSubRepoRequest{
				Id:                 id,
				OverrideFromBranch: c.lsRepoSyncCmd.overrideFromBranch,
				OverrideToBranch:   c.lsRepoSyncCmd.overrideToBranch,
			})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get stat:\n%s\n", err.Error())
			continue printloop
		}
		fmt.Println(PrintProtoText(stat))
	}
}

func (c *rootCmd) runSyncToFrom() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	s := cmd.GetOrPanic(svc.New(config))
	defer s.Close()

	if c.syncToFromCmd.noDryrun {
		resp := cmd.GetOrPanic(
			s.CommitsFromSubRepo(
				ctx,
				&svc.CommitsFromSubRepoRequest{
					Id:                 c.syncToFromCmd.id,
					OverrideFromBranch: c.syncToFromCmd.overrideFromBranch,
					OverrideToBranch:   c.syncToFromCmd.overrideToBranch,
					AllowPgpSignature:  c.syncToFromCmd.allowGpg,
					DoPush:             true,
				}))
		fmt.Println(PrintProtoText(resp))
	} else {
		resp := cmd.GetOrPanic(
			s.CheckCommitsFromSubRepo(
				ctx,
				&svc.CheckCommitsFromSubRepoRequest{
					Id:                 c.syncToFromCmd.id,
					OverrideFromBranch: c.syncToFromCmd.overrideFromBranch,
					OverrideToBranch:   c.syncToFromCmd.overrideToBranch,
					AllowPgpSignature:  c.syncToFromCmd.allowGpg,
				}))

		fmt.Println(PrintProtoText(resp))
	}
}
