package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository реализует интерфейс repository.DBRepository для Postgres
type Repository struct {
	pool *pgxpool.Pool
}

// New создает новый пул соединений к Postgres.
// Обрати внимание: функция возвращает конкретную структуру *Repository,
// но благодаря методам она автоматически удовлетворяет интерфейсу!
func New(ctx context.Context, connString string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Обязательно закрываем пул, если пинг не прошел
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Repository{pool: pool}, nil
}

// Close закрывает пул соединений к базе
func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// GetDatabaseInfo делает тестовые запросы и выводит структуру БД.
// Твой отличный метод без изменений, просто адаптирован под структуру Repository!
func (r *Repository) GetDatabaseInfo(ctx context.Context) error {
	var dbName string
	err := r.pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		return fmt.Errorf("failed to get db name: %w", err)
	}

	fmt.Printf("\n=== Информация о базе данных ===\n")
	fmt.Printf("Подключено к БД: %s\n", dbName)

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name;
	`
	rows, err := r.pool.Query(ctx, query)
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
