package verifier

import (
	"context"
	"fmt"
	"strconv"
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

// QoteIdentifier correctly processes names such as "public.movies" -> "public"."movies"
func (v *Verifier) quoteIdentifier(target string) string {
	parts := strings.Split(target, ".")
	for i, part := range parts {
		parts[i] = fmt.Sprintf(`"%s"`, part)
	}
	return strings.Join(parts, ".")
}

// EscapeLiteral to prevent single quotes from "breaking" usernames
func escapeLiteral(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

func (v *Verifier) RunAssert(ctx context.Context, task AssertTask) (bool, error) {
	var query string
	var expected = task.Expected
	var condition = task.Condition

	quotedTarget := v.quoteIdentifier(task.Target)

	switch task.Type {
	case "existence_ext":
		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = '%s')::text", escapeLiteral(task.Target))
		condition, expected = "eq", "true"

	case "existence_role":
		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = '%s')::text", escapeLiteral(task.Target))
		condition, expected = "eq", "true"

	case "existence_schema":
		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s')::text", escapeLiteral(task.Target))
		condition, expected = "eq", "true"

	case "row_count":
		query = fmt.Sprintf("SELECT count(*)::text FROM %s", quotedTarget)

	case "table_size":
		query = fmt.Sprintf("SELECT pg_total_relation_size('%s'::regclass)::text", escapeLiteral(quotedTarget))

		bytesQuery := fmt.Sprintf("SELECT pg_size_bytes('%v')::text", escapeLiteral(fmt.Sprintf("%v", task.Expected)))
		expectedStr, err := v.source.ExecuteQuery(ctx, bytesQuery)
		if err != nil {
			return false, fmt.Errorf("invalid table_size format '%v': %v", task.Expected, err)
		}
		expected = expectedStr

	case "null_ratio":
		query = fmt.Sprintf(
			`SELECT COALESCE(SUM(CASE WHEN "%s" IS NULL THEN 1 ELSE 0 END)::float / NULLIF(COUNT(*), 0), 0)::text FROM %s`,
			task.Column, quotedTarget,
		)
		condition, expected = "lt", task.MaxPercent

	case "privilege":
		query = fmt.Sprintf("SELECT has_table_privilege('%s', '%s', '%s')::text",
			escapeLiteral(task.Role), escapeLiteral(quotedTarget), escapeLiteral(task.Action))
		condition, expected = "eq", fmt.Sprintf("%t", task.IsAllowed)

	case "query":
		query = task.Query

	case "freshness":
		query = fmt.Sprintf(`SELECT EXTRACT(EPOCH FROM max("%s"))::text FROM %s`, task.Column, quotedTarget)
		actualRaw, err := v.source.ExecuteQuery(ctx, query)
		if err != nil {
			return false, err
		}
		if actualRaw == "" || actualRaw == "null" {
			return false, fmt.Errorf("freshness check failed: table is empty or column has only NULLs")
		}

		epochStr := strings.Split(actualRaw, ".")[0]
		epochInt, err := strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			return false, fmt.Errorf("cannot parse epoch time '%s': %w", actualRaw, err)
		}

		lastDate := time.Unix(epochInt, 0)
		maxAge, _ := time.ParseDuration(task.MaxAge)

		if time.Since(lastDate) > maxAge {
			return false, fmt.Errorf("data is too old: last entry %v", lastDate.Format(time.RFC822))
		}
		return true, nil

	default:
		return false, fmt.Errorf("unknown task type: %s", task.Type)
	}

	actualRaw, err := v.source.ExecuteQuery(ctx, query)
	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}

	return Compare(actualRaw, expected, condition)
}
