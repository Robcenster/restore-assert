// Формирование строк psql/pg_restore
package postgres

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Robcenster/restore-assert/internal/config"
)

// TODO: Почистить комменты, заменить константы на const, убрать русский
func buildRestoreCommand(dbCfg config.Database, rCfg config.Restore, bType BackupType, containerPath string) ([]string, error) {
	if containerPath == "" {
		return nil, errors.New("container path cannot be empty")
	}

	// Если это не дамп всего кластера, требуем данные для подключения из конфига
	if bType != TypeDumpAll {
		if dbCfg.User == "" || dbCfg.DBName == "" {
			return nil, errors.New("restore error: database user and name must be set")
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

		// Включаем подробный режим (можно добавить дважды "-v", "-v" для экстра-деталей, но одного обычно хватает)
		if rCfg.FullRestoreLogs {
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
			return nil, errors.New("pg_restore does not support parallel_jobs with TAR format")
		}

		// 2. Проверка конфликта: параллельность vs транзакция
		if rCfg.ParallelJobs > 1 && rCfg.SingleTransaction {
			return nil, errors.New("cannot use parallel_jobs (>1) and single_transaction simultaneously")
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

		if rCfg.FullRestoreLogs {
			cmd = append(cmd, "-a") // Аналог -v в psql: echo-all (выводит все запросы)
		}
		if rCfg.OnErrorStop {
			cmd = append(cmd, "-v", "ON_ERROR_STOP=1")
		}
		if rCfg.SingleTransaction {
			cmd = append(cmd, "-1")
		}

		// Важное примечание: psql не поддерживает флаги --no-owner или --no-acl.
		// Текстовый дамп — это просто набор SQL-запросов. Игнорировать владельцев здесь
		// можно только подавляя ошибки (OnErrorStop=false) или изменяя сам файл до импорта.

		cmd = append(cmd, "-f", containerPath)

	case TypeDumpAll:
		// === Дамп кластера (psql) ===

		// Ответ на сомнение: "Меня смущает имя и пароль бд, разве их нет в самом файле"
		// Да, в файле есть `CREATE DATABASE` и `\connect`.
		// НО! Утилита psql — это клиент. Чтобы начать читать файл и отправлять команды серверу,
		// ей необходимо установить первичное подключение к какой-либо существующей базе данных.
		// База "postgres" и пользователь "postgres" существуют всегда, поэтому это стандартная точка входа.
		user := dbCfg.User
		if user == "" {
			user = "postgres"
		}

		cmd = []string{
			"psql",
			"-U", user,
			"-d", "postgres", // 'postgres' тут только как точка входа, это ок
		}

		if rCfg.FullRestoreLogs {
			cmd = append(cmd, "-a")
		}
		if rCfg.OnErrorStop {
			cmd = append(cmd, "-v", "ON_ERROR_STOP=1")
		}

		// SingleTransaction здесь намеренно опущен. pg_dumpall генерирует команды
		// CREATE DATABASE / CREATE ROLE, которые технически невозможно выполнить внутри блока BEGIN/COMMIT.

		cmd = append(cmd, "-f", containerPath)

	default:
		return nil, fmt.Errorf("unsupported backup type: %s", bType)
	}
	fmt.Println(cmd)
	return cmd, nil
}

