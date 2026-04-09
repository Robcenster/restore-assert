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

	// 1. Создаем роли, если они указаны в конфиге
	if len(p.cfg.Database.Roles) > 0 {
		if err := p.repo.EnsureRoles(ctx, p.cfg.Database.Roles); err != nil {
			return fmt.Errorf("setup roles error: %w", err)
		}
	}

	// 2. Подключаем расширения, если они указаны
	// !!!Необходимо в конфиг добавить параметр, если в дампе есть create db!!!
	if len(p.cfg.Database.Extensions) > 0 {
		if err := p.repo.EnsureExtensions(ctx, p.cfg.Database.Extensions, p.cfg.Restore.ModifyTemplate); err != nil {
			return fmt.Errorf("setup extensions error: %w", err)
		}
	}

	fmt.Println("⏳ [Step 1/3] Restoring database...")
	if err := p.container.ExecuteRestore(ctx, backupPath); err != nil {
		return fmt.Errorf("restoring database error: %w", err)
	}

	fmt.Println("📊 [Step 2/3] Database restored:")

	if p.cfg.Restore.ShowDatabaseInfo {
		report, err := p.repo.GetSimpleClusterReport(ctx)
		if err != nil {
			return err
		}
		formatter.PrintSimpleReport(report)
	}

	fmt.Print("🔍 [Step 3/4] Running ANALYZE to update statistics... ")

	if p.cfg.Restore.Analyze {
		if err := p.repo.Analyze(ctx); err != nil {
			// Мы можем либо прервать выполнение, либо просто залогировать ошибку,
			// так как база технически уже восстановлена.
			fmt.Printf("⚠️ Warning: %v\n", err)
		} else {
			fmt.Println("Done!")
		}
	}

	fmt.Println("🧪 [Step 4/4] Running asserts...")
	failedAssertCount := 0

	// Проверяем, есть ли хоть одна проверка внутри объекта Asserts
	if len(p.cfg.Asserts.Existence.Extensions) == 0 &&
		len(p.cfg.Asserts.Tables) == 0 {

		fmt.Println("ℹ️ [Step 4/4] No logic tests in config file")
		return nil
	}

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
