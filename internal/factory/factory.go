package factory

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/container"                // Твой пакет с интерфейсом
	"github.com/Robcenster/restore-assert/internal/container/postgres"       // Твоя реализация
	"github.com/Robcenster/restore-assert/internal/repository"               // Твой пакет с интерфейсом репозитория
	repo "github.com/Robcenster/restore-assert/internal/repository/postgres" // Твоя реализация репозитория
)

// NewContainerProvider создает нужный контейнер на основе конфига
func NewContainerProvider(cfg *config.Config) (container.Provider, error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		return postgres.NewPostgresContainer(cfg), nil

	case config.EngineMSSQL:
		return nil, fmt.Errorf("mssql container provider is not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported engine type: %s", cfg.Engine)
	}
}

// NewRepository создает нужный репозиторий на основе конфига.
// Мы передаем сюда ctx и созданный ранее cp (container.Provider).
func NewRepository(ctx context.Context, cfg *config.Config, cp container.Provider) (repository.DBRepository, error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		// 1. Вытаскиваем строку подключения из провайдера контейнера
		connStr := cp.GetConnectionString()
		if connStr == "" {
			return nil, fmt.Errorf("connection string is empty, is the container started?")
		}

		// 2. Вызываем конструктор репозитория, передавая DSN
		return repo.New(ctx, connStr)

	case config.EngineMSSQL:
		return nil, fmt.Errorf("mssql repository is not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported engine type: %s", cfg.Engine)
	}
}
