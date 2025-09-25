package notification

type notificationServiceImpl struct {
	notificationProvider NotificationProvider
}

func NewNotificationService(provider NotificationProvider) NotificationService {
	return &notificationServiceImpl{notificationProvider: provider}
}

func (s *notificationServiceImpl) SendNotification() {
	s.notificationProvider.SendNotification()
}
