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
	container container.Container
	repo      repository.DBRepository
	cfg       *config.Config
	verifier  *verifier.Verifier
}

func NewPipeline(ct container.Container, repo repository.DBRepository, cfg *config.Config) *Pipeline {
	return &Pipeline{
		container: ct,
		repo:      repo,
		cfg:       cfg,
		verifier:  verifier.NewVerifier(repo),
	}
}

func (p *Pipeline) RunCheck(ctx context.Context, backupPath string) error {
	if len(p.cfg.Database.Roles) > 0 {
		if err := p.repo.EnsureRoles(ctx, p.cfg.Database.Roles); err != nil {
			return fmt.Errorf("setup roles error: %w", err)
		}
	}

	// TODO: modify template if in dump create db appears
	if len(p.cfg.Database.Extensions) > 0 {
		if err := p.repo.EnsureExtensions(ctx, p.cfg.Database.Extensions, p.cfg.Restore.ModifyTemplate); err != nil {
			return fmt.Errorf("setup extensions error: %w", err)
		}
	}

	fmt.Println("[Step 1/4] Restoring database...")
	if err := p.container.ExecuteRestore(ctx, backupPath); err != nil {
		return fmt.Errorf("restoring database error: %w", err)
	}

	fmt.Println("[Step 2/4] Database restored:")

	if p.cfg.Restore.ShowDatabaseInfo {
		report, err := p.repo.GetSimpleClusterReport(ctx)
		if err != nil {
			return err
		}
		formatter.PrintSimpleReport(report)
	}

	fmt.Print("[Step 3/4] Running ANALYZE to update statistics... ")

	if p.cfg.Restore.Analyze {
		if err := p.repo.Analyze(ctx); err != nil {
			fmt.Printf("⚠️ Warning, analyze error: %v\n", err)
		} else {
			fmt.Println("Done!")
		}
	}

	fmt.Println("[Step 4/4] Running asserts...")
	failedAssertCount := 0

	if len(p.cfg.Asserts.Existence.Extensions) == 0 &&
		len(p.cfg.Asserts.Tables) == 0 {

		fmt.Println("No logic tests in config file")
		return nil
	}

	tasks := p.verifier.CreateTasks(p.cfg.Asserts)
	for _, assert := range tasks {
		success, err := p.verifier.RunAssert(ctx, assert)
		if err != nil && !success {
			fmt.Printf("❌ Error executing assert '%s': %v\n", assert.Name, err)
			failedAssertCount++
			continue
		}
		if !success {
			fmt.Printf("❌ Assert failed: %s\n", assert.Name)
			failedAssertCount++
		} else {
			if p.cfg.Restore.ShowSuccessTests {
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
