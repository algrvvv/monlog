package notification_drivers

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/algrvvv/monlog/internal/logger/log"
	"github.com/algrvvv/monlog/internal/notify"
)

func init() {
	log.PrintInfo("load terminal notifier")

	notify.RegisterDriver("terminal", false, func() notify.NotificationSender {
		tn, err := NewTerminalNotifier()
		if err != nil {
			log.PrintErrorf("failed to load terminal notifier: %v", err)
			return nil
		}

		return tn
	})

	log.PrintInfo("terminal notifier loaded")
}

type TerminalNotifier struct{}

// NewTerminalNotifier функция для загрузки нового терминального
// отправителя уведомлений. Перед загрузкой идет проверка на наличие нужной утилиты
// с помощью которой и происходит уведомление
func NewTerminalNotifier() (notify.NotificationSender, error) {
	cmd := exec.Command("which", "terminal-notifierrrr")
	if err := cmd.Run(); err != nil {
		return nil, errors.New("util for notify not install")
	}

	return &TerminalNotifier{}, nil
}

// Send метод, который отправляет уведомление через утилиту
func (t *TerminalNotifier) Send(n *notify.Notification) error {
	title := fmt.Sprintf("\"Уведомление (%s)\"", n.Server.Name)
	message := fmt.Sprintf("\"[%s] %s\"", n.Level, n.Message)

	args := []string{
		"-title", title, "-message", message, "-timeout", "10",
		"-sound", "Glass",
	}

	cmd := exec.Command("terminal-notifier", args...)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
