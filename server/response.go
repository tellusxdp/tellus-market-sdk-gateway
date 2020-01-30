package server

import (
	"encoding/json"
	"net/http"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

type LoggingResponseWriter struct {
	responseWriter http.ResponseWriter
	StatusCode     int
	HeaderMap      http.Header
	Size           int
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	msg := &ErrorMessage{Error: message}

	j, _ := json.Marshal(msg)
	w.Write(j)
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, 200, w.Header(), 0}
}

func (w *LoggingResponseWriter) Header() http.Header {
	header := w.responseWriter.Header()
	w.HeaderMap = header
	return header
}

func (w *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.responseWriter.Write(b)
	w.Size += size
	return size, err
}

func (w *LoggingResponseWriter) WriteHeader(statusCode int) {
	w.responseWriter.WriteHeader(statusCode)
	w.StatusCode = statusCode
}
