package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

type createPRRequest struct {
	ID       string `json:"pull_request_id"`
	Title    string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type mergePRRequest struct {
	PRID string `json:"pull_request_id"`
}

type reassignPRRequest struct {
	PRID      string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}

// createPR handles the HTTP request to create a new pull request.
func (h *Handler) createPR(w http.ResponseWriter, r *http.Request) {
	var req createPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	if strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.AuthorID) == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "pull_request_id, pull_request_name and author_id are required")
		return
	}

	createdPR, err := h.service.Create(r.Context(), domain.PullRequest{
		ID:       req.ID,
		Title:    req.Title,
		AuthorID: req.AuthorID,
	})
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
	var req mergePRRequest
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
	var req reassignPRRequest
	var temp struct {
		PRID          string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
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
