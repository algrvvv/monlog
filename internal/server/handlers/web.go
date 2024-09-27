package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/utils"
)

type serverForTemp struct {
	ID   int
	Name string
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
)

func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	temp, err := template.ParseFiles("templates/home.html")
	if err != nil {
		logger.Error(err.Error(), err)
		utils.RenderError(w, "Произошла ошибка на сервере", http.StatusInternalServerError)
		return
	}

	var servers []serverForTemp
	for i, s := range config.Cfg.Servers {
		if !s.Enabled {
			continue
		}
		serv := serverForTemp{
			ID:   i,
			Name: s.Name,
		}
		servers = append(servers, serv)
	}

	_ = temp.Execute(w, servers)
}

func GetLogsByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error(err.Error(), err)
		utils.RenderError(w, "Предоставлен некорректный айди сервера", http.StatusBadRequest)
		return
	}
	servers := config.Cfg.Servers

	if id >= 0 && id < len(servers) {
		var temp *template.Template
		temp, err = template.ParseFiles("templates/log.html")
		if err != nil {
			logger.Error(err.Error(), err)
			utils.RenderError(w, "Ошибка парсинга страницы", http.StatusInternalServerError)
			return
		}

		_ = temp.Execute(w, struct {
			RowsLoad int
		}{
			RowsLoad: config.Cfg.App.NumberRowsToLoad,
		})
	} else {
		utils.RenderError(w, "Предоставлен некорректный айди сервера", http.StatusBadRequest)
	}
}
