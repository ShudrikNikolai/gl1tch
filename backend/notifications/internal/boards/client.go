// Package boards — минимальный gRPC-клиент к сервису boards.
package boards

import (
	"context"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/grpcx"
	boardsv1 "github.com/ShudrikNikolai/gl1tch/backend/shared/proto/api/boards/v1"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	api  boardsv1.PostServiceClient
}

func New(target string) (*Client, error) {
	conn, err := grpcx.Dial(target, 0)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, api: boardsv1.NewPostServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) HasUserPosted(ctx context.Context, threadID, userID string) (bool, error) {
	resp, err := c.api.HasUserPosted(ctx, &boardsv1.HasUserPostedRequest{ThreadId: threadID, UserId: userID})
	if err != nil {
		return false, err
	}

	return resp.GetPosted(), nil
}
