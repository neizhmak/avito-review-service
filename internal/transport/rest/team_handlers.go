package rest

import (
	"encoding/json"
	"net/http"

	"github.com/neizhmak/avito-review-service/internal/domain"
)

type createTeamRequest struct {
	TeamName string        `json:"team_name"`
	Members  []domain.User `json:"members,omitempty"`
}

type deactivateTeamRequest struct {
	TeamName string `json:"team_name"`
}

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	if req.TeamName == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "team_name is required")
		return
	}

	createdTeam, err := h.service.CreateTeam(r.Context(), domain.Team{
		Name:    req.TeamName,
		Members: req.Members,
	})
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"team": createdTeam,
	})
}

func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, http.StatusBadRequest, "ERROR", "team_name is required")
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusOK, team)
}

func (h *Handler) deactivateTeam(w http.ResponseWriter, r *http.Request) {
	var req deactivateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "ERROR", "invalid json")
		return
	}

	if err := h.service.DeactivateTeam(r.Context(), req.TeamName); err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}
