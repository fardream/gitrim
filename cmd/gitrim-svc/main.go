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
	syncToSubCmd    *syncToSubCmd
	lsRepoSyncCmd   *lsRepoSyncCmd

	webhookCmd *cobra.Command
}

func newRootCmd() *rootCmd {
	c := &rootCmd{
		Command: &cobra.Command{
			Use:   "gitrim-svc",
			Short: "gitrim webhook service",
			Args:  cobra.NoArgs,
		},
		webhookCmd: &cobra.Command{
			Use:   "web",
			Short: "run web server",
			Args:  cobra.NoArgs,
		},
	}

	c.PersistentFlags().StringVarP(&c.configPath, "config", "c", c.configPath, "path to the configuration")
	c.MarkPersistentFlagFilename("config")

	c.webhookCmd.Run = func(*cobra.Command, []string) {
		c.runWebhook()
	}

	c.initRepoSyncCmd = newInitRepoSyncCmd(func(*cobra.Command, []string) {
		c.runInitRepoSync()
	})
	c.syncToSubCmd = newSyncToSubCmd(func(*cobra.Command, []string) {
		c.runSyncToSub()
	})
	c.lsRepoSyncCmd = newLsRepoSyncCmd(func(cmd *cobra.Command, args []string) {
		c.runLs()
	})

	c.AddCommand(c.initRepoSyncCmd.Command, c.syncToSubCmd.Command, c.lsRepoSyncCmd.Command)

	return c
}

func (c *rootCmd) runInitRepoSync() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	svc := cmd.GetOrPanic(svc.New(config))
	defer svc.Close()

	filter := cmd.GetOrPanic(os.ReadFile(c.initRepoSyncCmd.filterFile))

	c.initRepoSyncCmd.request.Filter = string(filter)

	resp := cmd.GetOrPanic(svc.InitRepoSync(ctx, c.initRepoSyncCmd.request))

	fmt.Println(resp.String())
}

func (c *rootCmd) runWebhook() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	svc := cmd.GetOrPanic(svc.New(config))

	cmd.OrPanic(svc.Start(ctx))
}

func (c *rootCmd) runSyncToSub() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	s := cmd.GetOrPanic(svc.New(config))
	defer s.Close()

	resp := cmd.GetOrPanic(s.SyncToSubRepo(ctx, &svc.SyncToSubRepoRequest{
		Id:    c.syncToSubCmd.id,
		Force: c.syncToSubCmd.force,
	}))

	fmt.Println(resp.String())
}

func (c *rootCmd) runLs() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	s := cmd.GetOrPanic(svc.New(config))
	defer s.Close()

	for _, id := range c.lsRepoSyncCmd.id {
		resp, err := s.GetRepoSync(ctx, &svc.GetRepoSyncRequest{Id: id})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list id: %s\nerror:\n%s\n", id, err.Error())
		} else {
			fmt.Println(resp.RepoSync.String())
		}
	}
}
