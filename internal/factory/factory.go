package factory

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/container"
	"github.com/Robcenster/restore-assert/internal/container/postgres"
	"github.com/Robcenster/restore-assert/internal/repository"
	repo "github.com/Robcenster/restore-assert/internal/repository/postgres"
)

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

func NewRepository(ctx context.Context, cfg *config.Config, cp container.Provider) (repository.DBRepository, error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		connStr := cp.GetConnectionString()
		if connStr == "" {
			return nil, fmt.Errorf("connection string is empty, is the container started?")
		}

		return repo.New(ctx, connStr)

	case config.EngineMSSQL:
		return nil, fmt.Errorf("mssql repository is not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported engine type: %s", cfg.Engine)
	}
}
