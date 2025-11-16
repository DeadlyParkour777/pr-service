package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamHandler_E2E_CreateAndGetTeam(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	createBody := `
	{
		"team_name": "e2e-team",
		"members": [
			{"user_id": "e2e-user-1", "username": "E2E Alice", "is_active": true}
		]
	}`

	createResp, err := http.Post(testServerURL+"/team/add", "application/json", strings.NewReader(createBody))
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode, "Expected status 201 Created")

	getResp, err := http.Get(testServerURL + "/team/get?team_name=e2e-team")
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	body, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	var teamResp TeamResponse
	err = json.Unmarshal(body, &teamResp)
	require.NoError(t, err)

	assert.Equal(t, "e2e-team", teamResp.TeamName)
	require.Len(t, teamResp.Members, 1)
	assert.Equal(t, "e2e-user-1", teamResp.Members[0].UserID)
}

func TestTeamHandler_E2E_CreateTeam_ValidationFailure(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	createBody := `{"members": []}`

	resp, err := http.Post(testServerURL+"/team/add", "application/json", strings.NewReader(createBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp APIErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}
