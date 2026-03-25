package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Docker   Docker   `yaml:"docker"`
	Database Database `yaml:"database"`
}

type Docker struct {
	Image         string `yaml:"image" env-default:"postgres:15-alpine"`
	ContainerName string `yaml:"container_name" env-default:"restore-assert-pg"`
	MemoryLimit   string `yaml:"memory_limit" env-default:"1GB"`
	CPULimit      string `yaml:"cpu_limit" env-default:"1.0"`
	AutoRemove    bool   `yaml:"auto_remove" env-default:"true"`
}

type Database struct {
	DBName     string            `yaml:"db_name" env-default:"restore_test"`
	User       string            `yaml:"user" env-default:"postgres"`
	Password   string            `yaml:"password" env-default:"password"`
	Extensions []string          `yaml:"extensions"`
	Roles      []string          `yaml:"roles"`
	ConfigFile string            `yaml:"config_file"` // ПУТЬ ДО ФАЙЛА (НЕОБЯЗАТЕЛЬНО)
	Settings   map[string]string `yaml:"settings"`    // INLINE-НАСТРОЙКИ (FALLBACK)
}

func LoadConfig(path string) (*Config, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("config file is missing from the path: %s", path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg, nil
}
