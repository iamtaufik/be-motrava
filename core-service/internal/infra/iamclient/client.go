package iamclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	iamv1 "motrava/proto/gen/iam/v1"
)

type Client struct {
	conn   *grpc.ClientConn
	client iamv1.IAMServiceClient
	log    *slog.Logger
}

func New(address string, logger *slog.Logger) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("iam grpc dial: %w", err)
	}

	client := iamv1.NewIAMServiceClient(conn)
	return &Client{conn: conn, client: client, log: logger}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*iamv1.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &iamv1.ValidateTokenRequest{Token: token})
	if err != nil {
		c.log.Error("iam grpc validate token failed", "module", "iam_client", "error", err)
		return nil, fmt.Errorf("iam validate token: %w", err)
	}

	return resp, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*iamv1.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &iamv1.GetUserRequest{UserId: userID})
	if err != nil {
		c.log.Error("iam grpc get user failed", "module", "iam_client", "error", err)
		return nil, fmt.Errorf("iam get user: %w", err)
	}

	return resp, nil
}
