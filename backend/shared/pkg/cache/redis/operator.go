package redis

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
)

func (c *Client) SAdd(ctx context.Context, key, value string) error {
	_, err := c.do(ctx, "SADD", key, value)
	return err
}

func (c *Client) SRem(ctx context.Context, key, value string) error {
	_, err := c.do(ctx, "SREM", key, value)
	return err
}

func (c *Client) SIsMember(ctx context.Context, key, value string) (bool, error) {
	return redigo.Bool(c.do(ctx, "SISMEMBER", key, value))
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return redigo.Strings(c.do(ctx, "SMEMBERS", key))
}
