package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/algrvvv/monlog/internal/config"
	"github.com/algrvvv/monlog/internal/logger"
	"github.com/algrvvv/monlog/internal/utils"
)

type serverForTemp struct {
	ID   int
	Addr string
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	Mu sync.Mutex
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	temp, err := template.ParseFiles("templates/home.html")
	if err != nil {
		logger.Error(err.Error(), err)
		utils.RenderError(w, "Произошла ошибка на сервере", http.StatusInternalServerError)
		return
	}

	var servers []serverForTemp
	for i, s := range config.Cfg.Servers {
		serv := serverForTemp{
			ID:   i,
			Addr: fmt.Sprintf("%s:%d", s.Host, s.Port),
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

		_ = temp.Execute(w, nil)
	} else {
		utils.RenderError(w, "Предоставлен некорректный айди сервера", http.StatusBadRequest)
	}
}
