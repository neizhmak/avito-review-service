package rest

import "net/http"

// getStats handles the HTTP request to retrieve system statistics.
func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		status, code, msg := mapError(err)
		respondError(w, status, code, msg)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}
