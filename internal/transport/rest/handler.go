package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/neizhmak/avito-review-service/internal/service"
)

type Handler struct {
	service *service.PRService
}

// NewHandler creates a new REST handler with the given PR service.
func NewHandler(service *service.PRService) *Handler {
	return &Handler{service: service}
}

// InitRouter initializes the HTTP router with routes and middleware.
func (h *Handler) InitRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Post("/team/add", h.createTeam)
	r.Post("/team/deactivate", h.deactivateTeam)
	r.Get("/team/get", h.getTeam)
	r.Post("/users/setIsActive", h.setUserActive)
	r.Get("/users/getReview", h.getUserReviews)
	r.Post("/pullRequest/create", h.createPR)
	r.Post("/pullRequest/merge", h.mergePR)
	r.Post("/pullRequest/reassign", h.reassignReviewer)
	r.Get("/health/stats", h.getStats)

	return r
}

// respondJSON writes a JSON response with the given status code and payload.
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// respondError writes a JSON error response with the given status code and message.
func respondError(w http.ResponseWriter, status int, code string, message string) {
	type errorBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	type errorResponse struct {
		Error errorBody `json:"error"`
	}

	resp := errorResponse{
		Error: errorBody{Code: code, Message: message},
	}
	slog.Error("request failed", "status", status, "code", code, "message", message)
	respondJSON(w, status, resp)
}

func mapError(err error) (int, string, string) {
	var svcErr *service.ServiceError
	if errors.As(err, &svcErr) {
		switch svcErr.Code {
		case service.ErrCodeNotFound:
			return http.StatusNotFound, svcErr.Code, svcErr.Msg
		case service.ErrCodeTeamExists:
			return http.StatusBadRequest, svcErr.Code, svcErr.Msg
		case service.ErrCodePRExists, service.ErrCodePRMerged, service.ErrCodeNotAssigned, service.ErrCodeNoCandidate:
			return http.StatusConflict, svcErr.Code, svcErr.Msg
		default:
			slog.Error("unexpected service error", "error", err)
			return http.StatusInternalServerError, "ERROR", "internal error"
		}
	}

	slog.Error("unexpected error", "error", err)
	return http.StatusInternalServerError, "ERROR", "internal error"
}
