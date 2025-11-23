package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/neizhmak/avito-review-service/internal/domain"
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
func respondError(w http.ResponseWriter, status int, message string) {
	type errorBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	type errorResponse struct {
		Error errorBody `json:"error"`
	}

	resp := errorResponse{
		Error: errorBody{Code: "ERROR", Message: message},
	}
	respondJSON(w, status, resp)
}

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	// Decode JSON body
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate input
	if team.Name == "" {
		respondError(w, http.StatusBadRequest, "team_name is required")
		return
	}

	// Create team via service
	createdTeam, err := h.service.CreateTeam(r.Context(), team)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Respond with created team
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"team": createdTeam,
	})
}

// createPR handles the HTTP request to create a new pull request.
func (h *Handler) createPR(w http.ResponseWriter, r *http.Request) {
	var pr domain.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	pr.Status = "OPEN"

	createdPR, err := h.service.Create(r.Context(), pr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
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
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	mergedPR, err := h.service.Merge(r.Context(), req.PRID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr": mergedPR,
	})
}

// ReassignReviewer handles the HTTP request to reassign a reviewer on a pull request.
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
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	targetID := temp.OldUserID
	if targetID == "" {
		targetID = temp.OldReviewerID
	}
	req.PRID = temp.PRID
	req.OldUserID = targetID

	newReviewerID, err := h.service.Reassign(r.Context(), req.PRID, req.OldUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pr, err := h.service.GetPR(r.Context(), req.PRID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch updated pr")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}

// getTeam handles the HTTP request to retrieve a team by its name.
func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, http.StatusBadRequest, "team_name is required")
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, team)
}

// setUserActive handles the HTTP request to set a user's active status.
func (h *Handler) setUserActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updatedUser, err := h.service.SetUserActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": updatedUser,
	})
}

// getUserReviews handles the HTTP request to retrieve pull requests assigned to a user for review.
func (h *Handler) getUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	prs, err := h.service.GetUserReviews(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Ensure prs is not nil
	if prs == nil {
		prs = []domain.PullRequest{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

// deactivateTeam handles the HTTP request to deactivate all users in a team.
func (h *Handler) deactivateTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string `json:"team_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.service.DeactivateTeam(r.Context(), req.TeamName); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

// getStats handles the HTTP request to retrieve system statistics.
func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, stats)
}
