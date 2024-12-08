package notify

import "github.com/algrvvv/monlog/internal/config"

type Notification struct {
	// Server информация о сервере
	Server *config.ServerConfig
	// Time время из полученного лого
	Time string
	// Level уровень лога
	Level string
	// Message сообщение полученное из лога
	Message string
	// Log полная строка лога
	Log string
}

type NotificationSender interface {
	// Send метод для отправки уведомления. Уже настроенный параметр для уведомлений, который в себе содежит в себе
	// все данные. подробнее смотреть Notification
	Send(notification *Notification) error
}
