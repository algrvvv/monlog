package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func SendErrorJSON(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonData := map[string]string{
		"status":  strconv.Itoa(status),
		"message": message,
	}
	marshalledData, _ := json.Marshal(jsonData)
	_, _ = w.Write(marshalledData)
}
