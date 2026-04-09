package main

import (
	"fmt"

	"github.com/Robcenster/restore-assert/internal/cli"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("So, some panic get in a way, check this out:", r)
		}
	}()

	cli.Execute()
}
