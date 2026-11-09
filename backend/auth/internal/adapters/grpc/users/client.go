package users

import (
	"context"
	"fmt"

	"github.com/ShudrikNikolai/gl1tch/backend/auth/internal/usecase/models"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	usersv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/users/v1"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	api  usersv1.UserServiceClient
}

// New создаёт клиент
// target в формате gRPC -> "dns:///users:9090"
func New(target string) (*Client, error) {
	conn, err := grpcx.Dial(target, 0) // 0: дефолтный таймаут вызова из grpcx
	if err != nil {
		return nil, fmt.Errorf("users: dial %q: %w", target, err)
	}

	return &Client{conn: conn, api: usersv1.NewUserServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateUser(ctx context.Context, id, username string) (models.User, error) {
	resp, err := c.api.CreateUser(ctx, &usersv1.CreateUserRequest{UserId: id, Username: username})
	if err != nil {
		return models.User{}, err
	}

	return toUser(resp.GetUser()), nil
}

func (c *Client) GetUser(ctx context.Context, id string) (models.User, error) {
	resp, err := c.api.GetUser(ctx, &usersv1.GetUserRequest{UserId: id})
	if err != nil {
		return models.User{}, err
	}

	return toUser(resp.GetUser()), nil
}
