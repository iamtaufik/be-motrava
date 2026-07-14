package fcm

type FCMNotification struct {
	Title string
	Body  string
	Data  map[string]string
}

type FCMClient interface {
	SendNotification(deviceToken string, notification FCMNotification) error
	SendMulticast(deviceTokens []string, notification FCMNotification) []error
}
