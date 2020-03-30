package grpc

import (
	"context"
	pb "github.com/kind84/gospiga/proto"
)

type stub struct {
	client pb.FinderClient
}

// NewStub with grpc client.
func NewStub(client *pb.FinderClient) *stub {
	return &stub{*client}
}

// SearchRecipes over grpc and return recipe IDS matching the query.
func (s *stub) SearchRecipes(ctx context.Context, query string) ([]string, error) {
	req := &pb.SearchRecipesRequest{
		Query: query,
	}

	res, err := s.client.SearchRecipes(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Ids, nil
}

func (s *stub) AllRecipeTags(ctx context.Context) ([]string, error) {
	req := &pb.AllRecipeTagsRequest{}

	res, err := s.client.AllRecipeTags(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Tags, nil
}
