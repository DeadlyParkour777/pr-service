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

	if err := h.validate.Struct(req); err != nil {
		h.writeBadRequest(w, r, err.Error())
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

	if err := h.validate.Struct(req); err != nil {
		h.writeBadRequest(w, r, err.Error())
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

func (h *Handler) reassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignReviewerRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.writeBadRequest(w, r, "invalid json request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeBadRequest(w, r, err.Error())
		return
	}

	updatedPR, newReviewerID, err := h.prService.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	response := ConvertPRModelToDTO(*updatedPR)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]any{
		"pr":          response,
		"replaced_by": newReviewerID,
	})
}
