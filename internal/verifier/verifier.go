package verifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/repository"
)

type Verifier struct {
	source repository.DBRepository
}

func NewVerifier(source repository.DBRepository) *Verifier {
	return &Verifier{source: source}
}

// AssertTask — это универсальная, "расплющенная" задача для выполнения одной проверки
type AssertTask struct {
	Name       string
	Type       string // existence, row_count, table_size, freshness, null_ratio, privilege, query
	Target     string // Имя таблицы, роли, схемы или расширения
	Query      string // Тело запроса (для кастомных запросов)
	Condition  string
	Expected   any
	Column     string  // Для freshness и null_ratio
	MaxAge     string  // Для freshness
	MaxPercent float64 // Для null_ratio
	Role       string  // Для privilege
	Action     string  // Для privilege
	IsAllowed  bool    // Для privilege
}

// CreateTasks преобразует вложенный конфиг YAML в плоский список задач для Pipeline
func (v *Verifier) CreateTasks(asserts config.Asserts) []AssertTask {
	var tasks []AssertTask

	// 1. Быстрые проверки существования
	for _, ext := range asserts.Existence.Extensions {
		tasks = append(tasks, AssertTask{Name: "Extension: " + ext, Type: "existence_ext", Target: ext})
	}
	for _, role := range asserts.Existence.Roles {
		tasks = append(tasks, AssertTask{Name: "Role: " + role, Type: "existence_role", Target: role})
	}
	for _, schema := range asserts.Existence.Schemas {
		tasks = append(tasks, AssertTask{Name: "Schema: " + schema, Type: "existence_schema", Target: schema})
	}

	// 2. Метрики таблиц
	for _, t := range asserts.Tables {
		for _, m := range t.Metrics {
			tasks = append(tasks, AssertTask{
				Name:       fmt.Sprintf("Table %s [%s]", t.Name, m.Type),
				Type:       m.Type,
				Target:     t.Name,
				Condition:  m.Condition,
				Expected:   m.Expected,
				Column:     m.Column,
				MaxAge:     m.MaxAge,
				MaxPercent: m.MaxPercent,
			})
		}
	}

	// 3. Права доступа (Privileges)
	for _, p := range asserts.Privileges {
		for _, allowed := range p.Allowed {
			tasks = append(tasks, AssertTask{
				Name:      fmt.Sprintf("Privilege: %s can %s on %s", p.Role, allowed, p.Table),
				Type:      "privilege",
				Target:    p.Table,
				Role:      p.Role,
				Action:    allowed,
				IsAllowed: true,
			})
		}
		for _, forbidden := range p.Forbidden {
			tasks = append(tasks, AssertTask{
				Name:      fmt.Sprintf("Privilege: %s CANNOT %s on %s", p.Role, forbidden, p.Table),
				Type:      "privilege",
				Target:    p.Table,
				Role:      p.Role,
				Action:    forbidden,
				IsAllowed: false,
			})
		}
	}

	// 4. Кастомные запросы
	for _, q := range asserts.Queries {
		tasks = append(tasks, AssertTask{
			Name:      "Query: " + q.Name,
			Type:      "query",
			Query:     q.Query,
			Condition: q.Condition,
			Expected:  q.Expected,
		})
	}

	return tasks
}

// quoteIdentifier корректно обрабатывает имена типа "public.movies" -> "public"."movies"
func (v *Verifier) quoteIdentifier(target string) string {
	parts := strings.Split(target, ".")
	for i, part := range parts {
		parts[i] = fmt.Sprintf(`"%s"`, part)
	}
	return strings.Join(parts, ".")
}

