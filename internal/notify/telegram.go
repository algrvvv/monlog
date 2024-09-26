package notify

import (
	"errors"
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
)

type TelegramSender struct {
	Bot *tgbotapi.BotAPI
}

// NewTelegramSender функция для инициализации тг бота
func NewTelegramSender() (*TelegramSender, error) {
	botToken := config.Cfg.App.TGBotToken
	if botToken == "" {
		return nil, errors.New("telegram bot token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, errors.New("telegram bot init fail" + err.Error())
	}
	bot.Debug = true // TODO сделать доп конфигурацию
	return &TelegramSender{bot}, nil
}

// Send метод notify.TelegramSender для отправки уведомления по средствам тг бота.
// Для того, чтобы уведомление пришло - обзятаельно с учетной записи, айди которой вы указываете
// в `chat_ids`, нужно написать /start боту, иначе он просто не сможет отправить вам уведомление
func (t TelegramSender) Send(server config.ServerConfig, message string) error {
	chatIDs := server.ChatIDs
	for _, chatID := range chatIDs {
		go func() {
			id, err := strconv.ParseInt(chatID, 10, 64)
			if err != nil {
				logger.Error("Failed to convert chat id: "+err.Error(), err, slog.Any("chat_id", chatID))
				return
			}
			msg := tgbotapi.NewMessage(id, message)
			msg.ParseMode = tgbotapi.ModeHTML
			_, err = t.Bot.Send(msg)
			if err != nil {
				logger.Error("Failed to send message: "+err.Error(), err, slog.Any("chat_id", chatID))
				return
			}
			logger.Info("Telegram notify successfully sent")
		}()
	}
	return nil
}
