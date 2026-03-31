package cli

import (
	"fmt"

	"github.com/Robcenster/restore-assert/internal/app"
	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/factory"
	"github.com/spf13/cobra"
)

// restore-assert check
func NewCheckCmd() *cobra.Command {
	var cfgFile string

	checkCmd := &cobra.Command{
		Use:   "check [path/to/backup.sql]",
		Short: "Запустить валидацию бэкапа",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupPath := args[0]
			ctx := cmd.Context()

			// 1. Загружаем конфиг
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("ошибка загрузки конфига: %w", err)
			}

			// 2. Используем ФАБРИКУ для создания нужных компонентов.
			containerProvider, err := factory.NewContainerProvider(cfg)
			if err != nil {
				return fmt.Errorf("failed to create container provider: %w", err)
			}

			// Запускаем контейнер (это нужно сделать ДО создания репозитория)
			err = containerProvider.Start(ctx)
			if err != nil {
				return fmt.Errorf("failed to start container: %w", err)
			}
			defer containerProvider.Stop(ctx)

			dbRepo, err := factory.NewRepository(ctx, cfg, containerProvider)
			if err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}

			// 3. Создаем оркестратор (Пайплайн) и передаем ему "чистые" интерфейсы
			pipeline := app.NewPipeline(containerProvider, dbRepo)

			// 4. Запускаем весь процесс
			if err := pipeline.RunCheck(ctx, backupPath); err != nil {
				return fmt.Errorf("проверка провалилась: %w", err)
			}

			fmt.Println("✨ Проверка успешно завершена!")
			return nil
		},
	}

	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "restore-config.yaml", "Путь к файлу конфигурации")

	return checkCmd
}
