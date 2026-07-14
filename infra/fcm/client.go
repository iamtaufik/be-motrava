package fcm

import (
	"context"
	"fmt"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	fcmPort "motrava/core/port/fcm"
)

type firebaseClient struct {
	client *messaging.Client
	log    *slog.Logger
}

func NewFCMClient(ctx context.Context, credentialsFile, projectID string, logger *slog.Logger) (fcmPort.FCMClient, error) {
	opts := []option.ClientOption{}
	if credentialsFile != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, credentialsFile))
	}

	var app *firebase.App
	var err error
	if projectID != "" {
		app, err = firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, opts...)
	} else {
		app, err = firebase.NewApp(ctx, nil, opts...)
	}

	if err != nil {
		return nil, fmt.Errorf("firebase init: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging: %w", err)
	}

	logger.Info("fcm client initialized", "module", "fcm", "project_id", projectID)
	return &firebaseClient{client: client, log: logger}, nil
}

func (f *firebaseClient) SendNotification(deviceToken string, notification fcmPort.FCMNotification) error {
	msg := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
	}

	if notification.Data != nil {
		msg.Data = notification.Data
	}

	_, err := f.client.Send(context.Background(), msg)
	if err != nil {
		f.log.Error("fcm send failed", "module", "fcm", "error", err)
		return fmt.Errorf("fcm send: %w", err)
	}

	f.log.Info("fcm sent", "module", "fcm", "token", deviceToken[:20]+"...")
	return nil
}

func (f *firebaseClient) SendMulticast(deviceTokens []string, notification fcmPort.FCMNotification) []error {
	msg := &messaging.MulticastMessage{
		Tokens: deviceTokens,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
	}

	if notification.Data != nil {
		msg.Data = notification.Data
	}

	resp, err := f.client.SendEachForMulticast(context.Background(), msg)
	if err != nil {
		f.log.Error("fcm multicast failed", "module", "fcm", "error", err)
		return []error{fmt.Errorf("fcm multicast: %w", err)}
	}

	errs := make([]error, len(resp.Responses))
	for i, r := range resp.Responses {
		if r.Error != nil {
			errs[i] = r.Error
			f.log.Warn("fcm multicast partial failure", "module", "fcm", "index", i, "error", r.Error)
		}
	}

	f.log.Info("fcm multicast sent", "module", "fcm", "success", resp.SuccessCount, "failure", resp.FailureCount)
	return errs
}
