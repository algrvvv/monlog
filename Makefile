install:
	@chmod +x install.sh
	@./install.sh

format:
	@goimports -w .

build:
	mkdir -p bin/
	go build -o bin/monlog cmd/monlog/main.go

run:
	@go run cmd/monlog/main.go

dev: format run
