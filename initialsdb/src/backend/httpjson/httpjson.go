package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// MaxBodySize protects against accidental or malicious large payloads.
// Tune per deployment.
const MaxBodySize = 1 << 20 // 1 MB

// ----------------------------------------------------------------------------
// Request helpers
// ----------------------------------------------------------------------------

// Decode decodes a JSON request body into dst.
// It enforces:
//   - application/json Content-Type (if present)
//   - body size limit
//   - single JSON object (no trailing garbage)
func Decode(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()

	if ct := r.Header.Get("Content-Type"); ct != "" {
		// Be permissive: allow charset, etc.
		if ct != "application/json" && !hasJSONPrefix(ct) {
			return errors.New("content-type must be application/json")
		}
	}

	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	// Ensure there is no trailing junk
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("unexpected trailing data in JSON body")
	}

	return nil
}

func hasJSONPrefix(ct string) bool {
	return len(ct) >= 16 && ct[:16] == "application/json"
}

// ----------------------------------------------------------------------------
// Response helpers
// ----------------------------------------------------------------------------

// Write writes a successful JSON response with the given status code.
func Write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

// WriteOK is shorthand for 200 OK with JSON body.
func WriteOK(w http.ResponseWriter, v any) {
	Write(w, http.StatusOK, v)
}

// WriteCreated is shorthand for 201 Created with JSON body.
func WriteCreated(w http.ResponseWriter, v any) {
	Write(w, http.StatusCreated, v)
}

// WriteNoContent writes a 204 response with no body.
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ----------------------------------------------------------------------------
// Error helpers
// ----------------------------------------------------------------------------

// APIErrorShape matches your contract exactly.
// Import your authcontract package instead of duplicating if preferred.
type APIErrorShape struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// WriteError writes a typed JSON error envelope.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var resp APIErrorShape
	resp.Error.Code = code
	resp.Error.Message = message

	_ = json.NewEncoder(w).Encode(resp)
}

// Convenience helpers for common cases

func BadRequest(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusBadRequest, code, message)
}

func Unauthorized(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusUnauthorized, code, message)
}

func Forbidden(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusForbidden, code, message)
}

func NotFound(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusNotFound, code, message)
}

func Conflict(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusConflict, code, message)
}

func TooManyRequests(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusTooManyRequests, code, message)
}

func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
