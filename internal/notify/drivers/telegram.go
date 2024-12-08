package notification_drivers

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger/log"
	"github.com/algrvvv/monlog/internal/notify"
)

func init() {
	log.PrintInfo("load telegram notification driver")

	notify.RegisterDriver("telegram", true, func() notify.NotificationSender {
		bot, err := NewTelegramSender()
		if err != nil {
			log.PrintFatalf("failed to load telegram notification driver: %v", err)
			return nil
		}

		log.PrintInfo("telegram notification driver loaded")
		return bot
	})
}

type TelegramSender struct {
	Bot *tgbotapi.BotAPI
}

// NewTelegramSender функция для инициализации тг бота.
func NewTelegramSender() (notify.NotificationSender, error) {
	botToken := config.Cfg.App.TGBotToken
	if botToken == "" {
		return nil, errors.New("telegram bot token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, errors.New("telegram bot init fail" + err.Error())
	}
	bot.Debug = config.Cfg.App.Debug
	return &TelegramSender{bot}, nil
}

// Send метод notify.TelegramSender для отправки уведомления по средствам тг бота.
// Для того, чтобы уведомление пришло - обязательно с учетной записи, айди которой вы указываете
// в `chat_ids`, нужно написать /start боту, иначе он просто не сможет отправить вам уведомление.
func (t TelegramSender) Send(n *notify.Notification) error {
	message := fmt.Sprintf(
		"[%d] <u>%s</u>\n<b>Время:</b> %s\n<b>Уровень:</b> %s\n<b>Сообщение:</b> %s\n<b>Полная строка:</b> %s",
		n.Server.ID,
		n.Server.Name,
		n.Time,
		n.Level,
		n.Message,
		n.Log,
	)

	chatIDs := n.Server.Recipients
	for _, chatID := range chatIDs {
		go func() {
			id, err := strconv.ParseInt(chatID, 10, 64)
			if err != nil {
				log.PrintError(
					"Failed to convert chat id: "+err.Error(),
					err,
					slog.Any("chat_id", chatID),
				)
				return
			}
			msg := tgbotapi.NewMessage(id, message)
			msg.ParseMode = tgbotapi.ModeHTML
			_, err = t.Bot.Send(msg)
			if err != nil {
				log.PrintError(
					"Failed to send message: "+err.Error(),
					err,
					slog.Any("chat_id", chatID),
				)
				return
			}
			log.PrintInfo("Telegram notify successfully sent")
		}()
	}
	return nil
}
