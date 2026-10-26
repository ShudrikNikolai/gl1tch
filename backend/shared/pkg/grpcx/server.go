// Package grpcx собирает gRPC-серверы и клиенты проекта с общими интерсепторами:
// request id и данные пользователя через метаданные, deadline по умолчанию, логирование,
// recover и перевод ошибок apperr <-> статус gRPC.
// mTLS подключается через опции (grpc.Creds / WithTransportCredentials).
package grpcx

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr/grpcerr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
)

// NewServer создаёт сервер с интерсептором по умолчанию. Дополнительные опции (creds, лимиты) идут после него.
// Сервисы внутренние и доверяют метаданным gateway (x-request-id, x-user-id, x-roles).
func NewServer(log *slog.Logger, opts ...grpc.ServerOption) *grpc.Server {
	if log == nil {
		log = slog.Default()
	}
	base := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler( // ← добавить
			otelgrpc.WithFilter(func(info *stats.RPCTagInfo) bool {
				return info.FullMethodName != "/grpc.health.v1.Health/Check"
			}),
		)),
		grpc.ChainUnaryInterceptor(unaryServerInterceptor(log)),
		// Совпадает с keepalive клиента (client.go): иначе сервер закрывает соединение за «слишком частые» пинги.
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 20 * time.Second, PermitWithoutStream: true}),
	}

	return grpc.NewServer(append(base, opts...)...)
}

// RegisterHealth регистрирует стандартный grpc.health.v1; возвращённым сервером управляется статус.
func RegisterHealth(s *grpc.Server) *health.Server {
	h := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, h)

	return h
}

func unaryServerInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		ctx = contextFromIncoming(ctx)

		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(ctx, "grpc handler panic", slog.String("method", info.FullMethod),
					slog.Any("panic", r), slog.String("stack", string(debug.Stack())))
				resp, err = nil, grpcerr.ToStatus(fmt.Errorf("panic: %v", r))
			}
		}()

		resp, herr := handler(ctx, req)
		logCall(ctx, log, info.FullMethod, start, herr)

		return resp, grpcerr.ToStatus(herr)
	}
}

// logCall: успешные вызовы на уровне Debug (поток запросов не засоряет лог), ожидаемые ошибки клиента Warn,
// сбои сервиса Error вместе с причиной.
func logCall(ctx context.Context, log *slog.Logger, method string, start time.Time, err error) {
	attrs := []any{slog.String("method", method), slog.Duration("duration", time.Since(start))}
	if err == nil {
		log.DebugContext(ctx, "grpc call", attrs...)
		return
	}

	e := apperr.From(err)
	attrs = append(attrs, slog.String("code", e.Code), slog.Any("error", err))
	switch e.Kind {
	case apperr.KindInternal, apperr.KindUnavailable:
		log.ErrorContext(ctx, "grpc call failed", attrs...)
	default:
		log.WarnContext(ctx, "grpc call rejected", attrs...)
	}
}

func contextFromIncoming(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)

	rid := first(md, reqctx.HeaderRequestID)
	if rid == "" {
		rid = reqctx.NewRequestID()
	}
	ctx = reqctx.WithRequestID(ctx, rid)

	// TODO как дойдут руки переписать
	if uid := first(md, reqctx.HeaderUserID); uid != "" {
		var roles []enum.Role
		if r := first(md, reqctx.HeaderRoles); r != "" {
			parts := strings.Split(r, ",")
			roles = make([]enum.Role, 0, len(parts))
			for _, p := range parts {
				roles = append(roles, enum.Role(p))
			}

		}
		ctx = reqctx.WithUser(ctx, uid, roles)
	}

	return ctx
}

func first(md metadata.MD, key string) string {
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}

	return ""
}
