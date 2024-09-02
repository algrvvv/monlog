package utils

import (
	"html/template"
	"net/http"

	"github.com/algrvvv/monlog/internal/logger"
)

type templateError struct {
	Error      string // сообщение об ошибке
	StatusCode int    // статус ответа
	StatusText string // текстовый статус ответа
}

// RenderError TODO - исправить
func RenderError(w http.ResponseWriter, message string, statusCode int) {
	temp, err := template.New("error").ParseFiles("templates/error.html")
	if err != nil {
		logger.Error(err.Error(), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	te := templateError{
		message, statusCode, http.StatusText(statusCode),
	}
	err = temp.Execute(w, te)
	if err != nil {
		logger.Error(err.Error(), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