func (v *Verifier) RunAssert(ctx context.Context, task AssertTask) (bool, error) {
	var query string
	var expected = task.Expected
	var condition = task.Condition

	// Подготавливаем безопасно закавыченное имя таблицы/объекта
	// "movies" -> "movies", "public.movies" -> "public"."movies"
	quotedTarget := v.quoteIdentifier(task.Target)

	switch task.Type {
	case "existence_ext":
		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = '%s')::text", task.Target)
		condition, expected = "eq", "true"

	case "existence_role":
		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = '%s')::text", task.Target)
		condition, expected = "eq", "true"

	case "existence_schema":
		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s')::text", task.Target)
		condition, expected = "eq", "true"

	case "row_count":
		// Используем уже закавыченный таргет без лишних кавычек в fmt
		query = fmt.Sprintf("SELECT count(*)::text FROM %s", quotedTarget)

	case "table_size":
		// regclass — самый надежный способ сослаться на таблицу в Postgres
		query = fmt.Sprintf("SELECT pg_total_relation_size('%s'::regclass)::text", quotedTarget)

		bytesQuery := fmt.Sprintf("SELECT pg_size_bytes('%v')::text", task.Expected)
		expectedStr, err := v.source.ExecuteQuery(ctx, bytesQuery)
		if err != nil {
			return false, fmt.Errorf("invalid table_size format '%v': %v", task.Expected, err)
		}
		expected = expectedStr

	case "null_ratio":
		// Используем FILTER для чистоты кода (Postgres 9.4+)
		query = fmt.Sprintf(
			`SELECT (COUNT(*) FILTER (WHERE "%s" IS NULL)::float / NULLIF(COUNT(*), 0))::text FROM %s`,
			task.Column, quotedTarget,
		)
		condition, expected = "lt", task.MaxPercent

	case "privilege":
		// has_table_privilege принимает имя таблицы как строку, она сама разберется со схемами
		query = fmt.Sprintf("SELECT has_table_privilege('%s', '%s', '%s')::text", task.Role, task.Target, task.Action)
		condition, expected = "eq", fmt.Sprintf("%t", task.IsAllowed)

	case "query":
		query = task.Query

	case "freshness":
		query = fmt.Sprintf(`SELECT max("%s")::text FROM %s`, task.Column, quotedTarget)
		actualRaw, err := v.source.ExecuteQuery(ctx, query)
		if err != nil {
			return false, err
		}
		// Если в таблице нет данных, max() вернет NULL
		if actualRaw == "" || actualRaw == "null" {
			return false, fmt.Errorf("freshness check failed: table is empty or column has only NULLs")
		}

		lastDate, err := time.Parse(time.RFC3339, actualRaw)
		if err != nil {
			// Пробуем распарсить стандартный формат Postgres, если RFC3339 не прошел
			lastDate, err = time.Parse("2006-01-02 15:04:05", strings.Split(actualRaw, ".")[0])
			if err != nil {
				return false, fmt.Errorf("cannot parse time '%s': %w", actualRaw, err)
			}
		}
		maxAge, _ := time.ParseDuration(task.MaxAge)
		if time.Since(lastDate) > maxAge {
			return false, fmt.Errorf("data is too old: last entry %v", lastDate.Format(time.RFC822))
		}
		return true, nil

	case "sequence_health":
		// Используем pg_get_serial_sequence, чтобы найти сиквенс, привязанный к колонке id
		// Это универсальный способ для Postgres
		query = fmt.Sprintf(`
        SELECT (
            last_value < (SELECT MAX(id) FROM %s)
        )::text 
        FROM pg_sequences 
        WHERE schemaname = quote_ident(COALESCE(NULLIF(split_part('%s', '.', 1), '%s'), 'public'))
        AND tablename = quote_ident(COALESCE(NULLIF(split_part('%s', '.', 2), ''), '%s'))`,
			quotedTarget, task.Target, task.Target, task.Target, task.Target,
		)
		condition, expected = "eq", "false"

	default:
		return false, fmt.Errorf("unknown task type: %s", task.Type)
	}

	actualRaw, err := v.source.ExecuteQuery(ctx, query)
	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}

	return Compare(actualRaw, expected, condition)
}
