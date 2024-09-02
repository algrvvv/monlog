package notify

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/utils"
)

type TelegramSender struct {
}

func NewTelegramSender() *TelegramSender {
	return &TelegramSender{}
}

func (t TelegramSender) Send(serverID int, message string) error {
	if sid := utils.ValidateServerId(strconv.Itoa(serverID)); sid == -1 {
		return errors.New("invalid server id")
	}
	chatIDs := config.Cfg.Servers[serverID].ChatIDs
	for _, chatID := range chatIDs {
		fmt.Printf("[%v] sending message: %s\n", chatID, message)
	}
	return nil
}
