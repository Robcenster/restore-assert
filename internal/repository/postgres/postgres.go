package postgres

import (
	"context"
	"fmt"
	"net/url"

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

func (r *Repository) GetDatabaseInfo(ctx context.Context, baseConnStr string) error {
	fmt.Println("\n======================================")
	fmt.Println("   ПОЛНАЯ ИНФОРМАЦИЯ О КЛАСТЕРЕ       ")
	fmt.Println("======================================")

	// 1. Выводим роли (используем текущий r.pool)
	fmt.Println("👥 Роли и пользователи:")
	roleQuery := "SELECT rolname, rolsuper FROM pg_roles WHERE rolname NOT LIKE 'pg_%'"
	rows, _ := r.pool.Query(ctx, roleQuery)
	for rows.Next() {
		var name string
		var isSuper bool
		rows.Scan(&name, &isSuper)
		fmt.Printf(" - %s (Superuser: %v)\n", name, isSuper)
	}
	rows.Close()

	// 2. Получаем список всех БД
	var dbNames []string
	dbQuery := "SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'rdsadmin'"
	dbRows, err := r.pool.Query(ctx, dbQuery)
	if err != nil {
		return err
	}
	for dbRows.Next() {
		var name string
		dbRows.Scan(&name)
		dbNames = append(dbNames, name)
	}
	dbRows.Close()

	// 3. Инспектируем каждую БД
	for _, name := range dbNames {
		fmt.Printf("\n--- 🗄️ База: %s ---\n", name)

		// Формируем DSN для конкретной базы
		u, _ := url.Parse(baseConnStr)
		u.Path = "/" + name

		// Создаем временный пул для этой базы
		tempPool, err := pgxpool.New(ctx, u.String())
		if err != nil {
			fmt.Printf(" [!] Ошибка подключения: %v\n", err)
			continue
		}

		// Выводим таблицы
		tableQuery := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
		tRows, err := tempPool.Query(ctx, tableQuery)
		if err != nil {
			fmt.Printf(" [!] Ошибка запроса таблиц: %v\n", err)
			tempPool.Close()
			continue
		}

		count := 0
		for tRows.Next() {
			var tName string
			tRows.Scan(&tName)
			fmt.Printf("  📑 Таблица: %s\n", tName)
			count++
		}
		tRows.Close()

		if count == 0 {
			fmt.Println("  (таблиц нет)")
		}

		tempPool.Close() // Обязательно закрываем временный пул
	}

	return nil
}
