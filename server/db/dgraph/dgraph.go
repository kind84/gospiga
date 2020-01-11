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
		for i := 1; i < 4; i++ {
			err = nil
			time.Sleep(time.Second)
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

	// load schema
	// check if server is ready to go
	res, err := http.Get("http://alpha:8080/health")
	for err != nil || res.StatusCode != http.StatusOK {
		time.Sleep(time.Second)
		res, err = http.Get("http://alpha:8080/health")
	}
	log.Debug("dgraph server ready")

	op := loadRecipeSchema()

	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, err
	}

	return &DB{dgraph}, nil
}
