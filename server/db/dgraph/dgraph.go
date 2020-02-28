package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
		return nil, fmt.Errorf("failed to connect to dgraph: %w", err)
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
		return nil, fmt.Errorf("dgraph server not ready: %w", err)
	}

	drop := api.Operation{DropAll: true}
	err = dgraph.Alter(ctx, &drop)
	if err != nil {
		return nil, fmt.Errorf("failed to flush dgraph schema: %w", err)
	}

	op := loadRecipeSchema()

	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, fmt.Errorf("failed to load dgraph schema: %w", err)
	}

	// load graphql schema
	file, err := os.Open("/gql/schema.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to open graphql.schema file: %w", err)
	}
	defer file.Close()

	res, err = http.Post("http://alpha:8080/admin/schema", "x-www-form-urlencoded", file)
	if err != nil {
		return nil, fmt.Errorf("failed to load graphql schema into dgraph: %w", err)
	}
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

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
