package postgres

import (
	"io"
	"os"
	"strings"
)

type BackupType string

const (
	TypeCustom     BackupType = "custom"
	TypeDirectory  BackupType = "directory"
	TypeTar        BackupType = "tar"
	TypePlain      BackupType = "plain" // format SQL
	TypeDumpAll    BackupType = "dumpall"
	TypeCompressed BackupType = "compressed" // for .gz or .zst
	TypeUnknown    BackupType = "unknown"
)

// DetectBackupType определяет формат файла дампа PostgreSQL
func detectBackupType(dumpPath string) (BackupType, error) {
	info, err := os.Stat(dumpPath)
	if err != nil {
		return TypeUnknown, err
	}

	if info.IsDir() {
		return TypeDirectory, nil
	}

	f, err := os.Open(dumpPath)
	if err != nil {
		return TypeUnknown, err
	}
	defer f.Close()

	// Reading the first 512 bytes
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return TypeUnknown, err
	}

	if n == 0 {
		return TypeUnknown, nil
	}

	chunk := buf[:n]

	if len(chunk) >= 5 && string(chunk[:5]) == "PGDMP" {
		return TypeCustom, nil
	}

	if len(chunk) >= 2 && chunk[0] == 0x1f && chunk[1] == 0x8b {
		return TypeCompressed, nil
	}

	if len(chunk) >= 262 && string(chunk[257:262]) == "ustar" {
		return TypeTar, nil
	}

	content := string(chunk)

	if strings.Contains(content, "-- PostgreSQL database cluster dump") {
		return TypeDumpAll, nil
	}

	if strings.Contains(content, "-- PostgreSQL database dump") {
		return TypePlain, nil
	}

	if strings.Contains(content, "SET statement_timeout") ||
		strings.Contains(content, "SET client_encoding") {
		return TypePlain, nil
	}

	return TypeUnknown, nil
}
