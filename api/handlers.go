package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"PolicyService/internal/evaluate"
	"PolicyService/internal/logger"
	"PolicyService/internal/model"
)

func RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/health", healthHandler)
	mux.HandleFunc("/v1/evaluate", evaluateHandler)
	mux.HandleFunc("/v1/test", testHandler)
	mux.HandleFunc("/v1/pel/validate", pelValidateHandler)
}

func requirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		jsonErr(w, "method not allowed — use POST", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost:
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, map[string]string{"status": "ok", "service": "policy-service"})
	default:
		jsonErr(w, "method not allowed — use GET or POST", http.StatusMethodNotAllowed)
	}
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}
	var req model.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.EventCode) == "" {
		jsonErr(w, "event_code is required", http.StatusBadRequest)
		return
	}
	logger.Info("evaluate: event=%s policies=%d", req.EventCode, len(req.Policies))
	resp := evaluate.Run(req)
	writeJSON(w, resp)
}

// testHandler is identical to evaluate for now; main API skips execution_log writes for harness.
func testHandler(w http.ResponseWriter, r *http.Request) {
	evaluateHandler(w, r)
}

func pelValidateHandler(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}
	var req model.PelValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request body", http.StatusBadRequest)
		return
	}
	expr := strings.TrimSpace(req.Expression)
	if expr == "" {
		writeJSON(w, model.PelValidateResponse{Success: true, Valid: false, Message: "expression is empty"})
		return
	}
	ok, msg := evaluate.ValidatePEL(expr)
	writeJSON(w, model.PelValidateResponse{Success: true, Valid: ok, Message: msg})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
