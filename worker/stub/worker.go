package stub

import (
	"math/rand"
	"storagestats/pkg/task"
	"time"
)

type Worker struct{}

func (e Worker) DoWork(_ task.Task) (*task.RetrievalResult, error) {
	//nolint: gosec
	return task.NewSuccessfulRetrievalResult(
		time.Duration(rand.Int31()),
		int64(rand.Int31()),
		time.Duration(rand.Int31())), nil
}
