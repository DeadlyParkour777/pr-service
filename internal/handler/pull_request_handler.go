package handler

import (
	"net/http"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/go-chi/render"
)

func (h *Handler) createPullRequest(w http.ResponseWriter, r *http.Request) {
	var req CreatePullRequestRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.WriteError(w, r, err)
		return
	}

	prModel := model.PullRequest{
		ID:       req.PullRequestID,
		Name:     req.PullRequestName,
		AuthorID: req.AuthorID,
	}

	createdPR, err := h.prService.Create(r.Context(), prModel)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	response := ConvertPRModelToDTO(*createdPR)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]any{"pr": response})
}

func (h *Handler) mergePullRequest(w http.ResponseWriter, r *http.Request) {
	var req MergePullRequestRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.WriteError(w, r, err)
		return
	}

	mergedPR, err := h.prService.Merge(r.Context(), req.PullRequestID)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	response := ConvertPRModelToDTO(*mergedPR)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{"pr": response})
}
