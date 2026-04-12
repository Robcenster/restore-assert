package cli

import (
	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/factory"
	"github.com/Robcenster/restore-assert/internal/formatter"
	"github.com/spf13/cobra"
)

func NewInspectCmd() *cobra.Command {
	var cfgFile string

	inspectCmd := &cobra.Command{
		Use:   "inspect [path/to/backup]",
		Short: "Inspect backup content and structure",
		Long:  "Spins up a container, restores the backup and prints a summary of tables and roles.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupPath := args[0]
			ctx := cmd.Context()

			cfg, err := config.Load(cfgFile)
			if err != nil {
				cfg = &config.Config{} // default
			}

			cp, err := factory.NewContainerProvider(cfg)
			if err != nil {
				return err
			}

			if err := cp.Start(ctx); err != nil {
				return err
			}
			defer cp.Stop(ctx)

			if err := cp.ExecuteRestore(ctx, backupPath); err != nil {
				return err
			}

			dbRepo, err := factory.NewRepository(ctx, cfg, cp)
			if err != nil {
				return err
			}
			defer dbRepo.Close()

			report, err := dbRepo.GetSimpleClusterReport(ctx)
			if err != nil {
				return err
			}

			formatter.PrintSimpleReport(report)
			return nil
		},
	}

	inspectCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "path to config file")
	return inspectCmd
}
