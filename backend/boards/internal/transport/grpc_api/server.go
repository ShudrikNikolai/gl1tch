// Package grpc_api gRPC-сервисы форума. Здесь только перевод protobuf в модели usecase и обратно:
// бизнес-правила и права проверяет usecase, а ошибки (apperr) превращает в статус интерцептор grpcx.
package grpc_api

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
	"google.golang.org/grpc"
)

type BoardGrpcServer struct {
	boardsv1.UnimplementedBoardServiceServer
	uc usecase.BoardUseCase
}

type ThreadGrpcServer struct {
	boardsv1.UnimplementedThreadServiceServer
	uc usecase.ThreadUseCase
}

type PostGrpcServer struct {
	boardsv1.UnimplementedPostServiceServer
	uc usecase.PostUseCase
}

// Register подключает все три сервиса форума к gRPC-серверу.
func Register(s grpc.ServiceRegistrar, boards usecase.BoardUseCase, threads usecase.ThreadUseCase, posts usecase.PostUseCase) {
	boardsv1.RegisterBoardServiceServer(s, &BoardGrpcServer{uc: boards})
	boardsv1.RegisterThreadServiceServer(s, &ThreadGrpcServer{uc: threads})
	boardsv1.RegisterPostServiceServer(s, &PostGrpcServer{uc: posts})
}
