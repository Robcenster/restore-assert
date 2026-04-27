package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/formatter"
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
	f         formatter.Formatter
}

func NewPostgresContainer(cfg *config.Config, f formatter.Formatter) *PostgresContainer {
	return &PostgresContainer{cfg: cfg, f: f}
}

// Launching a temporary container
func (p *PostgresContainer) Start(ctx context.Context) error {
	// Testcontainers panic guard
	defer func() {
		if r := recover(); r != nil {
			p.f.Error("testcontainers panicked: %v", r)
		}
	}()

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
			hc.Memory = mem
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

// ExecuteRestore handles all the tedious work involved in creating a backup
func (p *PostgresContainer) ExecuteRestore(ctx context.Context, hostFilePath string) error {
	// Preliminary check on the hosting service
	absPath, err := filepath.Abs(hostFilePath)
	if err != nil {
		return fmt.Errorf("invalid host path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", absPath)
	}

	// Creating Safe Pathways
	baseName := filepath.Base(absPath)
	containerPath := path.Join("/tmp", baseName)

	bType, err := detectBackupType(absPath)
	if err != nil {
		return fmt.Errorf("backup detection failed: %w", err)
	}

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

	// Remove the file/folder from the container after the function completes
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.container.Exec(cleanupCtx, []string{"rm", "-rf", containerPath})
	}()

	cmd, err := buildRestoreCommand(p.cfg.Database, p.cfg.Restore, bType, containerPath)
	if err != nil {
		return err
	}

	exitCode, reader, err := p.container.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	// Reading logs
	const maxLogSize = 1 * 1024 * 1024 // 1MB
	outputBytes, _ := io.ReadAll(io.LimitReader(reader, maxLogSize))
	output := strings.TrimSpace(string(outputBytes))

	if exitCode != 0 {
		if len(output) > 0 {
			p.f.Error("Restore command failed with exit code %d. Details:", exitCode)
			p.f.Info("--- DUMP LOGS ---\n%s\n-----------------", output)
		}
		return fmt.Errorf("restore failed (exit code %d)", exitCode)
	}

	if len(output) > 0 {
		p.f.Warning("Restore completed with messages:\n%s", output)
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
