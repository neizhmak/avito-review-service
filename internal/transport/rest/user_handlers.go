package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

type setUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func (h *Handler) setUserActive(w http.ResponseWriter, r *http.Request) {
	var req setUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	if strings.TrimSpace(req.UserID) == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "user_id is required")
		return
	}

	updatedUser, err := h.service.SetUserActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": updatedUser,
	})
}

func (h *Handler) getUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "user_id is required")
		return
	}

	prs, err := h.service.GetUserReviews(r.Context(), userID)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	if prs == nil {
		prs = []domain.PullRequestShort{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
