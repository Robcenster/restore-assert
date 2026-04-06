package cli

import (
	"fmt"

	"github.com/Robcenster/restore-assert/internal/app"
	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/factory"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	var cfgFile string

	checkCmd := &cobra.Command{
		Use:   "check [path/to/backup.sql]",
		Short: "Check database dump",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupPath := args[0]
			ctx := cmd.Context()

			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("loading config error: %v", err)
			}

			containerProvider, err := factory.NewContainerProvider(cfg)
			if err != nil {
				return fmt.Errorf("failed to create container provider: %v", err)
			}

			err = containerProvider.Start(ctx)
			if err != nil {
				return fmt.Errorf("failed to start container: %v", err)
			}
			defer containerProvider.Stop(ctx)

			dbRepo, err := factory.NewRepository(ctx, cfg, containerProvider)
			if err != nil {
				return fmt.Errorf("failed to create repository: %v", err)
			}

			pipeline := app.NewPipeline(containerProvider, dbRepo, cfg)

			if err := pipeline.RunCheck(ctx, backupPath); err != nil {
				return fmt.Errorf("проверка провалилась: %v", err)
			}
			return nil
		},
	}

	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "restore-config.yaml", "Configuration path")

	return checkCmd
}
