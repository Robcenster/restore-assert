package cli

import (
	"fmt"
	"os"

	"github.com/Robcenster/restore-assert/internal/app"
	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/factory"
	"github.com/Robcenster/restore-assert/internal/formatter"
	"github.com/Robcenster/restore-assert/internal/formatter/terminal"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	var cfgFile string

	checkCmd := &cobra.Command{
		Use:           "check [path/to/backup.sql]",
		Short:         "Check database dump",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			backupPath := args[0]
			ctx := cmd.Context()

			var f formatter.Formatter = terminal.NewPrinter(os.Stdout)

			// Print the error using the formatter, then return the error to Cobra so that it calls os.Exit(1)
			fatal := func(msg string, err error) error {
				f.Error("%s: %v", msg, err)
				return err
			}

			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fatal("loading config error", err)
			}

			containerProvider, err := factory.NewContainerProvider(cfg, f)
			if err != nil {
				return fatal("failed to create container provider: %v", err)
			}

			f.Step("Starting container")
			err = containerProvider.Start(ctx)
			if err != nil {
				return fatal("failed to start container: %v", err)
			}
			defer containerProvider.Stop(ctx)

			dbRepo, err := factory.NewRepository(ctx, cfg, containerProvider)
			if err != nil {
				return fatal("failed to create repository: %v", err)
			}

			pipeline := app.NewPipeline(containerProvider, dbRepo, cfg, f)

			if err := pipeline.RunCheck(ctx, backupPath); err != nil {
				return fmt.Errorf("check failed: %w", err)
			}
			return nil
		},
	}

	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "restore-config.yaml", "Configuration path")

	return checkCmd
}
