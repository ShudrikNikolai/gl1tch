package grpcx

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr/grpcerr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
)

// DefaultTimeout применяется, если у вызова нет своего deadline.
const DefaultTimeout = 2 * time.Second

// roundRobin: клиентская балансировка по всем адресам, которые вернул resolver (для dns:///сервис:9090 это все реплики).
const roundRobin = `{"loadBalancingConfig":[{"round_robin":{}}]}`

// Dial создаёт клиента. target вида "dns:///posts:9090". Если у вызова нет deadline, ставится defaultTimeout
// (0 означает DefaultTimeout). В исходящие метаданные добавляются request id и пользователь из context, а ошибки
// статусов gRPC приходят вызывающему как *apperr.Error. По умолчанию соединение без TLS;
// TODO: для mTLS нужно будет добавить grpc.WithTransportCredentials(...) в opts (позже переданные опции перекрывают умолчания).
func Dial(target string, defaultTimeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if defaultTimeout <= 0 {
		defaultTimeout = DefaultTimeout
	}
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(roundRobin),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 30 * time.Second, Timeout: 10 * time.Second, PermitWithoutStream: true}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(unaryClientInterceptor(defaultTimeout)),
	}

	return grpc.NewClient(target, append(base, opts...)...)
}

func unaryClientInterceptor(defaultTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
			defer cancel()
		}

		return grpcerr.FromStatus(invoker(outgoing(ctx), method, req, reply, cc, opts...))
	}
}

func outgoing(ctx context.Context) context.Context {
	rid := reqctx.RequestID(ctx)
	if rid == "" {
		rid = reqctx.NewRequestID()
	}
	pairs := []string{reqctx.HeaderRequestID, rid}
	if uid := reqctx.UserID(ctx); uid != "" {
		pairs = append(pairs, reqctx.HeaderUserID, uid)
		roles := reqctx.Roles(ctx)
		parts := make([]string, len(roles))
		for i, r := range roles {
			parts[i] = string(r)
		}
		pairs = append(pairs, reqctx.HeaderRoles, strings.Join(parts, ","))
	}

	return metadata.AppendToOutgoingContext(ctx, pairs...)
}
