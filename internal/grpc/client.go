package grpc

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ASTeterin/gophkeeper/api"
)

type Client struct {
	conn          *grpc.ClientConn
	userClient    pb.UserServiceClient
	dataClient    pb.PrivateDataServiceClient
	currentUserID uuid.UUID
}

func NewClient(address string) (*Client, error) {
	log.Printf("Attempting to connect to gRPC at: %s", address)
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		conn:       conn,
		userClient: pb.NewUserServiceClient(conn),
		dataClient: pb.NewPrivateDataServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Register(ctx context.Context, login, password string) (uuid.UUID, error) {
	resp, err := c.userClient.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return uuid.Nil, err
	}
	c.currentUserID, _ = uuid.Parse(resp.UserId)
	return c.currentUserID, nil
}

func (c *Client) Authenticate(ctx context.Context, login, password string) (uuid.UUID, error) {
	resp, err := c.userClient.Authenticate(ctx, &pb.AuthenticateRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return uuid.Nil, err
	}
	c.currentUserID, _ = uuid.Parse(resp.UserId)
	return c.currentUserID, nil
}

func (c *Client) AddData(ctx context.Context, key, desc string, data []byte) error {
	if c.currentUserID == uuid.Nil {
		return fmt.Errorf("not authenticated")
	}
	_, err := c.dataClient.Store(ctx, &pb.StoreRequest{
		UserId:      c.currentUserID.String(),
		DataKey:     key,
		Description: desc,
		Data:        data,
	})
	return err
}

func (c *Client) GetAllData(ctx context.Context) ([]*pb.PrivateData, error) {
	if c.currentUserID == uuid.Nil {
		return nil, fmt.Errorf("not authenticated")
	}
	resp, err := c.dataClient.GetAll(ctx, &pb.GetAllRequest{
		UserId: c.currentUserID.String(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetByKey(ctx context.Context, key string) (*pb.PrivateData, error) {
	if c.currentUserID == uuid.Nil {
		return nil, fmt.Errorf("not authenticated")
	}
	resp, err := c.dataClient.GetByKey(ctx, &pb.GetByKeyRequest{
		UserId: c.currentUserID.String(),
		Key:    key,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) SyncData(ctx context.Context, items []SyncItem) error {
	if c.currentUserID == uuid.Nil {
		return fmt.Errorf("not authenticated")
	}

	protoItems := make([]*pb.DataItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, &pb.DataItem{
			DataKey:     item.Key,
			Description: item.Description,
			Data:        item.Data,
		})
	}

	_, err := c.dataClient.ReplaceAll(ctx, &pb.ReplaceAllRequest{
		UserId: c.currentUserID.String(),
		Items:  protoItems,
	})
	return err
}

type SyncItem struct {
	Key         string
	Description string
	Data        []byte
}
