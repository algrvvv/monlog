package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/algrvvv/monlog/internal/app"
	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/utils"
)

func APIGetLinesByID(serverLoggers []*app.ServerLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := utils.ValidateServerId(r.PathValue("id"))
		if serverID < 0 {
			utils.SendErrorJSON(w, "invalid server id", http.StatusBadRequest)
			return
		}
		serverLogger := serverLoggers[serverID]
		total, err := serverLogger.File.GetLineCount()
		logger.Info(fmt.Sprintf("Total lines readed in %s: %d", serverLogger.File.Name(), total))
		if err != nil {
			logger.Error(err.Error(), err)
			utils.SendErrorJSON(w, "Ошибка получения данных", http.StatusBadGateway)
			return
		}
		rows := config.Cfg.App.NumberRowsToLoad
		content := serverLogger.File.ReadLines(total-rows, total)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(map[string]interface{}{
			"total": total,
			"lines": content,
		})
		if err != nil {
			logger.Error(err.Error(), err)
			utils.SendErrorJSON(w, "Ошибка парсинга данных", http.StatusBadGateway)
			return
		}
		_, _ = w.Write(jsonData)
	}
}
