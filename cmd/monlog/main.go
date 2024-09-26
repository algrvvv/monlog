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
	"github.com/algrvvv/monlog/internal/notify"
	"github.com/algrvvv/monlog/internal/server"
	"github.com/algrvvv/monlog/internal/state"
)

func main() {
	err := logger.NewLogger("monlog.log")
	if err != nil {
		log.Fatal(err)
	}

	if err = config.LoadConfig("config.yml"); err != nil {
		logger.Fatal(err.Error(), err)
	}

	if err = state.InitializeState(); err != nil {
		logger.Fatal(err.Error(), err)
	}

	if err = notify.LoadSenders(); err != nil {
		logger.Fatal(err.Error(), err)
	}

	notify.InitNotifier()
	go notify.Notifier.HandleNewItem(state.ParseLineAndSendNotify)

	serv, serverLoggers := server.NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for _, sl := range serverLoggers {
		wg.Add(1)
		go sl.StartLogging(ctx, wg)
	}

	go func() {
		logger.Info(fmt.Sprintf("Starting server on :%d", config.Cfg.App.Port))
		if err = serv.ListenAndServe(); err != nil {
			logger.Error(err.Error(), err)
		}
	}()

	<-sigs
	logger.Info("Received signal, shutting down...")

	cancel()
	wg.Wait()

	logger.Info("Server shutting down...")
	if err = serv.Shutdown(ctx); err != nil {
		logger.Error(err.Error(), err)
		return
	}

	logger.Info("Server shutdown completed successfully")
}
