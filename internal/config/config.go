package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type EngineType string

const (
	EnginePostgres EngineType = "postgres"
	EngineMSSQL    EngineType = "mssql"

	defaultConfigPath = "restore-config.yaml"
)

type Config struct {
	Engine   EngineType `yaml:"engine" env:"DB_ENGINE" env-default:"postgres"`
	Docker   Docker     `yaml:"docker"`
	Database Database   `yaml:"database"`
	Restore  Restore    `yaml:"restore"`
	Asserts  Asserts    `yaml:"asserts"`
}

type Docker struct {
	Image         string `yaml:"image" env:"DOCKER_IMAGE" env-default:"postgres:17-alpine"`
	ContainerName string `yaml:"container_name" env:"CONTAINER_NAME" env-default:"restore-assert-pg"`
	MemoryLimit   string `yaml:"memory_limit" env-default:"1GB"`
	AutoRemove    bool   `yaml:"auto_remove" env-default:"true"`
	WaitTimeout   string `yaml:"wait_timeout" env-default:"60s"`
}

type Database struct {
	DBName     string            `yaml:"db_name" env-default:"postgres"`
	User       string            `yaml:"user" env-default:"postgres"`
	Password   string            `yaml:"password" env-default:"postgres"`
	Extensions []string          `yaml:"extensions"`
	Roles      []string          `yaml:"roles"`
	Settings   map[string]string `yaml:"settings"`
}

type Restore struct {
	Analyze           bool `yaml:"analyze"`
	OnErrorStop       bool `yaml:"on_error_stop"`
	SingleTransaction bool `yaml:"single_transaction"`
	ParallelJobs      int  `yaml:"parallel_jobs" env-default:"1"`
	NoOwner           bool `yaml:"no_owner"`
	NoPrivileges      bool `yaml:"no_privileges"`
	ModifyTemplate    bool `yaml:"modify_template"`
	ShowRestoreLogs   bool `yaml:"show_restore_logs"`
	ShowDatabaseInfo  bool `yaml:"show_db_info"`
	ShowSuccessTests  bool `yaml:"show_success_tests"`
}

type Asserts struct {
	Existence  Existence   `yaml:"existence"`
	Tables     []Table     `yaml:"tables"`
	Privileges []Privilege `yaml:"privileges"`
	Queries    []Query     `yaml:"queries"`
}

type Existence struct {
	Extensions []string `yaml:"extensions"`
	Roles      []string `yaml:"roles"`
	Schemas    []string `yaml:"schemas"`
}

type Table struct {
	Name    string   `yaml:"name"`
	Metrics []Metric `yaml:"metrics"`
}

type Metric struct {
	Type       string  `yaml:"type"`        // row_count, table_size, sequence_health, freshness, null_ratio
	Condition  string  `yaml:"condition"`   // eq, gt, lt
	Expected   any     `yaml:"expected"`    // interface{}, чтобы парсить 100, "50MB", "1h"
	Column     string  `yaml:"column"`      // Для freshness, null_ratio
	MaxAge     string  `yaml:"max_age"`     // Для freshness
	MaxPercent float64 `yaml:"max_percent"` // Для null_ratio
	Severity   string  `yaml:"severity"`    // error, warn
}

type Privilege struct {
	Role      string   `yaml:"role"`
	Table     string   `yaml:"table"`
	Allowed   []string `yaml:"allowed"`
	Forbidden []string `yaml:"forbidden"`
}

type Query struct {
	Name      string `yaml:"name"`
	Query     string `yaml:"query"`
	Condition string `yaml:"condition"`
	Expected  any    `yaml:"expected"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = defaultConfigPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at: %s", path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
