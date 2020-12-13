package grpc

import (
	"context"

	"gospiga/pkg/types"
	pb "gospiga/proto"
)

type stub struct {
	client pb.FinderClient
}

// NewStub with grpc client.
func NewStub(client *pb.FinderClient) *stub {
	return &stub{*client}
}

func (s *stub) AllRecipeTags(ctx context.Context) ([]string, error) {
	req := &pb.AllRecipeTagsRequest{}

	res, err := s.client.AllRecipeTags(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Tags, nil
}

func (s *stub) RecipesFT(ctx context.Context, args *types.SearchRecipesArgs) ([]string, error) {
	var (
		first *pb.RecipesFTRequest_First
		after *pb.RecipesFTRequest_After
		query *pb.RecipesFTRequest_Query
	)

	if args.First != nil {
		first = &pb.RecipesFTRequest_First{First: int32(*args.First)}
	}
	if args.After != nil {
		after = &pb.RecipesFTRequest_After{After: *args.After}
	}
	if args.Query != nil {
		query = &pb.RecipesFTRequest_Query{Query: *args.Query}
	}

	req := pb.RecipesFTRequest{
		OptionalFirst: first,
		OptionalAfter: after,
		Tags:          args.Tags,
		Ingredients:   args.Ingredients,
		OptionalQuery: query,
	}

	res, err := s.client.RecipesFT(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res.Ids, nil
}
