package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"

	"google.golang.org/grpc"

	"gospiga/pkg/log"
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
		log.Warnf("failed to connect to dgraph, retrying..")
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
			log.Debugf("dgraph server ready")
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			break
		}
		time.Sleep(5 * 100 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("dgraph server not ready: %w", err)
	}

	time.Sleep(5 * time.Second)
	// load graphql schema
	log.Debugf("loading graphql schema..")
	file, err := os.Open("/gql/schema.graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to open graphql.schema file: %w", err)
	}
	defer file.Close()

	res, err = http.Post("http://alpha:8080/admin/schema", "x-www-form-urlencoded", file)
	if err != nil {
		return nil, fmt.Errorf("failed to load graphql schema into dgraph: %w", err)
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	time.Sleep(5 * time.Second)
	// load dgraph schema
	log.Debugf("loading dgraph schema..")
	op := loadRecipeSchema()
	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, fmt.Errorf("failed to load dgraph schema: %w", err)
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
