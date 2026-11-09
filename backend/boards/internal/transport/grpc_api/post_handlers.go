package grpc_api

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *PostGrpcServer) CreatePost(ctx context.Context, req *boardsv1.CreatePostRequest) (*boardsv1.CreatePostResponse, error) {
	p, err := s.uc.CreatePost(ctx, actorFrom(ctx), models.CreatePostInput{
		ThreadID: req.GetThreadId(),
		ParentID: req.GetParentPostId(),
		Body:     req.GetBody(),
	})
	if err != nil {
		return nil, err
	}

	return &boardsv1.CreatePostResponse{Post: postToProto(p)}, nil
}

func (s *PostGrpcServer) ListPosts(ctx context.Context, req *boardsv1.ListPostsRequest) (*boardsv1.ListPostsResponse, error) {
	res, err := s.uc.ListPosts(ctx, actorFrom(ctx), models.ListPostsInput{
		ThreadID: req.GetThreadId(),
		Page:     pageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, err
	}

	resp := &boardsv1.ListPostsResponse{Posts: make([]*boardsv1.Post, 0, len(res.Posts)), Page: pageResponse(res.Page)}
	for _, p := range res.Posts {
		resp.Posts = append(resp.Posts, postToProto(p))
	}

	return resp, nil
}

func (s *PostGrpcServer) GetPost(ctx context.Context, req *boardsv1.GetPostRequest) (*boardsv1.GetPostResponse, error) {
	p, err := s.uc.GetPost(ctx, actorFrom(ctx), req.GetPostId())
	if err != nil {
		return nil, err
	}

	return &boardsv1.GetPostResponse{Post: postToProto(p)}, nil
}

func (s *PostGrpcServer) UpdatePost(ctx context.Context, req *boardsv1.UpdatePostRequest) (*boardsv1.UpdatePostResponse, error) {
	p, err := s.uc.UpdatePost(ctx, actorFrom(ctx), req.GetPostId(), req.GetBody())
	if err != nil {
		return nil, err
	}

	return &boardsv1.UpdatePostResponse{Post: postToProto(p)}, nil
}

func (s *PostGrpcServer) DeletePost(ctx context.Context, req *boardsv1.DeletePostRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeletePost(ctx, actorFrom(ctx), req.GetPostId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PostGrpcServer) HasUserPosted(ctx context.Context, req *boardsv1.HasUserPostedRequest) (*boardsv1.HasUserPostedResponse, error) {
	posted, err := s.uc.HasUserPosted(ctx, req.GetThreadId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &boardsv1.HasUserPostedResponse{Posted: posted}, nil
}
