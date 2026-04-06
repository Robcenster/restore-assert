package postgres

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/formatter"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Repository{pool: pool}, nil
}

func (r *Repository) GetDatabaseInfo(ctx context.Context) (map[string][]formatter.DbObject, error) {
	query := `
    SELECT 
        n.nspname AS schema_name,
        c.relname AS object_name,
        CASE c.relkind
            WHEN 'r' THEN 'table'
            WHEN 'v' THEN 'view'
            WHEN 'm' THEN 'materialized_view'
            ELSE 'other'
        END AS object_type
    FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
      AND c.relkind IN ('r', 'v', 'm')
    ORDER BY schema_name, object_type, object_name;`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database info: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]formatter.DbObject)

	for rows.Next() {
		var schemaName string
		var objectName string
		var objectType string

		if err := rows.Scan(&schemaName, &objectName, &objectType); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		result[schemaName] = append(result[schemaName], formatter.DbObject{
			Name: objectName,
			Type: objectType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) InitializeEnvironment(ctx context.Context, roles []string, extensions []string) error {
	for _, role := range roles {
		query := fmt.Sprintf(`
			DO $$ 
			BEGIN 
				IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s') THEN 
					CREATE ROLE %s; 
				END IF; 
			END $$;`, role, role)

		_, err := r.pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create role %s: %w", role, err)
		}
	}

	for _, ext := range extensions {
		query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\";", ext)
		_, err := r.pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create extension %s: %w", ext, err)
		}
	}

	return nil
}

func (r *Repository) ExecuteQuery(ctx context.Context, query string) (string, error) {
	var result string

	err := r.pool.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения ассерта: %w", err)
	}

	return result, nil
}

func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
