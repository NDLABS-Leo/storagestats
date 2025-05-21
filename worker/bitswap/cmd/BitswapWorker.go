package main

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"storagestats/pkg/task"
	"storagestats/worker/bitswap"
)

func main() {
	worker := bitswap.Worker{}
	process, err := task.NewTaskWorkerProcess(context.Background(), task.Bitswap, worker)
	if err != nil {
		panic(err)
	}

	defer process.Close()

	err = process.Poll(context.Background())
	if err != nil {
		logging.Logger("task-worker").With("protocol", task.Bitswap).Error(err)
	}
}
