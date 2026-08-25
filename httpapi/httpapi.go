package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"membership13/domain"
	"membership13/query"
	"membership13/service"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler { return &Handler{service: svc} }

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/records", h.records)
	mux.HandleFunc("/records/", h.recordAction)
	return withHeaders(mux)
}

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := h.service.Health(); err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) records(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var input struct {
			ID          string `json:"id"`
			MemberID    int    `json:"member_id"`
			BenefitCode string `json:"benefit_code"`
		}
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record := domain.NewRecord(input.ID, input.MemberID, input.BenefitCode, time.Now())
		if err := h.service.RegisterRecord(record); err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, service.ErrDuplicateRecord) {
				status = http.StatusConflict
			}
			writeError(w, status, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	case http.MethodGet:
		memberID, _ := strconv.Atoi(r.URL.Query().Get("member"))
		filter := query.RecordFilter{MemberID: memberID, Status: r.URL.Query().Get("status"), BenefitCode: r.URL.Query().Get("benefit")}
		records, err := h.service.QueryRecords(filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, records)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) recordAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		writeError(w, http.StatusNotFound, "record id is required")
		return
	}
	id := parts[1]
	if len(parts) == 2 && r.Method == http.MethodGet {
		record, err := h.service.GetRecord(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, record)
		return
	}
	if len(parts) != 3 || r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "unsupported record action")
		return
	}
	switch parts[2] {
	case "process":
		record, err := h.service.ProcessRecord(id)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, record)
	case "archive":
		record, err := h.service.ArchiveRecord(id)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, record)
	default:
		writeError(w, http.StatusNotFound, "unknown action")
	}
}

func decodeJSON(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
