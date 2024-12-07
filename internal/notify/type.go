package notify

import "github.com/algrvvv/monlog/internal/config"

type Notification struct {
	Server  *config.ServerConfig
	Time    string
	Level   string
	Message string
	Log     string
}

type NotificationSender interface {
	// Send метод для отправки уведомления. Принимает сервер типа config.ServerConfig и само сообщение.
	// Сервер нужен для того, чтобы оттуда достать данные для отправки, к примеру айди пользователей в тг.
	Send(notification *Notification) error
}
