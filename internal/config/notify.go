package config

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	subscribers = make([]chan struct{}, 0)
	mu          = &sync.RWMutex{}

	// переменная, которая сохраняет состояние, если вдруг во время новой подписки
	// конфиг уже был загружен
	isLoaded atomic.Bool
	// ErrConfigAlreadyLoaded ошибка, говорящая о том, что конфигурация уже загружена
	ErrConfigAlreadyLoaded = errors.New("config already loaded")
)

// AddSubscriber функция, которая добавляет подписчиков на рассылку уведомлений
// в момент, когда конфигурация приложения будет загружена.
// NOTE: полезна для lazyLoad драйверов
func AddSubscriber(nch chan struct{}) error {
	if isLoaded.Load() {
		return ErrConfigAlreadyLoaded
	}

	mu.Lock()
	defer mu.Unlock()

	subscribers = append(subscribers, nch)
	return nil
}

func notifySubscibers() {
	mu.RLock()
	defer mu.RUnlock()
	defer isLoaded.Store(true)

	for _, ch := range subscribers {
		close(ch)
	}
}
