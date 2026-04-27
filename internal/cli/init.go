package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var fileName string
	var filePath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new config by copying the template",
		RunE: func(cmd *cobra.Command, args []string) error {
			srcPath := filepath.Join("config", "template", "restore-config.yaml")

			ext := filepath.Ext(fileName)
			if ext != ".yaml" && ext != ".yml" {
				return fmt.Errorf("invalid file extension '%s': only .yaml is allowed", ext)
			}

			dstPath := filepath.Join(filePath, fileName)

			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				return fmt.Errorf("template file not found at %s. Please ensure it exists", srcPath)
			}

			if _, err := os.Stat(dstPath); err == nil {
				return fmt.Errorf("config file '%s' already exists", dstPath)
			}

			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", filePath, err)
			}

			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy template: %w", err)
			}

			fmt.Printf("Successfully initialized config: %s\n", dstPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&fileName, "name", "n", "restore-config.yaml", "Name of the configuration file")
	cmd.Flags().StringVarP(&filePath, "path", "p", "config", "Directory path where the config will be created")

	return cmd
}

// A utility function for copying a file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }() 

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }() 

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
