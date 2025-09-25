// Package notification contains the ports or interefaces for the notification module
package notification

type NotificationService interface {
	SendNotification()
}

type NotificationProvider interface {
	SendNotification()
}
