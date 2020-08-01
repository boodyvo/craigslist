package notification_manager

type NotificationManager interface {
	Stop()
	Subscribe(notificationChan <-chan string)
}
