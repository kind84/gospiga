// go:generate protoc -I ../../grpc ../../grpc/finder.proto --go_out=plugins=grpc:../../grpc

package grpc

import (
	"context"

	pb "github.com/kind84/gospiga/proto"
)

type finderServer struct {
	app App
}

func NewFinderServer(app App) *finderServer {
	return &finderServer{app}
}

// SearchRecipes implements grpc server interface method.
func (s *finderServer) SearchRecipes(ctx context.Context, req *pb.SearchRecipesRequest) (*pb.SearchRecipesResponse, error) {
	ids, err := s.app.SearchRecipes(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	var res pb.SearchRecipesResponse
	res.Ids = ids
	return &res, nil
}
