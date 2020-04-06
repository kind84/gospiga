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

func (s *stub) AllRecipeTags(ctx context.Context) ([]string, error) {
	req := &pb.AllRecipeTagsRequest{}

	res, err := s.client.AllRecipeTags(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Tags, nil
}
