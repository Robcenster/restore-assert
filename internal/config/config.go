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
)

type Config struct {
	Engine   EngineType `yaml:"engine" env:"DB_ENGINE" env-default:"postgres"`
	Docker   Docker     `yaml:"docker"`
	Database Database   `yaml:"database"`
	Restore  Restore    `yaml:"restore"`
}

type Docker struct {
	Image string `yaml:"image" env:"DOCKER_IMAGE" env-default:"postgres:17-alpine"`

	// TODO: Генерировать UUID?
	ContainerName string `yaml:"container_name" env:"CONTAINER_NAME" env-default:"restore-assert-pg"`
	MemoryLimit   string `yaml:"memory_limit" env-default:"1GB"`
	AutoRemove    bool   `yaml:"auto_remove" env-default:"true"`
	WaitTimeout   string `yaml:"wait_timeout" env-default:"60s"`
}

type Database struct {
	DBName     string            `yaml:"db_name" env-default:"restore_test"`
	User       string            `yaml:"user" env-default:"postgres"`
	Password   string            `yaml:"password" env-default:"postgres"`
	Extensions []string          `yaml:"extensions"`
	Roles      []string          `yaml:"roles"`
	ConfigFile string            `yaml:"config_file"` // ПУТЬ ДО ФАЙЛА (НЕОБЯЗАТЕЛЬНО)
	Settings   map[string]string `yaml:"settings"`    // INLINE-НАСТРОЙКИ (FALLBACK)
}

type Restore struct {
	Analyze           bool `yaml:"analyze" env-default:"true"`
	FullRestoreLogs   bool `yaml:"full_restore_logs"`
	OnErrorStop       bool `yaml:"on_error_stop"`
	SingleTransaction bool `yaml:"single_transaction"`
	ParallelJobs      int  `yaml:"parallel_jobs" env-default:"1"`
	NoOwner           bool `yaml:"no_owner"`
	NoPrivileges      bool `yaml:"no_privileges"`
}

func Load(path string) (*Config, error) {
	// Если путь не передан, попробуем взять дефолтный файл
	if path == "" {
		path = "restore-config.yaml"
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
