package notify

type NotificationSender interface {
	Send(serverID int, message string) error
}

func SendNotification(sender NotificationSender, serverID int, message string) error {
	return sender.Send(serverID, message)
}
