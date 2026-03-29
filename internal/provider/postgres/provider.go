// StartContainer, Restore, Init (Exec-логика)
package postgres

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresProvider struct {
	container *postgres.PostgresContainer
	connStr   string
}

func StartContainer(ctx context.Context, dCfg config.Docker, dbCfg config.Database) (*PostgresProvider, error) {
	// 1. Набор базовых опций
	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(dbCfg.DBName),
		postgres.WithUsername(dbCfg.User),
		postgres.WithPassword(dbCfg.Password),

		// Ждем, пока Postgres будет готов принимать соединения
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2 * time.Minute),
		),
	}

	// 2. Логика работы с конфигурацией Postgres
	if dbCfg.ConfigFile != "" {
		// Если пользователь передал путь — используем файл
		opts = append(opts, postgres.WithConfigFile(dbCfg.ConfigFile))
	} else if len(dbCfg.Settings) > 0 {
		// Иначе прокидываем настройки через флаги запуска -c
		var cmdArgs []string
		for key, value := range dbCfg.Settings {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("%s=%s", key, value))
		}
		opts = append(opts, testcontainers.WithConfigModifier(func(c *container.Config) {
			c.Cmd = append([]string{"postgres"}, cmdArgs...)
		}))
	}

	// 3. Имя контейнера (через официальный GenericContainerRequest)
	if dCfg.ContainerName != "" {
		opts = append(opts, testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name: dCfg.ContainerName,
			},
		}))
	}

	// 4. Ограничения ресурсов Docker (HostConfig)
	opts = append(opts, testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
		hc.AutoRemove = dCfg.AutoRemove
		if dCfg.MemoryLimit != "" {
			mem, _ := units.RAMInBytes(dCfg.MemoryLimit)
			hc.Resources.Memory = mem
		}
	}))

	// 5. Запуск через высокоуровневый модуль
	pgContainer, err := postgres.Run(ctx, dCfg.Image, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to run postgres container: %w", err)
	}

	// 6. Получение строки подключения
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &PostgresProvider{
		container: pgContainer,
		connStr:   connStr,
	}, nil
}

// ConnectionString возвращает DSN для подключения через pgx
func (p *PostgresProvider) ConnectionString() string {
	return p.connStr
}

// Close корректно завершает работу контейнера (используется в defer)
func (p *PostgresProvider) Close(ctx context.Context) error {
	if p.container != nil {
		return p.container.Terminate(ctx)
	}
	return nil
}

// InitDatabase создает необходимые роли и расширения перед восстановлением.
// Вызывается только для обычных дампов (не dumpall).
func (p *PostgresProvider) InitDatabase(ctx context.Context, dbCfg config.Database) error {
	// Создаем роли
	for _, role := range dbCfg.Roles {
		// Используем IF NOT EXISTS, чтобы избежать ошибок, если роль системная
		query := fmt.Sprintf("DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s') THEN CREATE ROLE %s; END IF; END $$;", role, role)
		cmd := []string{"psql", "-U", "postgres", "-d", dbCfg.DBName, "-c", query}

		exitCode, output, err := p.container.Exec(ctx, cmd)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to create role %s: code=%d, out=%s", role, exitCode, output)
		}
	}

	// Создаем расширения
	for _, ext := range dbCfg.Extensions {
		query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\";", ext)
		cmd := []string{"psql", "-U", "postgres", "-d", dbCfg.DBName, "-c", query}

		exitCode, output, err := p.container.Exec(ctx, cmd)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to create extension %s: code=%d, out=%s", ext, exitCode, output)
		}
	}

	return nil
}

// Restore выполняет полный цикл восстановления дампа в контейнер
func (p *PostgresProvider) Restore(ctx context.Context, hostDumpPath string, bType BackupType, cfg config.Config) error {
	// 1. Копируем файл с хоста (твоего ПК) в контейнер
	containerPath := "/tmp/dump_to_restore"
	if err := p.container.CopyFileToContainer(ctx, hostDumpPath, containerPath, 0644); err != nil {
		return fmt.Errorf("failed to copy dump to container: %w", err)
	}

	// 2. Инициализируем роли и расширения (ПРОПУСКАЕМ для DumpAll)
	if bType != TypeDumpAll {
		if err := p.InitDatabase(ctx, cfg.Database); err != nil {
			return fmt.Errorf("database init failed: %w", err)
		}
	}

	// 3. Формируем команду восстановления
	cmd, err := buildRestoreCommand(cfg.Database, cfg.Restore, bType, containerPath)
	if err != nil {
		return err
	}

	// 4. Выполняем восстановление
	exitCode, reader, err := p.container.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute restore command: %w", err)
	}
	// if exitCode != 0 {
	// 	return fmt.Errorf("restore command failed with exit code %d: %s", exitCode, output)
	// }

	// Читаем всё, что утилита написала в консоль (stdout + stderr)
	outputBytes, _ := io.ReadAll(reader)
	output := string(outputBytes)

	// ВАЖНО: Печатаем логи утилиты, чтобы видеть, что происходит
	if len(output) > 0 {
		fmt.Printf("\n--- [LOGS: %s] ---\n%s\n------------------\n", cmd[0], output)
	}

	if exitCode != 0 {
		return fmt.Errorf("restore failed with exit code %d", exitCode)
	}

	// 5. Выполняем Analyze, если требуется по конфигу
	if cfg.Restore.Analyze {
		targetDB := cfg.Database.DBName
		if bType == TypeDumpAll {
			targetDB = "postgres" // или можно подключиться к нужной, если мы знаем её имя
		}

		analyzeCmd := []string{"psql", "-U", cfg.Database.User, "-d", targetDB, "-c", "ANALYZE;"}
		analyzeCode, analyzeOut, _ := p.container.Exec(ctx, analyzeCmd)
		if analyzeCode != 0 {
			// Обычно падение Analyze не критично для самого факта восстановления,
			// но стоит залогировать или вернуть ошибку в зависимости от твоих требований.
			return fmt.Errorf("analyze failed with code %d: %s", analyzeCode, analyzeOut)
		}
	}

	return nil
}
