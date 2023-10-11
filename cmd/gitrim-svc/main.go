package main

import (
	"context"
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
}

func newRootCmd() *rootCmd {
	c := &rootCmd{
		Command: &cobra.Command{
			Use:   "gitrim-svc",
			Short: "gitrim webhook service",
			Args:  cobra.NoArgs,
		},
	}

	c.Flags().StringVarP(&c.configPath, "config", "c", c.configPath, "path to the configuration")

	c.Run = func(*cobra.Command, []string) {
		c.runSvc()
	}

	return c
}

func (c *rootCmd) runSvc() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := cmd.GetOrPanic(svc.ParseConfigYAML(cmd.GetOrPanic(os.ReadFile(c.configPath))))

	svc := cmd.GetOrPanic(svc.New(config))

	cmd.OrPanic(svc.Start(ctx))
}
