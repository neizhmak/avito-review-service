package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

// createPR handles the HTTP request to create a new pull request.
func (h *Handler) createPR(w http.ResponseWriter, r *http.Request) {
	var pr domain.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	if strings.TrimSpace(pr.ID) == "" || strings.TrimSpace(pr.Title) == "" || strings.TrimSpace(pr.AuthorID) == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "pull_request_id, pull_request_name and author_id are required")
		return
	}

	createdPR, err := h.service.Create(r.Context(), pr)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"pr": createdPR,
	})
}

// mergePR handles the HTTP request to merge a pull request.
func (h *Handler) mergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PRID string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	if strings.TrimSpace(req.PRID) == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "pull_request_id is required")
		return
	}

	mergedPR, err := h.service.Merge(r.Context(), req.PRID)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr": mergedPR,
	})
}

// reassignReviewer handles the HTTP request to reassign a reviewer on a pull request.
func (h *Handler) reassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PRID      string `json:"pull_request_id"`
		OldUserID string `json:"old_user_id"`
	}

	type Alias struct {
		PRID          string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
	var temp Alias
	if err := json.NewDecoder(r.Body).Decode(&temp); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	targetID := temp.OldUserID
	if targetID == "" {
		targetID = temp.OldReviewerID
	}
	req.PRID = temp.PRID
	req.OldUserID = targetID

	if strings.TrimSpace(req.PRID) == "" || strings.TrimSpace(req.OldUserID) == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "pull_request_id and old_user_id are required")
		return
	}

	newReviewerID, err := h.service.Reassign(r.Context(), req.PRID, req.OldUserID)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	pr, err := h.service.GetPR(r.Context(), req.PRID)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
