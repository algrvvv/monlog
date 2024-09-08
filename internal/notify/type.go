package notify

import "github.com/algrvvv/monlog/internal/config"

type NotificationSender interface {
	// Send метод для отправки уведомления. Принимает сервер типа config.ServerConfig и само сообщение.
	// Сервер нужен для того, чтобы оттуда достать данные для отправки, к примеру айди пользователей в тг.
	Send(server config.ServerConfig, message string) error
}

func SendNotification(sender NotificationSender, server config.ServerConfig, message string) error {
	return sender.Send(server, message)
}
