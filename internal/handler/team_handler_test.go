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

	token := getTestToken(t, "test-user")

	createBody := `
	{
		"team_name": "e2e-team",
		"members": [
			{"user_id": "e2e-user-1", "username": "E2E Alice", "is_active": true}
		]
	}`

	createReq, err := http.NewRequest("POST", testServerURL+"/team/add", strings.NewReader(createBody))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := http.DefaultClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode, "Expected status 201 Created")

	getReq, err := http.NewRequest("GET", testServerURL+"/team/get?team_name=e2e-team", nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+token)

	getResp, err := http.DefaultClient.Do(getReq)
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

	token := getTestToken(t, "test-user")
	createBody := `{"members": []}`

	req, err := http.NewRequest("POST", testServerURL+"/team/add", strings.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp APIErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}

func TestTeamHandler_E2E_CreateTeam_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	token := getTestToken(t, "test-user")

	createBody := `
	{
		"team_name": "duplicate-team",
		"members": [
			{"user_id": "user-1", "username": "Alice", "is_active": true}
		]
	}`

	req1, err := http.NewRequest("POST", testServerURL+"/team/add", strings.NewReader(createBody))
	require.NoError(t, err)
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token)

	resp1, err := http.DefaultClient.Do(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	req2, err := http.NewRequest("POST", testServerURL+"/team/add", strings.NewReader(createBody))
	require.NoError(t, err)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	var errResp APIErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "TEAM_EXISTS", errResp.Error.Code)
}

func TestTeamHandler_E2E_CreateTeam_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	token := getTestToken(t, "test-user")
	createBody := `{"team_name": "invalid-json",`

	req, err := http.NewRequest("POST", testServerURL+"/team/add", strings.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp APIErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Contains(t, errResp.Error.Message, "invalid json request")
}

func TestTeamHandler_E2E_GetTeam_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	token := getTestToken(t, "test-user")

	req, err := http.NewRequest("GET", testServerURL+"/team/get?team_name=non-existent-team", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	getResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)

	var errResp APIErrorResponse
	err = json.NewDecoder(getResp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestTeamHandler_E2E_GetTeam_MissingQueryParam(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	token := getTestToken(t, "test-user")

	req, err := http.NewRequest("GET", testServerURL+"/team/get", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	getResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, getResp.StatusCode)

	var errResp APIErrorResponse
	err = json.NewDecoder(getResp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
	assert.Contains(t, errResp.Error.Message, "missing required query parameter: team_name")
}
