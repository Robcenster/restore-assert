package repository

import (
	"context"

	"github.com/Robcenster/restore-assert/internal/formatter"
)

type DBRepository interface {
	ExecuteQuery(ctx context.Context, query string) (string, error)
	EnsureRoles(ctx context.Context, roles []string) error
	EnsureExtensions(ctx context.Context, extensions []string) error
	GetDatabaseInfo(ctx context.Context) (map[string][]formatter.DbObject, error)
	Close()
}
