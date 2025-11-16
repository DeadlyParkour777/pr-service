package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

func (h *Handler) setUserIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.writeBadRequest(w, r, "invalid json request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeBadRequest(w, r, err.Error())
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	response := ConvertFullUserModelToDTO(*user)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{"user": response})
}

func (h *Handler) getReviewsForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeBadRequest(w, r, "missing required query parameter: user_id")
		return
	}

	prs, err := h.userService.GetReviewsForUser(r.Context(), userID)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	prDTOs := make([]PullRequestResponse, len(prs))
	for i, pr := range prs {
		prDTOs[i] = ConvertPRModelToDTO(pr)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{"user_id": userID, "pull_requests": prDTOs})
}
