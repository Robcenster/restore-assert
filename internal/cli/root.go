package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:           "restore-assert",
		Short:         "Burn-test your backups before they are needed",
		Long:          `Restore-Assert is a CLI tool to validate database backups by restoring them into temporary Docker containers and running SQL-based assertions.`,
		SilenceUsage:  true, // an option to silence usage when an error occurs
		SilenceErrors: true, // an option to quiet errors down stream
	}

	rootCmd.SetHelpTemplate(simplifiedHelpTemplate)

	rootCmd.AddCommand(
		NewCheckCmd(),
		NewInitCmd(),
		NewVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("restore-assert v%s\n", Version)
		},
	}
}

const simplifiedHelpTemplate = `
	Usage:{{if .Runnable}}
	{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
	{{.CommandPath}} [command]{{end}}

	Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	{{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}

	Flags:
	{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

	Use "{{.CommandPath}} [command] --help" for more information about a command.`
