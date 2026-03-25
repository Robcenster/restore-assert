package postgres

import (
	"context"
	"fmt"
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
