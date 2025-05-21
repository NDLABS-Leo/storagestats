package main

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"storagestats/pkg/task"
	"storagestats/worker/graphsync"
)

func main() {
	worker := graphsync.Worker{}
	process, err := task.NewTaskWorkerProcess(context.Background(), task.GraphSync, worker)
	if err != nil {
		panic(err)
	}

	defer process.Close()

	err = process.Poll(context.Background())
	if err != nil {
		logging.Logger("task-worker").With("protocol", task.GraphSync).Error(err)
	}
}
