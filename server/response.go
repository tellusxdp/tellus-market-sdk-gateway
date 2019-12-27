package server

import "net/http"

import "encoding/json"

type ErrorMessage struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	msg := &ErrorMessage{Error: message}

	j, _ := json.Marshal(msg)
	w.Write(j)
}
