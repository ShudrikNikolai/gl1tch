package grpcx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/enum"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/reqctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

const method = "/test.Echo/Do"

type handlerFunc func(ctx context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error)

type echoService interface{}

var echoDesc = grpc.ServiceDesc{
	ServiceName: "test.Echo",
	HandlerType: (*echoService)(nil),
	Methods: []grpc.MethodDesc{{
		MethodName: "Do",
		Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
			in := new(wrapperspb.StringValue)
			if err := dec(in); err != nil {
				return nil, err
			}
			call := func(ctx context.Context, req any) (any, error) {
				return srv.(handlerFunc)(ctx, req.(*wrapperspb.StringValue))
			}
			if interceptor == nil {
				return call(ctx, in)
			}

			return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: method}, call)
		},
	}},
}

type env struct {
	conn   *grpc.ClientConn
	health interface {
		SetServingStatus(string, grpc_health_v1.HealthCheckResponse_ServingStatus)
	}
}

func start(t *testing.T, h handlerFunc, clientTimeout time.Duration, dialOpts ...grpc.DialOption) env {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := NewServer(discard)
	srv.RegisterService(&echoDesc, h)
	hs := RegisterHealth(srv)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	opts := append([]grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, dialOpts...)
	conn, err := Dial("passthrough:///bufnet", clientTimeout, opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return env{conn: conn, health: hs}
}

func call(ctx context.Context, conn *grpc.ClientConn, text string) (string, error) {
	out := new(wrapperspb.StringValue)
	if err := conn.Invoke(ctx, method, wrapperspb.String(text), out); err != nil {
		return "", err
	}

	return out.Value, nil
}

func TestMetadataPropagation(t *testing.T) {
	var gotRID, gotUID string
	var gotRoles []enum.Role
	e := start(t, func(ctx context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		gotRID, gotUID, gotRoles = reqctx.RequestID(ctx), reqctx.UserID(ctx), reqctx.Roles(ctx)
		return in, nil
	}, 0)

	ctx := reqctx.WithUser(reqctx.WithRequestID(context.Background(), "req-42"), "user-7", []enum.Role{"user", "moder"})
	if _, err := call(ctx, e.conn, "x"); err != nil {
		t.Fatal(err)
	}
	if gotRID != "req-42" || gotUID != "user-7" || len(gotRoles) != 2 || gotRoles[1] != "moder" {
		t.Fatalf("rid=%q uid=%q roles=%v", gotRID, gotUID, gotRoles)
	}
}

func TestRequestIDGeneratedWhenMissing(t *testing.T) {
	var clientSide, serverSide string
	e := start(t, func(ctx context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		serverSide = reqctx.RequestID(ctx)
		return in, nil
	}, 0)
	if _, err := call(context.Background(), e.conn, "x"); err != nil {
		t.Fatal(err)
	}
	clientSide = serverSide
	if len(clientSide) != 32 {
		t.Fatalf("expected a generated 32-char request id, got %q", serverSide)
	}
}

func TestAppErrorRoundTrip(t *testing.T) {
	e := start(t, func(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return nil, apperr.AlreadyExists("NICKNAME_TAKEN", "nickname is taken")
	}, 0)
	_, err := call(context.Background(), e.conn, "x")
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind != apperr.KindAlreadyExists || ae.Code != "NICKNAME_TAKEN" || ae.Message != "nickname is taken" {
		t.Fatalf("got %#v", err)
	}
}

func TestInternalErrorsDoNotLeak(t *testing.T) {
	e := start(t, func(context.Context, *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return nil, errors.New("pq: password authentication failed for user agora at 10.0.0.5")
	}, 0)
	_, err := call(context.Background(), e.conn, "x")
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind != apperr.KindInternal || ae.Code != apperr.CodeInternal {
		t.Fatalf("got %#v", err)
	}
	if strings.Contains(err.Error(), "10.0.0.5") || strings.Contains(ae.Message, "password") {
		t.Fatalf("internal details leaked to the frontend: %v", err)
	}
}

func TestPanicIsRecoveredAndServerSurvives(t *testing.T) {
	n := 0
	e := start(t, func(_ context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		n++
		if n == 1 {
			panic("boom")
		}
		return in, nil
	}, 0)

	_, err := call(context.Background(), e.conn, "first")
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind != apperr.KindInternal {
		t.Fatalf("panic must become an internal error, got %#v", err)
	}
	if got, err := call(context.Background(), e.conn, "second"); err != nil || got != "second" {
		t.Fatalf("server must keep serving after a panic: %q %v", got, err)
	}
}

func TestDefaultDeadlineApplied(t *testing.T) {
	seen := make(chan bool, 1) // обработчик продолжает работать после ответа клиенту: общаемся через канал
	e := start(t, func(ctx context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		_, ok := ctx.Deadline()
		seen <- ok
		<-ctx.Done() // зависаем до deadline
		return in, ctx.Err()
	}, 100*time.Millisecond)

	begin := time.Now()
	_, err := call(context.Background(), e.conn, "x")
	if err == nil || time.Since(begin) > 2*time.Second {
		t.Fatalf("call must fail by the default deadline, err=%v after %v", err, time.Since(begin))
	}
	select {
	case ok := <-seen:
		if !ok {
			t.Fatal("deadline must reach the server")
		}
	case <-time.After(time.Second):
		t.Fatal("handler was not called")
	}
	if k := apperr.KindOf(err); k != apperr.KindUnavailable {
		t.Fatalf("deadline exceeded must map to unavailable, got kind %d (%v)", k, err)
	}
}

func TestExistingDeadlineIsKept(t *testing.T) {
	var remaining time.Duration
	e := start(t, func(ctx context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		dl, _ := ctx.Deadline()
		remaining = time.Until(dl)
		return in, nil
	}, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := call(ctx, e.conn, "x"); err != nil {
		t.Fatal(err)
	}
	if remaining < 5*time.Second {
		t.Fatalf("explicit deadline must not be shortened, remaining %v", remaining)
	}
}

func TestHealthService(t *testing.T) {
	e := start(t, func(_ context.Context, in *wrapperspb.StringValue) (*wrapperspb.StringValue, error) { return in, nil }, 0)
	e.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	resp, err := grpc_health_v1.NewHealthClient(e.conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil || resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("%v %v", resp, err)
	}
}
