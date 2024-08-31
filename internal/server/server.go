package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/server/handlers"
	"github.com/algrvvv/monlog/internal/server/middlewares"
	"net/http"
)

func NewServer() (*http.Server, []*app.ServerLogger) {
	servers := config.Cfg.Servers
	servLoggers := make([]*app.ServerLogger, len(servers))
	for i, server := range servers {
		serverLogger := app.NewServerLogger(i, server)
		servLoggers[i] = serverLogger
	}

	server := http.NewServeMux()

	server.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/home", http.StatusFound)
	})
	server.HandleFunc("GET /home", handlers.IndexHandler)
	server.HandleFunc("GET /logs/{id}", handlers.GetLogsByID)

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /logs/prev/{id}", handlers.APIGetLinesByID(servLoggers))
	v1.HandleFunc("GET /logs/prev/count/{id}", handlers.APIGetPrevLogsByCount(servLoggers))
	server.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	server.HandleFunc("/ws/{id}", handlers.WsHandler(servLoggers))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Cfg.App.Port),
		Handler: middlewares.LogRequest(server),
	}, servLoggers
}

func RunServer(serv *http.Server, ctx context.Context) {
	logger.Info(fmt.Sprintf("Starting server on :%d", config.Cfg.App.Port))
	if err := serv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logger.Error(err.Error(), err)
	}

	<-ctx.Done()

	logger.Info("Shutting down server...")
	if err := serv.Shutdown(ctx); err != nil {
		logger.Error(err.Error(), err)
		return
	}
	logger.Info("Shutdown server complete")
}
