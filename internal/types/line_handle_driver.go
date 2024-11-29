package types

type LineHandleDriver interface {
	// GetName метод для получения названия драйвера
	GetName() (name string)
	// Handle метод для вашей кастомной обработки строки лога
	Handle(line string) (result string)
}
