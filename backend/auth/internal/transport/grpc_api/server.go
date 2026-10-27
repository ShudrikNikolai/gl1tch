package grpc_api

import (
	"strconv"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase"
	authv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/auth/v1"
	"google.golang.org/grpc"
)

// CookieConfig параметры cookie с refresh-токеном (для браузерных клиентов за Envoy).
type CookieConfig struct {
	Name   string
	Path   string // путь, как его видит браузер (с префиксом /api)
	Secure bool
	MaxAge time.Duration
}

type AuthGrpcServer struct {
	authv1.UnimplementedAuthServiceServer
	authcase usecase.AuthUseCase
	cookie   CookieConfig
}

func RegisterAuthGrpcServer(s grpc.ServiceRegistrar, authcase usecase.AuthUseCase, cookie CookieConfig) {
	authv1.RegisterAuthServiceServer(s, &AuthGrpcServer{authcase: authcase, cookie: cookie})
}

// set формирует значение заголовка Set-Cookie. Значение токена base64url, экранирование не нужно.
func (c CookieConfig) set(value string) string {
	return c.build(value, int(c.MaxAge.Seconds()))
}

// clear формирует Set-Cookie, который удаляет cookie в браузере.
func (c CookieConfig) clear() string {
	return c.build("", 0)
}

func (c CookieConfig) build(value string, maxAge int) string {
	parts := []string{
		c.Name + "=" + value,
		"HttpOnly",
		"SameSite=Strict",
		"Path=" + c.Path,
		"Max-Age=" + strconv.Itoa(maxAge),
	}
	if c.Secure {
		parts = append(parts, "Secure")
	}

	return strings.Join(parts, "; ")
}
