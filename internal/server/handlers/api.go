package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/utils"
)

func GetLinesByID(serverLoggers []*app.ServerLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := utils.ValidateServerID(r.PathValue("id"))
		if serverID < 0 {
			utils.SendErrorJSON(w, "invalid server id", http.StatusBadRequest)
			return
		}
		serverLogger := serverLoggers[serverID]
		total, err := serverLogger.File.GetLineCount()
		logger.Info(fmt.Sprintf("Total lines readed in %s: %d", serverLogger.File.Name(), total))
		if err != nil {
			logger.Error("Failed to get total lines file: "+err.Error(), err)
			utils.SendErrorJSON(w, "Ошибка получения данных", http.StatusBadGateway)
			return
		}
		rows := config.Cfg.App.NumberRowsToLoad
		content := serverLogger.File.ReadLines(
			total-rows,
			total,
			config.Cfg.Servers[serverID].LogDriver,
		)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(map[string]interface{}{
			"total":    total,
			"lines":    content,
			"lastLine": total - rows,
		})
		if err != nil {
			logger.Error(err.Error(), err)
			utils.SendErrorJSON(w, "Ошибка парсинга данных", http.StatusBadGateway)
			return
		}
		_, _ = w.Write(jsonData)
	}
}

func GetPrevLogsByCount(serverLoggers []*app.ServerLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := utils.ValidateServerID(r.PathValue("id"))
		if serverID < 0 {
			utils.SendErrorJSON(w, "invalid server id", http.StatusBadRequest)
			return
		}
		serverLogger := serverLoggers[serverID]
		startLine, err := strconv.Atoi(r.URL.Query().Get("start"))
		if err != nil {
			utils.SendErrorJSON(w, "Invalid start line", http.StatusBadRequest)
			logger.Error("Invalid start line: "+err.Error(), err)
			return
		}
		endLine, err := strconv.Atoi(r.URL.Query().Get("end"))
		if err != nil {
			utils.SendErrorJSON(w, "Invalid end line", http.StatusBadRequest)
			logger.Error("Invalid end line: "+err.Error(), err)
			return
		}

		total, err := serverLogger.File.GetLineCount()
		if err != nil {
			utils.SendErrorJSON(w, "Ошибка получения данных", http.StatusBadGateway)
			logger.Error("Failed to get total lines file: "+err.Error(), err)
			return
		}

		if startLine > endLine || startLine > total || endLine > total || startLine < 0 ||
			endLine < 0 {
			utils.SendErrorJSON(w, "one of the parameters is incorrect", http.StatusBadRequest)
			logger.Error("one of the parameters is incorrect", nil)
			return
		}

		lines := serverLogger.File.ReadLines(
			startLine,
			endLine,
			config.Cfg.Servers[serverID].LogDriver,
		)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(map[string]interface{}{
			"lastLine": startLine,
			"lines":    lines,
		})
		if err != nil {
			logger.Error(err.Error(), err)
			utils.SendErrorJSON(w, "Ошибка парсинга данных", http.StatusBadGateway)
			return
		}
		_, _ = w.Write(jsonData)
	}
}

func GetAllPrevLogs(serverLoggers []*app.ServerLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := utils.ValidateServerID(r.PathValue("id"))
		if serverID < 0 {
			utils.SendErrorJSON(w, "invalid server id", http.StatusBadRequest)
			return
		}
		serverLogger := serverLoggers[serverID]
		targetLine, err := strconv.Atoi(r.URL.Query().Get("target"))
		if err != nil {
			utils.SendErrorJSON(w, "Invalid start line", http.StatusBadRequest)
			logger.Error("Invalid start line: "+err.Error(), err)
			return
		}

		total, err := serverLogger.File.GetLineCount()
		if err != nil {
			utils.SendErrorJSON(w, "Ошибка получения данных", http.StatusBadGateway)
			logger.Error("Failed to get total lines file: "+err.Error(), err)
			return
		}

		if targetLine > total || targetLine < 0 {
			utils.SendErrorJSON(w, "one of the parameters is incorrect", http.StatusBadRequest)
			logger.Error("one of the parameters is incorrect", nil)
			return
		}

		reader := utils.ReaderCallback(config.Cfg.Servers[serverID].LogDriver)
		serverLogger.File.ReadFullFile(targetLine, reader)

		jsonData, err := json.Marshal(map[string]interface{}{
			"lastLine": 0,
			"lines":    reader([]byte("\nConnected to server\n")),
		})
		if err != nil {
			logger.Error(err.Error(), err)
			utils.SendErrorJSON(w, "Ошибка парсинга данных", http.StatusBadGateway)
			return
		}
		_, _ = w.Write(jsonData)
	}
}
