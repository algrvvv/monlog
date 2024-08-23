package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"

	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/logger"
)

func WsHandler(serverLoggers []*app.ServerLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Hijacker); !ok {
			logger.Error("ResponseWriter не реализует http.Hijacker", errors.New("ResponseWriter does not implement http.Hijacker"))
			//utils.RenderError(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		loggerIndex, err := strconv.Atoi(r.URL.Path[len("/ws/"):])
		if err != nil || loggerIndex > len(serverLoggers) || serverLoggers[loggerIndex] == nil {
			logger.Error(err.Error(), err)
			//utils.RenderError(w, "Передан некорректный айди сервера", http.StatusNotFound)
			return
		}

		logger.Info("new connection...")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error(err.Error(), err)
			//utils.RenderError(w, "Ошибка подключения вебсокетов", http.StatusBadRequest)
			return
		}
		logger.Info("ws connected")

		servLogger := serverLoggers[loggerIndex]
		logger.Info("got server: ", slog.Any("server", serverLoggers[loggerIndex]))
		servLogger.AppendWSConnection(conn)
		defer servLogger.RemoveWSConnection(conn)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("Connected to server"))

		for {
			_, _, err = conn.ReadMessage()
			if err != nil {
				logger.Warn(err.Error(), slog.Any("warn", err))
				break
			}
		}
	}
}
