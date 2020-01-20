package dgraph

import (
	"context"
	"net/http"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DB struct {
	*dgo.Dgraph
}

func NewDB(ctx context.Context) (*DB, error) {
	d, err := grpc.Dial("alpha:9080", grpc.WithInsecure())
	if err != nil {
		log.Warn("failed to connect to dgraph, retrying..")
		for i := 0; i < 10; i++ {
			err = nil
			time.Sleep(5 * 100 * time.Millisecond)
			d, err = grpc.Dial("alpha:9080", grpc.WithInsecure())
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Error("failed to connect to dgraph")
			return nil, err
		}
	}

	dgraph := dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)

	// check if server is ready to go
	res, err := http.Get("http://alpha:8080/health")
	for i := 0; i < 20; i++ {
		res, err = http.Get("http://alpha:8080/health")
		if err == nil {
			if res.StatusCode != http.StatusOK {
				break
			}
		}
		time.Sleep(5 * 100 * time.Millisecond)
	}
	log.Debug("dgraph server ready")

	op := loadRecipeSchema()

	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, err
	}

	return &DB{dgraph}, nil
}
