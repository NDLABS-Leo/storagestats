package main

import (
	"context"
	logging "github.com/ipfs/go-log/v2"
	"storagestats/pkg/task"
	"storagestats/worker/http"
)

func main() {
	worker := http.Worker{}
	process, err := task.NewTaskWorkerProcess(context.Background(), task.HTTP, worker)
	if err != nil {
		panic(err)
	}

	defer process.Close()

	err = process.Poll(context.Background())
	if err != nil {
		logging.Logger("task-worker").With("protocol", task.HTTP).Error(err)
	}
}
