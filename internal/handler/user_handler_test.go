package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/DeadlyParkour777/pr-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_E2E_SetUserIsActive(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	appService := service.NewService(service.Dependencies{TeamRepo: testStore.Team(), UserRepo: testStore.User(), PRRepo: testStore.PR()})
	teamModel := model.Team{Name: "e2e-user-team"}
	userModels := []model.User{{ID: "e2e-user-active", Username: "E2E User", IsActive: true}}
	_, _, err := appService.Team.Create(ctx, teamModel, userModels)
	require.NoError(t, err)

	setInactiveBody := `{"user_id": "e2e-user-active", "is_active": false}`
	req, err := http.NewRequest("POST", testServerURL+"/users/setIsActive", strings.NewReader(setInactiveBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp struct {
		User UserResponse `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	require.NoError(t, err)
	assert.Equal(t, "e2e-user-active", userResp.User.UserID)
	assert.False(t, userResp.User.IsActive)

	setNotFoundBody := `{"user_id": "non-existent-user", "is_active": false}`
	req, err = http.NewRequest("POST", testServerURL+"/users/setIsActive", strings.NewReader(setNotFoundBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var errResp APIErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)
}

func TestUserHandler_E2E_SetUserIsActive_ValidationFailure(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	invalidJsonBody := `{"user_id": "some-user",`
	req, err := http.NewRequest("POST", testServerURL+"/users/setIsActive", strings.NewReader(invalidJsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	missingFieldBody := `{"is_active": true}`
	req, err = http.NewRequest("POST", testServerURL+"/users/setIsActive", strings.NewReader(missingFieldBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var errResp APIErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", errResp.Error.Code)
}
