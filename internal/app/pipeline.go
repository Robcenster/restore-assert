package app

import (
	"context"
	"fmt"
	"path/filepath"

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
	formatter formatter.Formatter
}

func NewPipeline(ct container.Container, repo repository.DBRepository, cfg *config.Config, f formatter.Formatter) *Pipeline {
	return &Pipeline{
		container: ct,
		repo:      repo,
		cfg:       cfg,
		verifier:  verifier.NewVerifier(repo),
		formatter: f,
	}
}

func (p *Pipeline) RunCheck(ctx context.Context, backupPath string) error {
	if err := p.prepareEnvironment(ctx); err != nil {
		return err
	}

	p.formatter.Step(fmt.Sprintf("Restoring dump: %s", filepath.Base(backupPath)))
	if err := p.container.ExecuteRestore(ctx, backupPath); err != nil {
		return fmt.Errorf("restoring database error: %w", err)
	}
	p.formatter.Success("Dump was successfully deployed!")

	if p.cfg.Restore.Analyze {
		p.formatter.Step("Run ANALYZE (collect statistics)")
		if err := p.repo.Analyze(ctx); err != nil {
			p.formatter.Warning("⚠️ Warning ANALYZE: %v", err)
		} else {
			p.formatter.Success("Database statistics have been updated!")
		}
	}

	p.formatter.Step("Running asserts")
	failedAssertCount := 0

	if len(p.cfg.Asserts.Existence.Extensions) == 0 &&
		len(p.cfg.Asserts.Tables) == 0 {

		p.formatter.Info("No logical tests have been added")
		fmt.Println("No logic tests in config file")
		return nil
	}

	tasks := p.verifier.CreateTasks(p.cfg.Asserts)
	for _, assert := range tasks {
		success, err := p.verifier.RunAssert(ctx, assert)
		if err != nil && !success {
			p.formatter.Error("'%s': %v", assert.Name, err)
			failedAssertCount++
			continue
		}
		if !success {
			p.formatter.Error("Failed: %s", assert.Name)
			failedAssertCount++
		} else if p.cfg.Restore.ShowSuccessTests {
			p.formatter.Success("%s", assert.Name)
		}

	}

	if p.cfg.Restore.ShowDatabaseInfo {
		report, err := p.repo.GetSimpleClusterReport(ctx)
		if err != nil {
			return err
		}
		p.formatter.PrintClusterReport(report)
	}

	if failedAssertCount > 0 {
		p.formatter.Info("Total failed asserts: %d", failedAssertCount)
		return fmt.Errorf("%d assertions failed", failedAssertCount)
	}

	return nil
}

func (p *Pipeline) prepareEnvironment(ctx context.Context) error {
	if len(p.cfg.Database.Roles) > 0 {
		p.formatter.Step("Сreating database roles")
		if err := p.repo.EnsureRoles(ctx, p.cfg.Database.Roles); err != nil {
			return fmt.Errorf("setup roles error: %w", err)
		}
		p.formatter.Success("Roles created successfully!")
	}

	if len(p.cfg.Database.Extensions) > 0 {
		p.formatter.Step("Installing extensions")
		if err := p.repo.EnsureExtensions(ctx, p.cfg.Database.Extensions, p.cfg.Restore.ModifyTemplate); err != nil {
			return fmt.Errorf("setup extensions error: %w", err)
		}
		p.formatter.Success("Extensions installed!")
	}
	return nil
}
