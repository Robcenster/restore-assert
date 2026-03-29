package repopostgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type storage struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*storage, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &storage{Pool: pool}, nil
}

func (s *storage) Close() {
	s.Pool.Close()
}

// GetDatabaseInfo делает тестовые запросы и выводит структуру БД
func (s *storage) GetDatabaseInfo(ctx context.Context) error {
	// 1. Узнаем текущую базу данных
	var dbName string
	err := s.Pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		return fmt.Errorf("failed to get db name: %w", err)
	}

	fmt.Printf("\n=== Информация о базе данных ===\n")
	fmt.Printf("Подключено к БД: %s\n", dbName)

	// 2. Получаем список таблиц в схеме public
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name;
	`
	rows, err := s.Pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to fetch tables: %w", err)
	}
	defer rows.Close()

	fmt.Println("Таблицы (схема 'public'):")
	count := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		fmt.Printf(" - %s\n", tableName)
		count++
	}

	if count == 0 {
		fmt.Println(" (Таблиц не найдено)")
	}
	fmt.Printf("================================\n\n")

	return nil
}

