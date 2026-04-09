package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "restore-assert",
		Short: "Restore-Assert: Burn-test your backups before they are needed.",
		Long:  `A CLI tool to validate database backups by restoring them into temporary Docker containers and running SQL-based assertions.`,
	}

	rootCmd.AddCommand(NewCheckCmd())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
