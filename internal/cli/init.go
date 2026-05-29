package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	restoreassert "github.com/Robcenster/restore-assert"
	"github.com/spf13/cobra"
)


func NewInitCmd() *cobra.Command {
	var fileName string
	var filePath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new config by copying the template",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Проверяем расширение
			ext := filepath.Ext(fileName)
			if ext != ".yaml" && ext != ".yml" {
				return fmt.Errorf("invalid file extension '%s': only .yaml is allowed", ext)
			}

			// 2. Формируем путь назначения относительно места запуска команды (CWD)
			dstPath := filepath.Join(filePath, fileName)

			// 3. Проверяем, не существует ли уже конфиг, чтобы случайно его не затереть
			if _, err := os.Stat(dstPath); err == nil {
				return fmt.Errorf("config file '%s' already exists", dstPath)
			}

			// 4. Создаем целевую директорию (если пользователь указал кастомный путь)
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", filePath, err)
			}

			// 5. Записываем встроенные в .exe байты шаблона в файл на диске
			if err := os.WriteFile(dstPath, restoreassert.DefaultTemplate, 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Successfully initialized config: %s\n", dstPath)
			return nil
		},
	}

	// Флаги остаются прежними
	cmd.Flags().StringVarP(&fileName, "name", "n", "restore-config.yaml", "Name of the configuration file")
	cmd.Flags().StringVarP(&filePath, "path", "p", "config", "Directory path where the config will be created")

	return cmd
}
