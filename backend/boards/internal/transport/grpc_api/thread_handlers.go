package grpc_api

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ThreadGrpcServer) CreateThread(ctx context.Context, req *boardsv1.CreateThreadRequest) (*boardsv1.CreateThreadResponse, error) {
	res, err := s.uc.CreateThread(ctx, actorFrom(ctx), models.CreateThreadInput{
		BoardID:    req.GetBoardId(),
		Title:      req.GetTitle(),
		Body:       req.GetBody(),
		Category:   categoryFromProto(req.GetCategory()),
		Visibility: visibilityFromProto(req.GetVisibility()),
	})
	if err != nil {
		return nil, err
	}

	return &boardsv1.CreateThreadResponse{Thread: threadToProto(res.Thread), FirstPost: postToProto(res.FirstPost)}, nil
}

func (s *ThreadGrpcServer) ListThreads(ctx context.Context, req *boardsv1.ListThreadsRequest) (*boardsv1.ListThreadsResponse, error) {
	res, err := s.uc.ListThreads(ctx, actorFrom(ctx), models.ListThreadsInput{
		BoardID:  req.GetBoardId(),
		Category: categoryFromProto(req.GetCategory()),
		Status:   statusFromProto(req.GetStatus()),
		Search:   req.GetSearch(),
		AuthorID: req.GetAuthorUserId(),
		Order:    orderFromProto(req.GetOrderBy()),
		Page:     pageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, err
	}

	resp := &boardsv1.ListThreadsResponse{Threads: make([]*boardsv1.Thread, 0, len(res.Threads)), Page: pageResponse(res.Page)}
	for _, t := range res.Threads {
		resp.Threads = append(resp.Threads, threadToProto(t))
	}

	return resp, nil
}

func (s *ThreadGrpcServer) GetThread(ctx context.Context, req *boardsv1.GetThreadRequest) (*boardsv1.GetThreadResponse, error) {
	t, err := s.uc.GetThread(ctx, actorFrom(ctx), req.GetThreadId())
	if err != nil {
		return nil, err
	}

	return &boardsv1.GetThreadResponse{Thread: threadToProto(t)}, nil
}

func (s *ThreadGrpcServer) UpdateThread(ctx context.Context, req *boardsv1.UpdateThreadRequest) (*boardsv1.UpdateThreadResponse, error) {
	t, err := s.uc.UpdateThread(ctx, actorFrom(ctx), models.UpdateThreadInput{
		ThreadID:   req.GetThreadId(),
		Title:      req.GetTitle(),
		Category:   categoryFromProto(req.GetCategory()),
		Visibility: visibilityFromProto(req.GetVisibility()),
	})
	if err != nil {
		return nil, err
	}

	return &boardsv1.UpdateThreadResponse{Thread: threadToProto(t)}, nil
}

func (s *ThreadGrpcServer) LockThread(ctx context.Context, req *boardsv1.LockThreadRequest) (*boardsv1.LockThreadResponse, error) {
	t, err := s.uc.LockThread(ctx, actorFrom(ctx), req.GetThreadId(), req.GetReason())
	if err != nil {
		return nil, err
	}

	return &boardsv1.LockThreadResponse{Thread: threadToProto(t)}, nil
}

func (s *ThreadGrpcServer) UnlockThread(ctx context.Context, req *boardsv1.UnlockThreadRequest) (*boardsv1.UnlockThreadResponse, error) {
	t, err := s.uc.UnlockThread(ctx, actorFrom(ctx), req.GetThreadId(), req.GetReason())
	if err != nil {
		return nil, err
	}

	return &boardsv1.UnlockThreadResponse{Thread: threadToProto(t)}, nil
}

func (s *ThreadGrpcServer) PinThread(ctx context.Context, req *boardsv1.PinThreadRequest) (*boardsv1.PinThreadResponse, error) {
	t, err := s.uc.PinThread(ctx, actorFrom(ctx), req.GetThreadId())
	if err != nil {
		return nil, err
	}

	return &boardsv1.PinThreadResponse{Thread: threadToProto(t)}, nil
}

func (s *ThreadGrpcServer) UnpinThread(ctx context.Context, req *boardsv1.UnpinThreadRequest) (*boardsv1.UnpinThreadResponse, error) {
	t, err := s.uc.UnpinThread(ctx, actorFrom(ctx), req.GetThreadId())
	if err != nil {
		return nil, err
	}

	return &boardsv1.UnpinThreadResponse{Thread: threadToProto(t)}, nil
}

func (s *ThreadGrpcServer) DeleteThread(ctx context.Context, req *boardsv1.DeleteThreadRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteThread(ctx, actorFrom(ctx), req.GetThreadId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
