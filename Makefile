# действия при первоначальном запуске
install:
	@chmod +x install.sh
	@./install.sh

# использование линтеров
lin:
	@golangci-lint run

# билд проекта
build:
	@mkdir -p bin/
	@go build -o bin/monlog cmd/monlog/main.go

# запуск
run:
	@go run cmd/monlog/main.go

# использование линтеров и запус
dev: lin run
