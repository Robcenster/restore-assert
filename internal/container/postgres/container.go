package postgres

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	"github.com/testcontainers/testcontainers-go"
	pgmod "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer реализует интерфейс container.Provider
type PostgresContainer struct {
	cfg       *config.Config
	container *pgmod.PostgresContainer
	connStr   string
}

// NewPostgresContainer — конструктор
func NewPostgresContainer(cfg *config.Config) *PostgresContainer {
	return &PostgresContainer{cfg: cfg}
}

// Start запускает контейнер
func (p *PostgresContainer) Start(ctx context.Context) error {
	dbCfg := p.cfg.Database
	dCfg := p.cfg.Docker

	opts := []testcontainers.ContainerCustomizer{
		pgmod.WithDatabase(dbCfg.DBName),
		pgmod.WithUsername(dbCfg.User),
		pgmod.WithPassword(dbCfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2 * time.Minute),
		),
	}

	opts = append(opts, testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
		hc.AutoRemove = dCfg.AutoRemove
		if dCfg.MemoryLimit != "" {
			mem, _ := units.RAMInBytes(dCfg.MemoryLimit)
			hc.Resources.Memory = mem
		}
	}))

	pgContainer, err := pgmod.Run(ctx, dCfg.Image, opts...)
	if err != nil {
		return fmt.Errorf("failed to run postgres: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return fmt.Errorf("failed to get conn string: %w", err)
	}

	p.container = pgContainer
	p.connStr = connStr
	return nil
}

// ExecuteRestore делает всю грязную работу по заливке бэкапа
func (p *PostgresContainer) ExecuteRestore(ctx context.Context, hostFilePath string) error {
	containerPath := "/tmp/dump_to_restore"

	// 1. Детектим тип бэкапа
	bType, err := detectBackupType(hostFilePath)
	if err != nil {
		return fmt.Errorf("backup detection failed: %w", err)
	}

	// 2. Копируем файл с хоста в контейнер
	if err := p.container.CopyFileToContainer(ctx, hostFilePath, containerPath, 0644); err != nil {
		return fmt.Errorf("failed to copy dump to container: %w", err)
	}

	// 3. Если это не DumpAll, создаем роли и экстеншены ПЕРЕД накатыванием данных
	if bType != "dumpall" {
		if err := p.initDatabaseSchema(ctx); err != nil {
			return fmt.Errorf("failed to init pre-restore schema: %w", err)
		}
	}

	// 4. Формируем команду (psql или pg_restore)
	cmd, err := buildRestoreCommand(p.cfg.Database, p.cfg.Restore, bType, containerPath)
	if err != nil {
		return err
	}

	// 5. Запускаем команду внутри контейнера
	exitCode, reader, err := p.container.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute restore command: %w", err)
	}

	outputBytes, _ := io.ReadAll(reader)
	output := string(outputBytes)

	if len(output) > 0 {
		fmt.Printf("\n--- [RESTORE LOGS] ---\n%s\n------------------\n", output)
	}

	if exitCode != 0 {
		return fmt.Errorf("restore command failed with exit code: %d", exitCode)
	}

	return nil
}

// initDatabaseSchema создает необходимые роли и расширения перед восстановлением.
// Мы делаем это только для обычных дампов (для dumpall это не нужно, там всё есть внутри).
func (p *PostgresContainer) initDatabaseSchema(ctx context.Context) error {
	dbCfg := p.cfg.Database

	// 1. Создаем недостающие роли
	for _, role := range dbCfg.Roles {
		// Используем анонимный блок DO в Postgres, чтобы проверить существование роли
		// и не упасть с ошибкой, если роль уже есть (например, дефолтная 'postgres')
		query := fmt.Sprintf(`
			DO $$ 
			BEGIN 
				IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s') THEN 
					CREATE ROLE %s; 
				END IF; 
			END $$;`, role, role)

		if err := p.execPsql(ctx, dbCfg.DBName, query); err != nil {
			return fmt.Errorf("failed to create role %s: %w", role, err)
		}
	}

	// 2. Создаем необходимые расширения
	for _, ext := range dbCfg.Extensions {
		query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\";", ext)
		if err := p.execPsql(ctx, dbCfg.DBName, query); err != nil {
			return fmt.Errorf("failed to create extension %s: %w", ext, err)
		}
	}

	return nil
}

// execPsql — маленькая вспомогательная функция, чтобы не дублировать код вызова Exec
func (p *PostgresContainer) execPsql(ctx context.Context, dbName, query string) error {
	cmd := []string{"psql", "-U", "postgres", "-d", dbName, "-c", query}

	exitCode, reader, err := p.container.Exec(ctx, cmd)
	if err != nil {
		return err
	}

	outputBytes, _ := io.ReadAll(reader)
	if exitCode != 0 {
		return fmt.Errorf("psql failed with code %d: %s", exitCode, string(outputBytes))
	}

	return nil
}

// GetConnectionString возвращает DSN
func (p *PostgresContainer) GetConnectionString() string {
	return p.connStr
}

// Stop тушит контейнер
func (p *PostgresContainer) Stop(ctx context.Context) error {
	if p.container != nil {
		return p.container.Terminate(ctx)
	}
	return nil
}
