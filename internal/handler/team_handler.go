package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.writeBadRequest(w, r, "invalid json request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeBadRequest(w, r, err.Error())
		return
	}

	teamModel, userModels := ConvertCreateTeamDTOToModels(req)

	createdTeam, createdMembers, err := h.teamService.Create(r.Context(), teamModel, userModels)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	response := ConvertTeamModelsToDTO(*createdTeam, createdMembers)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]any{"team": response})
}

func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.writeBadRequest(w, r, "missing required query parameter: team_name")
		return
	}

	team, members, err := h.teamService.Get(r.Context(), teamName)
	if err != nil {
		h.WriteError(w, r, err)
		return
	}

	response := ConvertTeamModelsToDTO(*team, members)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
