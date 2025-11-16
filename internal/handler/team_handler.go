package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.WriteError(w, r, err)
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
		errResponse := APIErrorResponse{
			Error: struct {
				Code    string "json:\"code\""
				Message string "json:\"message\""
			}{
				Code:    "BAD_REQUEST",
				Message: "missing required query parameter: team_name",
			},
		}
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, errResponse)
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
