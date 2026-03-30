package repository

import (
	"context"
)

// DBRepository описывает контракт для выполнения SQL-проверок.
// Любая база данных (Postgres, MSSQL и т.д.) обязана уметь это делать.
type DBRepository interface {
	// GetDatabaseInfo выводит отладочную информацию о структуре базы
	GetDatabaseInfo(ctx context.Context) error
	Close()
}
