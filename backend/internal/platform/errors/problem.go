package errors

import (
	"encoding/json"
	"net/http"
	"time"
)

// Problem is the public RFC 9457 error contract. Detail must never contain secrets.
type Problem struct {
	Type        string       `json:"type"`
	Title       string       `json:"title"`
	Status      int          `json:"status"`
	Code        string       `json:"code"`
	Detail      string       `json:"detail"`
	TraceID     string       `json:"trace_id"`
	Timestamp   time.Time    `json:"timestamp"`
	FieldErrors []FieldError `json:"field_errors"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Write(w http.ResponseWriter, status int, code, title, detail, traceID string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type:  "https://api.appmarket.bookingnetwork.internal/errors/" + code,
		Title: title, Status: status, Code: code, Detail: detail,
		TraceID: traceID, Timestamp: time.Now().UTC(), FieldErrors: []FieldError{},
	})
}

func WriteProblem(w http.ResponseWriter, status int, code, detail, traceID string) {
	Write(w, status, code, code, detail, traceID)
}
