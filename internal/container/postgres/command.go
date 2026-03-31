// Формирование строк psql/pg_restore
package postgres

import (
	"fmt"

	"github.com/Robcenster/restore-assert/internal/config"
)

func buildRestoreCommand(dbCfg config.Database, rCfg config.Restore, bType BackupType, containerPath string) ([]string, error) {
	var cmd []string

	switch bType {
	case TypeCustom, TypeTar, TypeDirectory:
		// === Бинарные форматы (pg_restore) ===
		cmd = []string{
			"pg_restore",
			"-U", dbCfg.User,
			"-d", dbCfg.DBName,
			"-v", // Для более детального вывода логов (Так же потом перенести в config)
		}

		if bType == TypeCustom {
			cmd = append(cmd, "-Fc")
		} else if bType == TypeDirectory {
			cmd = append(cmd, "-Fd")
		} else if bType == TypeTar {
			cmd = append(cmd, "-Ft")
		}

		// Применяем настройки из твоего конфига
		if rCfg.NoOwner {
			cmd = append(cmd, "--no-owner")
		}
		if rCfg.NoPrivileges {
			cmd = append(cmd, "--no-acl") // в pg_restore это называется --no-acl
		}
		if rCfg.ParallelJobs > 1 && (bType == TypeCustom || bType == TypeDirectory) {
			cmd = append(cmd, "-j", fmt.Sprintf("%d", rCfg.ParallelJobs))
		}
		if rCfg.SingleTransaction {
			// ВАЖНО: pg_restore не поддерживает --single-transaction вместе с -j > 1
			if rCfg.ParallelJobs <= 1 {
				cmd = append(cmd, "--single-transaction")
			}
		}

		cmd = append(cmd, containerPath)

	case TypePlain:
		// === Текстовый формат (psql) ===
		cmd = []string{
			"psql",
			"-U", dbCfg.User,
			"-d", dbCfg.DBName,
		}

		if rCfg.OnErrorStop {
			cmd = append(cmd, "-v", "ON_ERROR_STOP=1")
		}
		if rCfg.SingleTransaction {
			cmd = append(cmd, "-1") // Флаг единой транзакции в psql
		}

		cmd = append(cmd, "-f", containerPath)

	case TypeDumpAll:
		// === Дамп кластера (psql) ===
		// Игнорируем User и DBName из конфига, так как dumpall восстанавливается под postgres
		cmd = []string{
			"psql",
			"-U", "postgres",
			"-d", "postgres",
		}

		if rCfg.OnErrorStop {
			cmd = append(cmd, "-v", "ON_ERROR_STOP=1")
		}
		// Для pg_dumpall --single-transaction часто не работает, так как внутри есть CREATE DATABASE

		cmd = append(cmd, "-f", containerPath)

	default:
		return nil, fmt.Errorf("unsupported backup type: %s", bType)
	}

	return cmd, nil
}
