package app

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/container"
	"github.com/Robcenster/restore-assert/internal/formatter"
	"github.com/Robcenster/restore-assert/internal/repository"
	"github.com/Robcenster/restore-assert/internal/verifier"
)

type Pipeline struct {
	container container.Provider
	repo      repository.DBRepository
	cfg       *config.Config
	verifier  *verifier.Verifier
}

func NewPipeline(ct container.Provider, repo repository.DBRepository, cfg *config.Config) *Pipeline {
	return &Pipeline{
		container: ct,
		repo:      repo,
		cfg:       cfg,
		verifier:  verifier.NewVerifier(repo),
	}
}

func (p *Pipeline) RunCheck(ctx context.Context, backupPath string) error {
	if err := p.repo.InitializeEnvironment(ctx, p.cfg.Database.Roles, p.cfg.Database.Extensions); err != nil {
		return fmt.Errorf("roles and extensions create error: %w", err)
	}

	fmt.Println("⏳ [Step 1/3] Restoring database...")
	if err := p.container.ExecuteRestore(ctx, backupPath); err != nil {
		return fmt.Errorf("restoring database error: %w", err)
	}

	fmt.Println("📊 [Step 2/3] Database restored:")
	if p.cfg.Restore.ShowDatabaseInfo {
		dbStructure, err := p.repo.GetDatabaseInfo(ctx)
		if err != nil {
			return fmt.Errorf("getting database info error: %w", err)
		}

		formatter.PrintDatabaseStructure(dbStructure)
	}

	// if len(p.cfg.Asserts) == 0 {
	// 	fmt.Println("ℹ️ [Step 3/3] No logic tests in config file")
	// 	return nil
	// }

	fmt.Println("🧪 [Step 3/3] Running asserts...")
	failedAssertCount := 0

	// НОВАЯ СТРОКА: Превращаем вложенный конфиг в плоский список задач
	tasks := p.verifier.CreateTasks(p.cfg.Asserts)

	// ТВОЙ ОРИГИНАЛЬНЫЙ ЦИКЛ, но теперь перебирает tasks
	for _, assert := range tasks {
		success, err := p.verifier.RunAssert(ctx, assert) // assert теперь имеет тип AssertTask
		if err != nil && !success {
			fmt.Printf("❌ Error executing assert '%s': %v\n", assert.Name, err)
			failedAssertCount++
			continue
		}
		if !success {
			fmt.Printf("❌ Assert failed: %s\n", assert.Name)
			failedAssertCount++
		} else {
			if !p.cfg.Restore.HideSuccessTests {
				fmt.Printf("✅ Assert passed: %s\n", assert.Name)
			}
		}
	}

	if failedAssertCount > 0 {
		return fmt.Errorf("total failed asserts: %d", failedAssertCount)
	}
	fmt.Println("✅ All asserts completed successfully!")
	return nil
}
