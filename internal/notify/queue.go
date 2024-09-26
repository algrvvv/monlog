package notify

import (
	"github.com/algrvvv/monlog/internal/logger"
)

// NQueue Notification Queue. Очередь уведомлений, которая содержит в себе канал для получения новых
// потенциально возможных уведомлений.
type NQueue struct {
	// Nchan канал, в который должны быть помещены потенциально возможные уведомления.
	Nchan chan *NQueueItem
}

// NQueueItem Notification Queue Item. Элемент, который может попасть в канал NQueue.Nchan
// для его дальнейшей обработки.
type NQueueItem struct {
	// ID интовое значение равное айдишнику сервера, строка которого была получена.
	// Значение айди берется из `config.yml`.
	ID int

	// Line строка, которая будет в дальнейшем обработана для отправки уведомления.
	Line string
}

// Notifier переменная типа NQueue - очередь для обработки новых потенциально возможных уведомлений.
// Для его использования сначала инициализируйте его, используя функцию InitNotifier.
// После этого запустите в новой горутине метод HandleNewItem для запуска обработки очереди.
var Notifier NQueue

// InitNotifier инициализация обработчика уведомлений Notifier
func InitNotifier() {
	Notifier = NQueue{
		Nchan: make(chan *NQueueItem),
	}
	logger.Info("Notifier initialized successfully")
}

// HandleNewItem метод, который служит обработчиком очереди. Он ловит данные из канала и обрабатывает их.
// TODO В дальнейшем стоит добавить использование контекста и многопоточности возможно.
// Для запуска работы инициализируйте notify.Notifier и запустите в отдельной горутине этот метод:
// go notify.Notifier.HandleNewItem()
func (n *NQueue) HandleNewItem(callback func(int, string)) {
	for item := range n.Nchan {
		callback(item.ID, item.Line)
	}
}
