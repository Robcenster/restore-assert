package main

import (
	"github.com/Robcenster/restore-assert/internal/cli"
)

func main() {
	// Просто вызываем Execute из пакета cli.
	// Если что-то пойдет не так, приложение само завершится с кодом 1.
	cli.Execute()
}
