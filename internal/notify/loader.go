package notify

// Telegram переменная, которая сохраняет в себе уже
// сконфигурированный тг бот для отправки уведомления.
// Используйте его для отправки уведомлений в методе notify.SendNotification
// Usage:
//
//	if err = notify.SendNotification(notify.Telegram, server, msg); err != nil {
//		logger.Error("Error sending notification: "+err.Error(), err)
//	}
var Telegram *TelegramSender

// LoadSenders функция для инициализации возможных способов отправки данных.
// Для добавления новых создайте в этом пакете новый файл, создайте структуру,
// которая реализует интерфейс notify.NotificationSender.
// А затем также как и с тг добавить константу и в функцию отправки данных
// передать вместо notify.Telegram свою переменную.
func LoadSenders() error {
	bot, err := NewTelegramSender()
	if err != nil {
		return err
	}
	Telegram = bot
	return nil
}
