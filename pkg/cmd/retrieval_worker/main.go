package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"storagestats/pkg/process"
)

func main() {
	processManager, err := process.NewProcessManager()
	if err != nil {
		panic(err)
	}

	processManager.Run(context.Background())
}
