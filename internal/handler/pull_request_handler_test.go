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

	appService := service.NewService(service.Dependencies{TeamRepo: testStore.Team(), UserRepo: testStore.User(), PRRepo: testStore.PR(), StatsRepo: testStore.PR()})
	teamModel := model.Team{Name: "pr-team"}
	users := []model.User{
		{ID: "pr-author", Username: "PR Author", IsActive: true},
		{ID: "pr-reviewer", Username: "PR Reviewer", IsActive: true},
	}
	_, _, err := appService.Team.Create(ctx, teamModel, users)
	require.NoError(t, err)

	token := getTestToken(t, "pr-author")
	createBody := `{"pull_request_id": "pr-1", "pull_request_name": "My First PR", "author_id": "pr-author"}`

	req, err := http.NewRequest("POST", testServerURL+"/pullRequest/create", strings.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
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

	req, err = http.NewRequest("POST", testServerURL+"/pullRequest/create", strings.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestPullRequestHandler_E2E_Merge(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	appService := service.NewService(service.Dependencies{TeamRepo: testStore.Team(), UserRepo: testStore.User(), PRRepo: testStore.PR(), StatsRepo: testStore.PR()})
	teamModel := model.Team{Name: "merge-team"}
	userModel := model.User{ID: "merge-author", Username: "Merge Author", IsActive: true}
	_, _, err := appService.Team.Create(ctx, teamModel, []model.User{userModel})
	require.NoError(t, err)

	pr := model.PullRequest{ID: "pr-to-merge", Name: "Test Merge", AuthorID: "merge-author", Status: model.StatusOpen}
	_, err = appService.PR.Create(ctx, pr)
	require.NoError(t, err)

	token := getTestToken(t, "merge-author")
	mergeBody := `{"pull_request_id": "pr-to-merge"}`

	req, err := http.NewRequest("POST", testServerURL+"/pullRequest/merge", strings.NewReader(mergeBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
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
	req, err = http.NewRequest("POST", testServerURL+"/pullRequest/merge", strings.NewReader(mergeBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPullRequestHandler_E2E_Reassign(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	appService := service.NewService(service.Dependencies{
		TeamRepo:  testStore.Team(),
		UserRepo:  testStore.User(),
		PRRepo:    testStore.PR(),
		StatsRepo: testStore.PR(),
	})
	teamModel := model.Team{Name: "reassign-team"}
	users := []model.User{
		{ID: "reassign-author", Username: "Reassign Author", IsActive: true},
		{ID: "reviewer-A", Username: "Reviewer A", IsActive: true},
		{ID: "reviewer-B", Username: "Reviewer B", IsActive: true},
		{ID: "candidate-C", Username: "Candidate C", IsActive: true},
	}
	_, _, err := appService.Team.Create(ctx, teamModel, users)
	require.NoError(t, err)

	prModel := model.PullRequest{ID: "pr-to-reassign", Name: "Test Reassign", AuthorID: "reassign-author"}
	createdPR, err := appService.PR.Create(ctx, prModel)
	require.NoError(t, err)
	require.Len(t, createdPR.AssignedReviewers, 2)

	userToReassign := createdPR.AssignedReviewers[0]
	stableReviewer := createdPR.AssignedReviewers[1]

	token := getTestToken(t, "reassign-author")
	reassignBody := `{"pull_request_id": "pr-to-reassign", "old_user_id": "` + userToReassign + `"}`

	req, err := http.NewRequest("POST", testServerURL+"/pullRequest/reassign", strings.NewReader(reassignBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var reassignResp struct {
		PR         PullRequestResponse `json:"pr"`
		ReplacedBy string              `json:"replaced_by"`
	}
	err = json.NewDecoder(resp.Body).Decode(&reassignResp)
	require.NoError(t, err)

	newReviewer := reassignResp.ReplacedBy
	assert.NotEqual(t, "reassign-author", newReviewer)
	assert.NotEqual(t, userToReassign, newReviewer)
	assert.NotEqual(t, stableReviewer, newReviewer)
	assert.NotContains(t, reassignResp.PR.AssignedReviewers, userToReassign)
	assert.Contains(t, reassignResp.PR.AssignedReviewers, newReviewer)
	assert.Contains(t, reassignResp.PR.AssignedReviewers, stableReviewer)
	assert.Len(t, reassignResp.PR.AssignedReviewers, 2)

	reassignBody = `{"pull_request_id": "non-existent-pr", "old_user_id": "` + userToReassign + `"}`
	req, err = http.NewRequest("POST", testServerURL+"/pullRequest/reassign", strings.NewReader(reassignBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	reassignBody = `{"pull_request_id": "pr-to-reassign", "old_user_id": "reassign-author"}`
	req, err = http.NewRequest("POST", testServerURL+"/pullRequest/reassign", strings.NewReader(reassignBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}
