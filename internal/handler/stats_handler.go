package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

func (h *Handler) getUserStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetUserStats(r.Context())
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{"user_stats": stats})
}
