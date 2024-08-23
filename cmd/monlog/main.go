package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/server"
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

	serv, serverLoggers := server.NewServer()
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, sl := range serverLoggers {
		wg.Add(1)
		go sl.StartLogging(ctx, &wg)
	}

	go func() {
		logger.Info(fmt.Sprintf("Starting server on :%d", config.Cfg.App.Port))
		if err = serv.ListenAndServe(); err != nil {
			logger.Error(err.Error(), err)
		}
	}()

	<-signs
	logger.Info("Received signal, shutting down...")

	cancel()
	// wg.Wait()

	logger.Info("Start Deleting local logs")
	for _, sl := range serverLoggers {
		if err = sl.File.CLoseAndRemove(); err != nil {
			logger.Error(err.Error(), err)
		}
	}
	logger.Info("Start Deleting all local logs")

	if err = serv.Shutdown(ctx); err != nil {
		logger.Error(err.Error(), err)
		return
	}

	logger.Info("Shutdown server complete")
}
