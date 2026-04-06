package app

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/container"
	"github.com/Robcenster/restore-assert/internal/repository"
)

type Pipeline struct {
	cp   container.Provider
	repo repository.DBRepository
}

func NewPipeline(cp container.Provider, repo repository.DBRepository) *Pipeline {
	return &Pipeline{
		cp:   cp,
		repo: repo,
	}
}

func (p *Pipeline) RunCheck(ctx context.Context, backupPath string) error {
	fmt.Println("🚀 Шаг 1: Начинаем заливку бэкапа в контейнер...")

	if err := p.cp.ExecuteRestore(ctx, backupPath); err != nil {
		return fmt.Errorf("ошибка при восстановлении базы: %w", err)
	}
	fmt.Println("✅ База успешно восстановлена!")

	fmt.Println("🚀 Шаг 2: Подключаемся к базе и собираем информацию...")

	// Получаем DSN от провайдера контейнера
	baseConnStr := p.cp.GetConnectionString()

	// Передаем DSN в репозиторий для динамического обхода
	if err := p.repo.GetDatabaseInfo(ctx, baseConnStr); err != nil {
		return fmt.Errorf("ошибка при получении инфы о БД: %w", err)
	}

	return nil
}
