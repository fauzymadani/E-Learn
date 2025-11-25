package grpcclient

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "elearning/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	client pb.NotificationServiceClient
	conn   *grpc.ClientConn
}

func NewNotificationClient(address string) (*NotificationClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}

	client := pb.NewNotificationServiceClient(conn)
	log.Printf("connected to notification service at %s", address)

	return &NotificationClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *NotificationClient) Close() error {
	return c.conn.Close()
}

func (c *NotificationClient) SendNotification(ctx context.Context, userID int64, notifType, title, message string) error {
	_, err := c.client.SendNotification(ctx, &pb.SendNotificationRequest{
		UserId:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
	})
	return err
}

func (c *NotificationClient) GetNotifications(ctx context.Context, userID int64, page, limit int32, unreadOnly bool) (*pb.GetNotificationsResponse, error) {
	return c.client.GetNotifications(ctx, &pb.GetNotificationsRequest{
		UserId:     userID,
		Page:       page,
		Limit:      limit,
		UnreadOnly: unreadOnly,
	})
}

func (c *NotificationClient) MarkAsRead(ctx context.Context, notificationID, userID int64) error {
	_, err := c.client.MarkAsRead(ctx, &pb.MarkAsReadRequest{
		NotificationId: notificationID,
		UserId:         userID,
	})
	return err
}

func (c *NotificationClient) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	resp, err := c.client.GetUnreadCount(ctx, &pb.GetUnreadCountRequest{
		UserId: userID,
	})
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (c *NotificationClient) MarkAllAsRead(ctx context.Context, userID int64) (int32, error) {
	resp, err := c.client.MarkAllAsRead(ctx, &pb.MarkAllAsReadRequest{
		UserId: userID,
	})
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
