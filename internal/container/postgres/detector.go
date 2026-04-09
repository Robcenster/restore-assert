package postgres

import (
	"io"
	"os"
	"strings"
)

type BackupType string

// TODO: Почистить комменты, заменить константы на const, убрать русский
const (
	TypeCustom     BackupType = "custom"
	TypeDirectory  BackupType = "directory"
	TypeTar        BackupType = "tar"
	TypePlain      BackupType = "plain" // В формате SQL
	TypeDumpAll    BackupType = "dumpall"
	TypeCompressed BackupType = "compressed" // Для .gz или .zst
	TypeUnknown    BackupType = "unknown"
)

// DetectBackupType определяет формат файла дампа PostgreSQL
func detectBackupType(dumpPath string) (BackupType, error) {
	info, err := os.Stat(dumpPath)
	if err != nil {
		return TypeUnknown, err
	}

	// 1. Проверка на формат Directory (-Fd)
	if info.IsDir() {
		return TypeDirectory, nil
	}

	f, err := os.Open(dumpPath)
	if err != nil {
		return TypeUnknown, err
	}
	defer f.Close()

	// 2. Читаем первые 512 байт (безопасный стандарт для сниффинга типов)
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return TypeUnknown, err // Ловим реальные ошибки чтения
	}

	if n == 0 {
		return TypeUnknown, nil // Файл абсолютно пуст
	}

	// Работаем только с фактически прочитанными данными
	chunk := buf[:n]

	// 3. Проверяем бинарные форматы по магическим байтам

	// Custom формат (-Fc)
	if len(chunk) >= 5 && string(chunk[:5]) == "PGDMP" {
		return TypeCustom, nil
	}

	// Проверка на GZIP (\x1f\x8b) - часто используется для сжатия текстовых дампов
	if len(chunk) >= 2 && chunk[0] == 0x1f && chunk[1] == 0x8b {
		return TypeCompressed, nil
	}

	// TAR формат (-Ft). Сигнатура "ustar" обычно находится на 257-м байте
	if len(chunk) >= 262 && string(chunk[257:262]) == "ustar" {
		return TypeTar, nil
	}

	// 4. Проверяем текстовые форматы (-Fp и pg_dumpall)
	content := string(chunk)

	if strings.Contains(content, "-- PostgreSQL database cluster dump") {
		return TypeDumpAll, nil
	}

	if strings.Contains(content, "-- PostgreSQL database dump") {
		return TypePlain, nil
	}

	// 5. Расширенная эвристика для SQL
	// Если строгий заголовок случайно удалили, но это точно дамп от pg_dump:
	// Он почти всегда начинается с настройки параметров сессии.
	if strings.Contains(content, "SET statement_timeout") ||
		strings.Contains(content, "SET client_encoding") {
		return TypePlain, nil
	}

	return TypeUnknown, nil
}
