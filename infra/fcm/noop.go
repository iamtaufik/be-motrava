package fcm

import (
	"log/slog"

	fcmPort "motrava/core/port/fcm"
)

type noopClient struct {
	log *slog.Logger
}

func NewNoopFCMClient(logger *slog.Logger) fcmPort.FCMClient {
	logger.Warn("fcm client running in no-op mode", "module", "fcm")
	return &noopClient{log: logger}
}

func (n *noopClient) SendNotification(deviceToken string, notification fcmPort.FCMNotification) error {
	n.log.Info("fcm no-op: notification skipped", "module", "fcm", "title", notification.Title, "token", deviceToken[:20]+"...")
	return nil
}

func (n *noopClient) SendMulticast(deviceTokens []string, notification fcmPort.FCMNotification) []error {
	n.log.Info("fcm no-op: multicast skipped", "module", "fcm", "title", notification.Title, "count", len(deviceTokens))
	return make([]error, len(deviceTokens))
}
