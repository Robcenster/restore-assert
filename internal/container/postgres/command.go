// Building commands to container
package postgres

import (
	"fmt"
	"strconv"

	"github.com/Robcenster/restore-assert/internal/config"
)

// TODO: Clean comments up, use consts, change to english
func buildRestoreCommand(dbCfg config.Database, rCfg config.Restore, bType BackupType, containerPath string) ([]string, error) {
	if containerPath == "" {
		return nil, fmt.Errorf("container path cannot be empty")
	}

	// Если это не дамп всего кластера, требуем данные для подключения из конфига
	if bType != TypeDumpAll {
		if dbCfg.User == "" || dbCfg.DBName == "" {
			return nil, fmt.Errorf("restore error: database user and name must be set")
		}
	}

	var cmd []string

	switch bType {
	case TypeCustom, TypeTar, TypeDirectory:
		// === Бинарные форматы (pg_restore) ===
		cmd = []string{
			"pg_restore",
			"-U", dbCfg.User,
			"-d", dbCfg.DBName,
		}

		if rCfg.ShowRestoreLogs {
			cmd = append(cmd, "-v")
		}

		// Определение формата
		switch bType {
		case TypeCustom:
			cmd = append(cmd, "-Fc")
		case TypeDirectory:
			cmd = append(cmd, "-Fd")
		case TypeTar:
			cmd = append(cmd, "-Ft")
		}

		// Игнорирование владельцев и привилегий
		if rCfg.NoOwner {
			cmd = append(cmd, "--no-owner")
		}
		if rCfg.NoPrivileges {
			cmd = append(cmd, "--no-acl")
		}

		// Останавливаться при первой ошибке (Исправлено: для pg_restore это --exit-on-error)
		if rCfg.OnErrorStop {
			cmd = append(cmd, "--exit-on-error")
		}

		// === Сложная логика потоков и транзакций ===

		// 1. Проверка для TAR формата (pg_restore не умеет восстанавливать TAR в несколько потоков)
		if bType == TypeTar && rCfg.ParallelJobs > 1 {
			return nil, fmt.Errorf("pg_restore does not support parallel_jobs with TAR format")
		}

		// 2. Проверка конфликта: параллельность vs транзакция
		if rCfg.ParallelJobs > 1 && rCfg.SingleTransaction {
			return nil, fmt.Errorf("cannot use parallel_jobs (>1) and single_transaction simultaneously")
		}

		if rCfg.ParallelJobs > 1 {
			cmd = append(cmd, "-j", strconv.Itoa(rCfg.ParallelJobs))
		} else if rCfg.SingleTransaction {
			cmd = append(cmd, "--single-transaction")
		}

		cmd = append(cmd, containerPath)

	case TypePlain:
		// === Текстовый формат (psql) ===
		cmd = []string{
			"psql",
			"-U", dbCfg.User,
			"-d", dbCfg.DBName,
		}

		if rCfg.ShowRestoreLogs {
			cmd = append(cmd, "-a") // Аналог -v в psql: echo-all (выводит все запросы)
		}
		if rCfg.OnErrorStop {
			cmd = append(cmd, "-v", "ON_ERROR_STOP=1")
		}
		if rCfg.SingleTransaction {
			cmd = append(cmd, "-1")
		}

		cmd = append(cmd, "-f", containerPath)

	case TypeDumpAll:
		// === Дамп кластера (psql) ===
		user := dbCfg.User
		if user == "" {
			user = "postgres"
		}

		cmd = []string{
			"psql",
			"-U", user,
			"-d", "postgres", // 'postgres' тут только как точка входа
		}

		if rCfg.ShowRestoreLogs {
			cmd = append(cmd, "-a")
		}
		if rCfg.OnErrorStop {
			cmd = append(cmd, "-v", "ON_ERROR_STOP=1")
		}

		cmd = append(cmd, "-f", containerPath)

	default:
		return nil, fmt.Errorf("unsupported backup type: %s", bType)
	}
	return cmd, nil
}
