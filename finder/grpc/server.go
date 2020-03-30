//go:generate protoc -I ../../proto --go_out=plugins=grpc:../../proto ../../proto/finder.proto

package grpc

import (
	"context"
	"fmt"

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
	ids, err := s.app.SearchRecipes(req.Query)
	if err != nil {
		return nil, err
	}

	return &pb.SearchRecipesResponse{Ids: ids}, nil
}

func (s *finderServer) AllRecipeTags(ctx context.Context, req *pb.AllRecipeTagsRequest) (*pb.AllRecipeTagsResponse, error) {
	tags, err := s.app.AllRecipeTags()
	if err != nil {
		return nil, fmt.Errorf("error retrieving recipe tags: %w", err)
	}

	return &pb.AllRecipeTagsResponse{Tags: tags}, nil
}
