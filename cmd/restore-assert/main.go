package main

import (
	"context"
	"log"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/provider"
	"github.com/Robcenster/restore-assert/internal/provider/postgres"
	"github.com/Robcenster/restore-assert/internal/repository/repopostgres"
)

func main() {
	ctx := context.Background()
	cPath := "config/template/postgres.yaml"
	dPath := "./backups/dump-admin_db-202603251145.tar"

	// 1. Загрузка конфига
	log.Println("[Шаг 1] Загрузка конфигурации...")
	cfg, err := config.LoadConfig(cPath)
	if err != nil {
		// log.Fatalf печатает ошибку и делает os.Exit(1) - программа останавливается!
		log.Fatalf("❌ Ошибка загрузки конфига: %v", err)
	}

	// 2. Анализ бэкапа
	log.Printf("[Шаг 2] Анализ файла дампа: %s\n", dPath)
	bType, err := postgres.DetectBackupType(dPath)
	if err != nil {
		log.Fatalf("❌ Ошибка анализа бэкапа: %v", err)
	}
	log.Printf("✅ Формат бэкапа определен как: %s\n", bType)

	// 3. Запуск Docker-контейнера
	log.Println("[Шаг 3] Поднятие PostgreSQL в Docker (это может занять время)...")
	var dbProvider provider.Provider
	dbProvider, err = postgres.StartContainer(ctx, cfg.Docker, cfg.Database)
	if err != nil {
		log.Fatalf("❌ Не удалось запустить контейнер: %v", err)
	}
	// Гарантируем удаление контейнера при завершении
	defer func() {
		log.Println("[Очистка] Удаление контейнера...")
		dbProvider.Close(ctx)
	}()
	log.Println("✅ Контейнер успешно запущен и готов к работе!")

	// 4. Восстановление
	log.Println("[Шаг 4] Начало процесса восстановления...")
	// ИСПРАВЛЕНИЕ: Теперь мы ловим ошибку из метода Restore
	err = dbProvider.Restore(ctx, dPath, bType, *cfg)
	if err != nil {
		log.Fatalf("❌ Ошибка при выполнении Restore: %v", err)
	}
	log.Println("✅ База данных успешно восстановлена из дампа!")

	// 5. Подключение через pgx и вывод информации
	log.Println("[Шаг 5] Подключение к восстановленной БД для проверок...")
	connStr := dbProvider.ConnectionString()

	// ИСПРАВЛЕНИЕ: Используем := для создания новой переменной db
	db, err := repopostgres.New(ctx, connStr)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к базе: %v", err)
	}
	// Теперь db.Close() будет работать, так как мы вернули экспортируемый *Storage
	defer db.Close()

	log.Println("✅ Подключение установлено. Считываем метаданные...")

	// Вызываем наш новый метод
	if err := db.GetDatabaseInfo(ctx); err != nil {
		log.Fatalf("❌ Ошибка при получении инфы о БД: %v", err)
	}

	log.Println("🎉 Пайплайн успешно завершен!")
}
