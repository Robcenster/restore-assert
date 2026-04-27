package postgres

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/formatter"
	"github.com/jackc/pgx/v5"
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

// EnsureRoles checks for the existence of roles and creates any that are missing
func (r *Repository) EnsureRoles(ctx context.Context, roles []string) error {
	for _, role := range roles {
		query := fmt.Sprintf(`
			DO $$ 
			BEGIN 
				IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s') THEN 
					CREATE ROLE %s; 
				END IF; 
			END $$;`, role, role)

		if _, err := r.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to ensure role %s: %w", role, err)
		}
	}
	return nil
}

// EnsureExtensions guarantees the availability of extensions for ANY type of dump
func (r *Repository) EnsureExtensions(ctx context.Context, extensions []string, modifyTemplate bool) error {
	if len(extensions) == 0 {
		return nil
	}

	for _, ext := range extensions {
		query := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA public;`, ext)
		if _, err := r.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to install %s in target DB: %w", ext, err)
		}
	}

	if !modifyTemplate {
		return nil
	}

	unlockQuery := `UPDATE pg_database SET datallowconn = true WHERE datname = 'template0';`
	if _, err := r.pool.Exec(ctx, unlockQuery); err != nil {
		return fmt.Errorf("failed to unlock template0: %w", err)
	}

	defer func() {
		lockQuery := `UPDATE pg_database SET datallowconn = false WHERE datname = 'template0';`
		_, _ = r.pool.Exec(context.Background(), lockQuery)
	}()

	templateCfg := r.pool.Config().ConnConfig.Copy()
	templateCfg.Database = "template0"
	conn, err := pgx.ConnectConfig(ctx, templateCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to template0: %w", err)
	}
	defer conn.Close(ctx)

	for _, ext := range extensions {
		query := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA public;`, ext)
		if _, err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to install %s in template0: %w", ext, err)
		}
	}

	return nil
}

func (r *Repository) Analyze(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to run ANALYZE: %w", err)
	}
	return nil
}

// GetSimpleClusterReport provides information about the database cluster
func (r *Repository) GetSimpleClusterReport(ctx context.Context) (*formatter.ClusterSnapshot, error) {
	cluster := &formatter.ClusterSnapshot{}

	_ = r.pool.QueryRow(ctx, "SELECT version()").Scan(&cluster.Version)

	roleRows, _ := r.pool.Query(ctx, "SELECT rolname FROM pg_roles WHERE rolname NOT LIKE 'pg_%'")
	for roleRows.Next() {
		var name string
		roleRows.Scan(&name)
		cluster.Roles = append(cluster.Roles, name)
	}
	roleRows.Close()

	var dbNames []string
	dbRows, _ := r.pool.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false")
	for dbRows.Next() {
		var name string
		dbRows.Scan(&name)
		dbNames = append(dbNames, name)
	}
	dbRows.Close()

	for _, dbName := range dbNames {
		dbSnap, err := r.getTablesForDB(ctx, dbName)
		if err != nil {
			continue
		}
		cluster.Databases = append(cluster.Databases, *dbSnap)
	}

	return cluster, nil
}

// GetTablesForDB provides information about database tables
func (r *Repository) getTablesForDB(ctx context.Context, dbName string) (*formatter.DatabaseSnapshot, error) {
	cfg := r.pool.Config().ConnConfig.Copy()
	cfg.Database = dbName
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	snap := &formatter.DatabaseSnapshot{
		Name:    dbName,
		Schemas: make(map[string][]string),
	}

	query := `
		SELECT n.nspname, c.relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND c.relkind = 'r'
		ORDER BY n.nspname, c.relname`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err == nil {
			snap.Schemas[schema] = append(snap.Schemas[schema], table)
		}
	}

	return snap, nil
}

// ExecuteQuery processes the requests submitted to it
func (r *Repository) ExecuteQuery(ctx context.Context, query string) (string, error) {
	var result string

	err := r.pool.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("db query error: %w", err)
	}

	return result, nil
}

// Close closes the connection to the database
func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
