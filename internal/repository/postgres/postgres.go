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

// EnsureRoles проверяет наличие ролей и создает отсутствующие.
func (r *Repository) EnsureRoles(ctx context.Context, roles []string) error {
	for _, role := range roles {
		// В Postgres нет команды CREATE ROLE IF NOT EXISTS, поэтому используем анонимный блок
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

// EnsureExtensions гарантирует наличие расширений для ЛЮБЫХ типов дампов
func (r *Repository) EnsureExtensions(ctx context.Context, extensions []string, modifyTemplate bool) error {
	if len(extensions) == 0 {
		return nil
	}

	// ШАГ 1: Установка в текущую подключенную БД (для простых дампов без CREATE DATABASE)
	for _, ext := range extensions {
		query := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA public;`, ext)
		if _, err := r.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to install %s in target DB: %w", ext, err)
		}
	}

	if !modifyTemplate{
		return nil
	}

	fmt.Println("Изменение шаблона!")
	// ШАГ 2: Модификация системного template0 (для дампов с CREATE DATABASE)
	// 2.1 Разрешаем подключения к template0 (по умолчанию запрещено)
	unlockQuery := `UPDATE pg_database SET datallowconn = true WHERE datname = 'template0';`
	if _, err := r.pool.Exec(ctx, unlockQuery); err != nil {
		return fmt.Errorf("failed to unlock template0: %w", err)
	}

	// Возвращаем настройки безопасности при выходе из функции
	defer func() {
		lockQuery := `UPDATE pg_database SET datallowconn = false WHERE datname = 'template0';`
		_, _ = r.pool.Exec(context.Background(), lockQuery)
	}()

	// 2.2 Подключаемся напрямую к template0
	templateCfg := r.pool.Config().ConnConfig.Copy()
	templateCfg.Database = "template0"
	conn, err := pgx.ConnectConfig(ctx, templateCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to template0: %w", err)
	}
	defer conn.Close(ctx)

	// 2.3 Устанавливаем расширения в самый корень Postgres
	for _, ext := range extensions {
		query := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" SCHEMA public;`, ext)
		if _, err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to install %s in template0: %w", ext, err)
		}
	}

	return nil
}

func (r *Repository) Analyze(ctx context.Context) error {
	// Выполняем ANALYZE для всей текущей базы данных.
	// Это обновит pg_statistic, что критично для планировщика.
	_, err := r.pool.Exec(ctx, "ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to run ANALYZE: %w", err)
	}
	return nil
}

func (r *Repository) GetSimpleClusterReport(ctx context.Context) (*formatter.ClusterSnapshot, error) {
	cluster := &formatter.ClusterSnapshot{}

	// 1. Версия
	_ = r.pool.QueryRow(ctx, "SELECT version()").Scan(&cluster.Version)

	// 2. Роли (все, кроме системных)
	roleRows, _ := r.pool.Query(ctx, "SELECT rolname FROM pg_roles WHERE rolname NOT LIKE 'pg_%'")
	for roleRows.Next() {
		var name string
		roleRows.Scan(&name)
		cluster.Roles = append(cluster.Roles, name)
	}
	roleRows.Close()

	// 3. Список баз (исключаем шаблоны)
	var dbNames []string
	dbRows, _ := r.pool.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false")
	for dbRows.Next() {
		var name string
		dbRows.Scan(&name)
		dbNames = append(dbNames, name)
	}
	dbRows.Close()

	// 4. Сбор таблиц по каждой базе
	for _, dbName := range dbNames {
		dbSnap, err := r.getTablesForDB(ctx, dbName)
		if err != nil {
			// Если база закрыта или нет прав — просто идем дальше
			continue
		}
		cluster.Databases = append(cluster.Databases, *dbSnap)
	}

	return cluster, nil
}

func (r *Repository) getTablesForDB(ctx context.Context, dbName string) (*formatter.DatabaseSnapshot, error) {
	// Создаем новое временное подключение к конкретной БД
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
