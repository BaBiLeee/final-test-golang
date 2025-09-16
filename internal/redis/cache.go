package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	r *redis.Client
}

func New(r *redis.Client) *Client { return &Client{r: r} }

func (c *Client) GetPost(ctx context.Context, id int) (*db.Post, error) {
	key := fmt.Sprintf("post:%d", id)
	val, err := c.r.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var p db.Post
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) SetPost(ctx context.Context, p *db.Post, ttl time.Duration) error {
	key := fmt.Sprintf("post:%d", p.ID)
	b, _ := json.Marshal(p)
	return c.r.Set(ctx, key, b, ttl).Err()
}

func (c *Client) DeletePost(ctx context.Context, id int) error {
	key := fmt.Sprintf("post:%d", id)
	return c.r.Del(ctx, key).Err()
}
