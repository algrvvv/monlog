package server

import (
	"fmt"
	"github.com/algrvvv/monlog/internal/server/handlers"
	"github.com/algrvvv/monlog/internal/server/middlewares"
	"net/http"

	"github.com/algrvvv/monlog/internal/config"
)

func NewServer() *http.Server {
	server := http.NewServeMux()

	server.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/home", http.StatusFound)
	})
	server.HandleFunc("GET /home", handlers.IndexHandler)
	server.HandleFunc("GET /logs/{id}", handlers.GetLogsByID)

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
	server.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	ws := http.NewServeMux()
	ws.HandleFunc("/", handlers.WsHandler)
	server.Handle("/ws", ws)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Cfg.App.Port),
		Handler: middlewares.LogRequest(server),
	}
}
