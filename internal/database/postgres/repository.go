package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type storage struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*storage, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &storage{Pool: pool}, nil
}

func (s *storage) Close() {
	s.Pool.Close()
}
