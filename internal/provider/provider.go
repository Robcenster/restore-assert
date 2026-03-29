// Описание интерфеса Provider
package provider

import (
	"context"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/provider/postgres"
)

type Provider interface {
	// Подготовка (роли, расширения, юзеры)
	InitDatabase(ctx context.Context, dbCfg config.Database) error

	// Само восстановление
	Restore(ctx context.Context, hostFilePath string, bType postgres.BackupType, cfg config.Config) error

	// Получение DSN для подключения репозитория (pgx/mysql-driver)
	ConnectionString() string

	// Остановка и удаление контейнера
	Close(ctx context.Context) error
}
