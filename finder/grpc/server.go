package grpc

// ** TO USE OPTIONAL FIELDS: //go:generate protoc --experimental_allow_proto3_optional=true -I ../../proto --go_out=plugins=grpc:../../proto ../../proto/finder.proto
//go:generate protoc -I ../../proto --go_out=plugins=grpc:../../proto ../../proto/finder.proto

import (
	"context"
	"fmt"

	"gospiga/pkg/types"
	pb "gospiga/proto"
)

type finderServer struct {
	app App
}

func NewFinderServer(app App) *finderServer {
	return &finderServer{app}
}

func (s *finderServer) AllRecipeTags(ctx context.Context, req *pb.AllRecipeTagsRequest) (*pb.AllRecipeTagsResponse, error) {
	tags, err := s.app.AllRecipeTags()
	if err != nil {
		return nil, fmt.Errorf("error retrieving recipe tags: %w", err)
	}

	return &pb.AllRecipeTagsResponse{Tags: tags}, nil
}

func (s *finderServer) RecipesFT(ctx context.Context, req *pb.RecipesFTRequest) (*pb.RecipesFTResponse, error) {
	var (
		first *int
		after *string
		query *string
	)

	if rf, ok := req.OptionalFirst.(*pb.RecipesFTRequest_First); ok {
		first = &[]int{int(rf.First)}[0]
	}
	if ra, ok := req.OptionalAfter.(*pb.RecipesFTRequest_After); ok {
		after = &ra.After
	}
	if rq, ok := req.OptionalQuery.(*pb.RecipesFTRequest_Query); ok {
		query = &rq.Query
	}

	ids, err := s.app.SearchIDs(&types.SearchRecipesArgs{
		First:       first,
		After:       after,
		Tags:        req.Tags,
		Ingredients: req.Ingredients,
		Query:       query,
	})

	return &pb.RecipesFTResponse{Ids: ids}, err
}
