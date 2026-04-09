package postgres

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	"github.com/testcontainers/testcontainers-go"
	pgmod "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	cfg       *config.Config
	container *pgmod.PostgresContainer
	connStr   string
}

func NewPostgresContainer(cfg *config.Config) *PostgresContainer {
	return &PostgresContainer{cfg: cfg}
}

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
	// 1. Предварительная проверка на хосте
	absPath, err := filepath.Abs(hostFilePath)
	if err != nil {
		return fmt.Errorf("invalid host path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", absPath)
	}

	// Формируем безопасные пути
	baseName := filepath.Base(absPath)
	containerPath := path.Join("/tmp", baseName)

	// 2. Детектим тип бэкапа
	bType, err := detectBackupType(absPath)
	if err != nil {
		return fmt.Errorf("backup detection failed: %w", err)
	}

	// 3. Копирование данных
	// Используем switch только для выбора метода, логику обработки ошибок выносим вниз
	var copyErr error
	switch bType {
	case TypeDirectory:
		copyErr = p.container.CopyDirToContainer(ctx, absPath, containerPath, 0755)
	case TypeCustom, TypeTar, TypePlain, TypeDumpAll:
		copyErr = p.container.CopyFileToContainer(ctx, absPath, containerPath, 0644)
	default:
		return fmt.Errorf("unsupported backup type: %s", bType)
	}

	if copyErr != nil {
		return fmt.Errorf("failed to transfer %s to container: %w", bType, copyErr)
	}

	// ГАРАНТИРОВАННАЯ ОЧИСТКА: Удаляем файл/папку из контейнера после завершения функции
	// Это важно, если контейнер живет долго (например, во время запуска пачки тестов)
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.container.Exec(cleanupCtx, []string{"rm", "-rf", containerPath})
	}()

	// 5. Формирование и запуск команды
	cmd, err := buildRestoreCommand(p.cfg.Database, p.cfg.Restore, bType, containerPath)
	if err != nil {
		return err
	}

	exitCode, reader, err := p.container.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	// БЕЗОПАСНОЕ ЧТЕНИЕ ЛОГОВ: Ограничиваем размер, чтобы не упасть по OOM
	// если дамп выдаст миллион ворнингов
	const maxLogSize = 1 * 1024 * 1024 // 1MB
	outputBytes, _ := io.ReadAll(io.LimitReader(reader, maxLogSize))
	output := string(outputBytes)

	if exitCode != 0 {
		return fmt.Errorf("restore failed (code %d). Logs:\n%s", exitCode, output)
	}

	if len(output) > 0 {
		// Используем логгер вместо fmt для гибкости
		log.Printf("Restore logs for %s:\n%s", baseName, output)
	}
	return nil
}

func (p *PostgresContainer) GetConnectionString() string {
	return p.connStr
}

func (p *PostgresContainer) Stop(ctx context.Context) error {
	if p.container != nil {
		return p.container.Terminate(ctx)
	}
	return nil
}
