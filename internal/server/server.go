package server

import (
	"fmt"
	"net/http"

	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/server/handlers"
	"github.com/algrvvv/monlog/internal/server/middlewares"
)

func NewServer() (*http.Server, []*app.ServerLogger) {
	servers := config.Cfg.Servers
	servLoggers := make([]*app.ServerLogger, len(servers))
	for i, server := range servers {
		logger := app.NewServerLogger(i, server)
		servLoggers[i] = logger
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
	server.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	server.HandleFunc("/ws/{id}", handlers.WsHandler(servLoggers))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Cfg.App.Port),
		Handler: middlewares.LogRequest(server),
	}, servLoggers
}
