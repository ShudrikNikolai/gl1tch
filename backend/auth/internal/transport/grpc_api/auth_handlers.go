package grpc_api

import (
	"context"
	"net/http"
	"slices"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	authv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// Envoy добавляет этот заголовок всем REST-запросам. Прямые gRPC-клиенты
	// и получают refresh-токен в теле ответа.
	headerTokenTransport = "x-token-transport"
	transportCookie      = "cookie"
	headerCookie         = "cookie"
	headerSetCookie      = "set-cookie"
)

// Ошибки отдаются как есть: grpcx-интерцептор превращает apperr в gRPC-статус с ErrorInfo.

func (ags *AuthGrpcServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	res, err := ags.authcase.Register(ctx, RegisterConverter(req))
	if err != nil {
		return nil, err
	}

	resp := RegisterResponseConverter(res)
	if err := ags.refreshToCookie(ctx, resp.GetTokens()); err != nil {
		return nil, err
	}

	return resp, nil
}

func (ags *AuthGrpcServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	res, err := ags.authcase.Login(ctx, LoginConverter(req))
	if err != nil {
		return nil, err
	}

	resp := LoginResponseConverter(res)
	if err := ags.refreshToCookie(ctx, resp.GetTokens()); err != nil {
		return nil, err
	}

	return resp, nil
}

func (ags *AuthGrpcServer) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	res, err := ags.authcase.Refresh(ctx, ags.refreshTokenFrom(ctx, req.GetRefreshToken()))
	if err != nil {
		return nil, err
	}

	resp := RefreshResponseConverter(res)
	if err := ags.refreshToCookie(ctx, resp.GetTokens()); err != nil {
		return nil, err
	}

	return resp, nil
}

func (ags *AuthGrpcServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := ags.authcase.Logout(ctx, ags.refreshTokenFrom(ctx, req.GetRefreshToken())); err != nil {
		return nil, err
	}

	if cookieMode(ctx) {
		if err := setCookieHeader(ctx, ags.cookie.clear()); err != nil {
			return nil, err
		}
	}

	return &authv1.LogoutResponse{}, nil
}

// refreshToCookie для браузерных запросов переносит refresh-токен из тела в HttpOnly cookie.
func (ags *AuthGrpcServer) refreshToCookie(ctx context.Context, tokens *authv1.TokenPair) error {
	if tokens == nil || !cookieMode(ctx) {
		return nil
	}
	if err := setCookieHeader(ctx, ags.cookie.set(tokens.GetRefreshToken())); err != nil {
		return err
	}
	tokens.RefreshToken = ""

	return nil
}

// refreshTokenFrom берёт токен из тела, а если его там нет, из cookie.
func (ags *AuthGrpcServer) refreshTokenFrom(ctx context.Context, fromBody string) string {
	if fromBody != "" {
		return fromBody
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	r := &http.Request{Header: http.Header{"Cookie": md.Get(headerCookie)}}
	c, err := r.Cookie(ags.cookie.Name)
	if err != nil {
		return ""
	}

	return c.Value
}

func cookieMode(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)

	return ok && slices.Contains(md.Get(headerTokenTransport), transportCookie)
}

func setCookieHeader(ctx context.Context, value string) error {
	if err := grpc.SetHeader(ctx, metadata.Pairs(headerSetCookie, value)); err != nil {
		return apperr.Wrap(apperr.KindInternal, "", "failed to set cookie", err)
	}

	return nil
}
