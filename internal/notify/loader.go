package notify

import (
	"strings"

	"github.com/algrvvv/monlog/internal/config"
)

// структура ошибки с динамическим полем названия драйвера
type ErrDriverNotFound struct {
	driverName string
}

func (e *ErrDriverNotFound) Error() string {
	return "driver not found: " + e.driverName
}

type DriverFactory func() NotificationSender

var drivers = make(map[string]NotificationSender)

// RegisterDriver функция для регистрации нового драйвера для уведомления пользователя
// о каком то событии на сервере.
// Параметры:
//
//	name - название драйвера
//	lazyLoad - перед загрузкой дождаться загрузки конфигурации
//	factory - функция для подготовки и загрузке вашего драйвера
func RegisterDriver(name string, lazyLoad bool, factory DriverFactory) {
	if _, ok := drivers[name]; ok {
		panic("Driver already registered: " + name)
	}

	if !lazyLoad {
		drivers[name] = factory()
		return
	}

	go func() {
		nCh := make(chan struct{})

		err := config.AddSubscriber(nCh)
		if err == nil {
			<-nCh
		}

		drivers[name] = factory()
	}()
}

// Notify функция для отправки нового уведомления.
// Самостоятельно проверяет состояние и наличие нужного драйвера
func Notify(n *Notification) error {
	driverName := strings.TrimSpace(n.Server.Notify)
	if driverName == "" || driverName == "none" {
		return nil
	}

	if driver, ok := drivers[driverName]; ok {
		return driver.Send(n)
	}

	return &ErrDriverNotFound{driverName: driverName}
}
