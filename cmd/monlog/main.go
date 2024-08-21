package main

import (
	"fmt"
	"github.com/algrvvv/monlog/internal"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/server"
	"log"
)

func main() {
	err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	err = logger.NewLogger("monlog.log")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for i, s := range config.Cfg.Servers {
			go internal.ConnectAndReadLogs(i, s)
		}
	}()

	serv := server.NewServer()
	logger.Info(fmt.Sprintf("Starting server on :%d", config.Cfg.App.Port))
	if err = serv.ListenAndServe(); err != nil {
		logger.Error(err.Error(), err)
	}
}
