package repository

import (
	"context"

	"github.com/Robcenster/restore-assert/internal/formatter"
)

type DBRepository interface {
	ExecuteQuery(ctx context.Context, query string) (string, error)
	EnsureRoles(ctx context.Context, roles []string) error
	EnsureExtensions(ctx context.Context, extensions []string, modifyTemplate bool) error
	Analyze(ctx context.Context) error
	GetSimpleClusterReport(ctx context.Context) (*formatter.ClusterSnapshot, error)
	Close()
}
