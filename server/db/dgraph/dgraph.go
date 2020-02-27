package dgraph

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
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
	var d *grpc.ClientConn
	var err error
	for i := 0; i < 10; i++ {
		err = nil
		d, err = grpc.Dial("alpha:9080", grpc.WithInsecure())
		if err == nil {
			break
		}
		log.Warn("failed to connect to dgraph, retrying..")
		time.Sleep(5 * 100 * time.Millisecond)
	}
	if err != nil {
		log.Error("failed to connect to dgraph")
		return nil, err
	}

	dgraph := dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)

	// check if server is ready to go
	var res *http.Response
	for i := 0; i < 20; i++ {
		err = nil
		res, err = http.Get("http://alpha:8080/health")
		if err == nil && res.StatusCode == http.StatusOK {
			log.Debug("dgraph server ready")
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
			break
		}
		time.Sleep(5 * 100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}

	op := loadRecipeSchema()

	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, err
	}

	return &DB{dgraph}, nil
}

func (db *DB) count(ctx context.Context, onType string) (int, error) {
	vars := map[string]string{"$type": onType}
	q := `
		query Count($type: string){
			countType(func: type($type)) {
				totalCount: count(uid)
			}
		}
	`

	resp, err := db.Dgraph.NewReadOnlyTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return 0, err
	}

	var root struct {
		CountType []struct {
			TotalCount int `json:"totalCount"`
		} `json:"countType"`
	}
	err = json.Unmarshal(resp.Json, &root)
	if err != nil {
		return 0, err
	}
	return root.CountType[0].TotalCount, nil
}
