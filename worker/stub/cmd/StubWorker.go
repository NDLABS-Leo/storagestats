package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"storagestats/pkg/task"
	"storagestats/worker/stub"
)

func main() {
	worker := stub.Worker{}
	process, err := task.NewTaskWorkerProcess(context.Background(), task.Stub, worker)
	if err != nil {
		panic(err)
	}

	defer process.Close()

	err = process.Poll(context.Background())
	if err != nil {
		panic(err)
	}
}
