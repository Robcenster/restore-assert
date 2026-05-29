package postgres

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectBackupType(t *testing.T) {
	tempDir := t.TempDir()

	// Empty file
	emptyFile := filepath.Join(tempDir, "empty.dump")
	require.NoError(t, os.WriteFile(emptyFile, []byte{}, 0644))

	// File containing junk data
	garbageFile := filepath.Join(tempDir, "garbage.txt")
	require.NoError(t, os.WriteFile(garbageFile, []byte("just some random text without any pg signatures"), 0644))

	// Custom (-Fc)
	customFile := filepath.Join(tempDir, "custom.dump")
	require.NoError(t, os.WriteFile(customFile, []byte("PGDMP\x01\x02\x03\x04\x05"), 0644))

	// GZIP/Compressed (.gz)
	gzipFile := filepath.Join(tempDir, "compressed.gz")
	require.NoError(t, os.WriteFile(gzipFile, []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}, 0644))

	// TAR (-Ft)
	tarFile := filepath.Join(tempDir, "backup.tar")
	tarBuf := make([]byte, 512)
	copy(tarBuf[257:], "ustar")
	require.NoError(t, os.WriteFile(tarFile, tarBuf, 0644))

	// DumpAll (pg_dumpall)
	dumpAllFile := filepath.Join(tempDir, "dumpall.sql")
	require.NoError(t, os.WriteFile(dumpAllFile, []byte("-- PostgreSQL database cluster dump\nCREATE ROLE test;"), 0644))

	// Plain (pg_dump strict)
	plainFile := filepath.Join(tempDir, "plain.sql")
	require.NoError(t, os.WriteFile(plainFile, []byte("-- PostgreSQL database dump\nSELECT 1;"), 0644))

	// Plain
	heuristicFile := filepath.Join(tempDir, "heuristic.sql")
	require.NoError(t, os.WriteFile(heuristicFile, []byte("SET statement_timeout = 0;\nSET client_encoding = 'UTF8';"), 0644))

	tests := []struct {
		name          string
		filePath      string
		expectedType  BackupType
		expectedError bool
	}{
		{
			name:          "Directory format",
			filePath:      tempDir, // Передаем саму временную папку
			expectedType:  TypeDirectory,
			expectedError: false,
		},
		{
			name:          "Custom format (PGDMP)",
			filePath:      customFile,
			expectedType:  TypeCustom,
			expectedError: false,
		},
		{
			name:          "Compressed format (GZIP)",
			filePath:      gzipFile,
			expectedType:  TypeCompressed,
			expectedError: false,
		},
		{
			name:          "TAR format (ustar signature)",
			filePath:      tarFile,
			expectedType:  TypeTar,
			expectedError: false,
		},
		{
			name:          "DumpAll text format",
			filePath:      dumpAllFile,
			expectedType:  TypeDumpAll,
			expectedError: false,
		},
		{
			name:          "Plain text format (strict)",
			filePath:      plainFile,
			expectedType:  TypePlain,
			expectedError: false,
		},
		{
			name:          "Plain text format (heuristic settings)",
			filePath:      heuristicFile,
			expectedType:  TypePlain,
			expectedError: false,
		},
		{
			name:          "Empty file",
			filePath:      emptyFile,
			expectedType:  TypeUnknown,
			expectedError: false,
		},
		{
			name:          "Garbage data",
			filePath:      garbageFile,
			expectedType:  TypeUnknown,
			expectedError: false,
		},
		{
			name:          "Non-existent file",
			filePath:      filepath.Join(tempDir, "does_not_exist.sql"),
			expectedType:  TypeUnknown,
			expectedError: true, // Ожидаем ошибку от os.Stat или os.Open
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bType, err := detectBackupType(tt.filePath)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, bType)
			}
		})
	}
}