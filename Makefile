format:
	@goimports -w .

build:
	mkdir -p bin/linux
	go build -o bin/monlog cmd/monlog/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/monlog cmd/monlog/main.go

run:
	@go run cmd/monlog/main.go

dev: format run
