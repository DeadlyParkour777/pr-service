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

func TestPullRequestHandler_E2E_Create(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	appService := service.NewService(service.Dependencies{TeamRepo: testStore.Team(), UserRepo: testStore.User(), PRRepo: testStore.PR()})
	teamModel := model.Team{Name: "pr-team"}
	users := []model.User{
		{ID: "pr-author", Username: "PR Author", IsActive: true},
		{ID: "pr-reviewer", Username: "PR Reviewer", IsActive: true},
	}
	_, _, err := appService.Team.Create(ctx, teamModel, users)
	require.NoError(t, err)

	createBody := `{"pull_request_id": "pr-1", "pull_request_name": "My First PR", "author_id": "pr-author"}`
	resp, err := http.Post(testServerURL+"/pullRequest/create", "application/json", strings.NewReader(createBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var createResp struct {
		PR PullRequestResponse `json:"pr"`
	}
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(t, err)
	assert.Equal(t, "pr-1", createResp.PR.PullRequestID)
	assert.NotEmpty(t, createResp.PR.AssignedReviewers)

	resp, err = http.Post(testServerURL+"/pullRequest/create", "application/json", strings.NewReader(createBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestPullRequestHandler_E2E_Merge(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	appService := service.NewService(service.Dependencies{TeamRepo: testStore.Team(), UserRepo: testStore.User(), PRRepo: testStore.PR()})
	teamModel := model.Team{Name: "merge-team"}
	userModel := model.User{ID: "merge-author", Username: "Merge Author", IsActive: true}
	_, _, err := appService.Team.Create(ctx, teamModel, []model.User{userModel})
	require.NoError(t, err)

	pr := model.PullRequest{ID: "pr-to-merge", Name: "Test Merge", AuthorID: "merge-author", Status: model.StatusOpen}
	_, err = appService.PR.Create(ctx, pr)
	require.NoError(t, err)

	mergeBody := `{"pull_request_id": "pr-to-merge"}`
	resp, err := http.Post(testServerURL+"/pullRequest/merge", "application/json", strings.NewReader(mergeBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var mergeResp struct {
		PR PullRequestResponse `json:"pr"`
	}
	err = json.NewDecoder(resp.Body).Decode(&mergeResp)
	require.NoError(t, err)
	assert.Equal(t, "MERGED", mergeResp.PR.Status)

	mergeBody = `{"pull_request_id": "non-existent-pr"}`
	resp, err = http.Post(testServerURL+"/pullRequest/merge", "application/json", strings.NewReader(mergeBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
