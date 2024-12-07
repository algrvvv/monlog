package notification_drivers

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/algrvvv/monlog/internal/notify"
)

func init() {
	log.Println("[INFO] load terminal notifier")

	notify.RegisterDriver("terminal", false, func() notify.NotificationSender {
		return NewTerminalNotifier()
	})

	log.Println("[INFO] terminal notifier loaded")
}

type TerminalNotifier struct{}

// NewTerminalNotifier функция для загрузки нового терминального
// отправителя уведомлений. Перед загрузкой идет проверка на наличие нужной утилиты
// с помощью которой и происходит уведомление
func NewTerminalNotifier() notify.NotificationSender {
	cmd := exec.Command("which", "terminal-notifier")
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to load terminal notifier: %v", err)
	}

	return &TerminalNotifier{}
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
