package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/database/postgres"
)

func main() {
	ctx := context.Background()

	// 1. Загружаем всё из YAML через cleanenv
	cfg, _ := config.LoadConfig("config/template/postgres.yaml")

	// 2. Стартуем Docker, просто передав нужные части конфига
	provider, err := postgres.StartContainer(ctx, cfg.Docker, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close(ctx)

	fmt.Println("🚀 База запущена с лимитом:", cfg.Docker.MemoryLimit)
}
