package grpc_api

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
)

// Ошибки отдаются как есть: интерцептор grpcx превращает apperr в gRPC-статус с ErrorInfo.

func (s *BoardGrpcServer) CreateBoard(ctx context.Context, req *boardsv1.CreateBoardRequest) (*boardsv1.CreateBoardResponse, error) {
	b, err := s.uc.CreateBoard(ctx, actorFrom(ctx), models.CreateBoardInput{
		Slug:        req.GetSlug(),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Visibility:  visibilityFromProto(req.GetVisibility()),
		SortOrder:   int(req.GetSortOrder()),
	})
	if err != nil {
		return nil, err
	}

	return &boardsv1.CreateBoardResponse{Board: boardToProto(b)}, nil
}

func (s *BoardGrpcServer) ListBoards(ctx context.Context, _ *boardsv1.ListBoardsRequest) (*boardsv1.ListBoardsResponse, error) {
	items, err := s.uc.ListBoards(ctx, actorFrom(ctx))
	if err != nil {
		return nil, err
	}

	resp := &boardsv1.ListBoardsResponse{Boards: make([]*boardsv1.Board, 0, len(items))}
	for _, b := range items {
		resp.Boards = append(resp.Boards, boardToProto(b))
	}

	return resp, nil
}

func (s *BoardGrpcServer) GetBoard(ctx context.Context, req *boardsv1.GetBoardRequest) (*boardsv1.GetBoardResponse, error) {
	b, err := s.uc.GetBoard(ctx, actorFrom(ctx), req.GetBoardId())
	if err != nil {
		return nil, err
	}

	return &boardsv1.GetBoardResponse{Board: boardToProto(b)}, nil
}
