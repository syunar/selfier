package notification

import "fmt"

type notificationProviderResend struct {
	// TODO: implement
}

func NewNotificationProvider() NotificationProvider {
	return &notificationProviderResend{}
}

func (n *notificationProviderResend) SendNotification() {
	// TODO: implement
	fmt.Print("send notification")
}
