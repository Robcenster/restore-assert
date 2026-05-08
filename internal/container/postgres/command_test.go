package postgres

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Robcenster/restore-assert/internal/config"
)

func TestBuildRestoreCommand(t *testing.T) {
	validDB := config.Database{
		User:   "testuser",
		DBName: "testdb",
	}

	tests := map[string]struct {
		dbCfg         config.Database
		rCfg          config.Restore
		bType         BackupType
		containerPath string
		wantCmd       []string
		wantErr       bool
	}{
		"Error: Empty container path": {
			dbCfg:         validDB,
			rCfg:          config.Restore{},
			bType:         TypeCustom,
			containerPath: "",
			wantCmd:       nil,
			wantErr:       true,
		},
		"Error: Missing DB credentials for Custom": {
			dbCfg:         config.Database{},
			rCfg:          config.Restore{},
			bType:         TypeCustom,
			containerPath: "/backup.dump",
			wantCmd:       nil,
			wantErr:       true,
		},
		"Error: TAR format with parallel jobs": {
			dbCfg:         validDB,
			rCfg:          config.Restore{ParallelJobs: 4},
			bType:         TypeTar,
			containerPath: "/backup.tar",
			wantCmd:       nil,
			wantErr:       true,
		},
		"Error: Conflict Parallel and SingleTransaction": {
			dbCfg:         validDB,
			rCfg:          config.Restore{ParallelJobs: 4, SingleTransaction: true},
			bType:         TypeCustom,
			containerPath: "/backup.dump",
			wantCmd:       nil,
			wantErr:       true,
		},
		"Error: Unsupported backup type": {
			dbCfg:         validDB,
			rCfg:          config.Restore{},
			bType:         BackupType("UNKNOWN_FORMAT"),
			containerPath: "/backup.dump",
			wantCmd:       nil,
			wantErr:       true,
		},

		"Success: Directory format with all flags (Parallel)": {
			dbCfg: validDB,
			rCfg: config.Restore{
				ShowRestoreLogs: true, //  Should result in -v
				NoOwner:         true, //  Should result in --no-owner
				NoPrivileges:    true, //  Should result in --no-acl
				OnErrorStop:     true, //  Should result in --exit-on-error
				ParallelJobs:    4,    //  Should result in -j 4
			},
			bType:         TypeDirectory,
			containerPath: "/backup_dir",
			wantCmd: []string{
				"pg_restore", "-U", "testuser", "-d", "testdb",
				"-v", "-Fd", "--no-owner", "--no-acl", "--exit-on-error", "-j", "4", "/backup_dir",
			},
			wantErr: false,
		},
		"Success: Custom format with Single Transaction": {
			dbCfg: validDB,
			rCfg: config.Restore{
				SingleTransaction: true,
				ParallelJobs:      1, // <= 1 allows transaction
			},
			bType:         TypeCustom,
			containerPath: "/backup.dump",
			wantCmd: []string{
				"pg_restore", "-U", "testuser", "-d", "testdb",
				"-Fc", "--single-transaction", "/backup.dump",
			},
			wantErr: false,
		},
		"Success: Tar format basic": {
			dbCfg:         validDB,
			rCfg:          config.Restore{ParallelJobs: 0},
			bType:         TypeTar,
			containerPath: "/backup.tar",
			wantCmd: []string{
				"pg_restore", "-U", "testuser", "-d", "testdb",
				"-Ft", "/backup.tar",
			},
			wantErr: false,
		},

		"Success: Plain format (SQL) with flags": {
			dbCfg: validDB,
			rCfg: config.Restore{
				ShowRestoreLogs:   true, // Should result in -a
				OnErrorStop:       true, // Should result in -v ON_ERROR_STOP=1
				SingleTransaction: true, // Should result in -1
				NoOwner:           true, // Should be ignored
			},
			bType:         TypePlain,
			containerPath: "/backup.sql",
			wantCmd: []string{
				"psql", "-U", "testuser", "-d", "testdb",
				"-a", "-v", "ON_ERROR_STOP=1", "-1", "-f", "/backup.sql",
			},
			wantErr: false,
		},
		"Success: DumpAll format (forces postgres user, ignores single-transaction)": {
			dbCfg: config.Database{
				User:   "random_user", // Передаем кастомного пользователя
				DBName: "random_db",
			},
			rCfg: config.Restore{
				SingleTransaction: true,
				OnErrorStop:       true,
			},
			bType:         TypeDumpAll,
			containerPath: "/cluster.sql",
			wantCmd: []string{
				"psql", "-U", "random_user", "-d", "postgres", // Исправлено: теперь ждем "random_user"
				"-v", "ON_ERROR_STOP=1", "-f", "/cluster.sql",
			},
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotCmd, err := buildRestoreCommand(tc.dbCfg, tc.rCfg, tc.bType, tc.containerPath)

			if (err != nil) != tc.wantErr {
				t.Fatalf("BuildRestoreCommand() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr {
				return
			}

			if !reflect.DeepEqual(gotCmd, tc.wantCmd) {
				t.Errorf("\nExpected command:\n%v\nGot command:\n%v",
					strings.Join(tc.wantCmd, " "),
					strings.Join(gotCmd, " "))
			}
		})
	}
}
