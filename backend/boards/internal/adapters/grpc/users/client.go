// Package users клиент к сервису users. Возвращает модели usecase, protobuf-типы дальше не уходят.
package users

import (
	"context"
	"fmt"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
	"google.golang.org/grpc"
)

type Client struct {
	conn  *grpc.ClientConn
	api   usersv1.UserServiceClient
	cache *cache
}

// New создаёт клиент. Соединение ленивое: недоступный users не мешает старту сервиса.
// target в формате gRPC, например "dns:///users:9090". cacheTTL 0 отключает кеш.
func New(target string, cacheTTL time.Duration) (*Client, error) {
	conn, err := grpcx.Dial(target, 0) // 0: дефолтный таймаут вызова из grpcx
	if err != nil {
		return nil, fmt.Errorf("users: dial %q: %w", target, err)
	}

	return &Client{conn: conn, api: usersv1.NewUserServiceClient(conn), cache: newCache(cacheTTL)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Get возвращает профиль пользователя. Контекст несёт личность вызывающего (x-user-id, x-roles),
// grpcx передаёт её в users как есть. Кеш общий для всех вызовов.
func (c *Client) Get(ctx context.Context, id string) (models.User, error) {
	if u, ok := c.cache.get(id); ok {
		return u, nil
	}

	resp, err := c.api.GetUser(ctx, &usersv1.GetUserRequest{UserId: id})
	if err != nil {
		return models.User{}, err
	}

	u := toUser(resp.GetUser())
	c.cache.put(u)

	return u, nil
}
