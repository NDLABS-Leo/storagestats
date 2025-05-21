package graphsync

import (
	"context"
	"github.com/ipfs/go-cid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"storagestats/pkg/net"
	"storagestats/pkg/task"
)

type Worker struct{}

func (e Worker) DoWork(tsk task.Task) (*task.RetrievalResult, error) {
	ctx := context.Background()
	host, err := net.InitHost(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init host")
	}

	client := net.NewGraphsyncClient(host, tsk.Timeout)
	addrInfo, err := tsk.Provider.GetPeerAddr()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get peer addr")
	}
	contentCID := cid.MustParse(tsk.Content.CID)
	//nolint:wrapcheck
	return client.Retrieve(ctx, addrInfo, contentCID)
}
