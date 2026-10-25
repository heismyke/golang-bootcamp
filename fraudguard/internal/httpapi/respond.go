package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"fraudguard/internal/transaction"
)

const maxBodyBytes = 1 << 20 // 1 MiB

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// badRequestError is a malformed request: invalid JSON, wrong types, or bad
// query parameters.
type badRequestError struct{ msg string }

func (e *badRequestError) Error() string { return e.msg }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError maps domain errors to HTTP responses. Unknown errors are logged
// and hidden from the client, since they may contain internal details.
func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	detail := errorDetail{RequestID: requestIDFrom(r.Context())}
	var (
		status int
		ve     *transaction.ValidationError
		bre    *badRequestError
	)
	switch {
	case errors.As(err, &bre):
		status, detail.Code, detail.Message = http.StatusBadRequest, "bad_request", bre.msg
	case errors.As(err, &ve):
		status, detail.Code, detail.Message = http.StatusUnprocessableEntity, "validation_failed", "request has invalid fields"
		detail.Fields = ve.Fields
	case errors.Is(err, transaction.ErrNotFound):
		status, detail.Code, detail.Message = http.StatusNotFound, "not_found", err.Error()
	case errors.Is(err, transaction.ErrConflict):
		status, detail.Code, detail.Message = http.StatusConflict, "conflict", err.Error()
	case errors.Is(err, context.Canceled):
		// The client went away; nobody will read the response.
		return
	default:
		s.log.ErrorContext(r.Context(), "request failed", "error", err, "request_id", detail.RequestID)
		status, detail.Code, detail.Message = http.StatusInternalServerError, "internal", "internal server error"
	}
	writeJSON(w, status, errorBody{Error: detail})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return &badRequestError{fmt.Sprintf("request body must not exceed %d bytes", maxBodyBytes)}
		}
		return &badRequestError{"invalid JSON body: " + err.Error()}
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return &badRequestError{"request body must contain a single JSON object"}
	}
	return nil
}
