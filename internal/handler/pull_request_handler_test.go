package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupE2ETestData(t *testing.T) {
	truncateTables(context.Background())

	createBody := `
	{
		"team_name": "e2e-pr-team",
		"members": [
			{"user_id": "author-e2e", "username": "E2E Author", "is_active": true},
			{"user_id": "reviewer-e2e-1", "username": "E2E Reviewer 1", "is_active": true},
			{"user_id": "reviewer-e2e-2", "username": "E2E Reviewer 2", "is_active": true},
			{"user_id": "new-reviewer-e2e", "username": "E2E New Reviewer", "is_active": true}
		]
	}`

	resp, err := http.Post(testServerURL+"/team/add", "application/json", strings.NewReader(createBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
}

func TestPullRequestHandler_E2E_FullLifecycle(t *testing.T) {
	setupE2ETestData(t)

	createPRBody := `{"pull_request_id": "pr-e2e-1", "pull_request_name": "E2E Test PR", "author_id": "author-e2e"}`

	createResp, err := http.Post(testServerURL+"/pullRequest/create", "application/json", strings.NewReader(createPRBody))
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode, "PR creation should succeed")

	var createRespBody struct {
		PR PullRequestResponse `json:"pr"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	assert.Equal(t, "pr-e2e-1", createRespBody.PR.PullRequestID)
	assert.Equal(t, model.StatusOpen, model.PRStatus(createRespBody.PR.Status))
	assert.Len(t, createRespBody.PR.AssignedReviewers, 2, "Should assign 2 reviewers")

	mergePRBody := `{"pull_request_id": "pr-e2e-1"}`

	mergeResp, err := http.Post(testServerURL+"/pullRequest/merge", "application/json", strings.NewReader(mergePRBody))
	require.NoError(t, err)
	defer mergeResp.Body.Close()

	assert.Equal(t, http.StatusOK, mergeResp.StatusCode, "PR merge should succeed")

	var mergeRespBody struct {
		PR PullRequestResponse `json:"pr"`
	}
	err = json.NewDecoder(mergeResp.Body).Decode(&mergeRespBody)
	require.NoError(t, err)

	assert.Equal(t, model.StatusMerged, model.PRStatus(mergeRespBody.PR.Status), "PR status should be MERGED")

	reassignBody := `{"pull_request_id": "pr-e2e-1", "old_user_id": "` + createRespBody.PR.AssignedReviewers[0] + `"}`

	reassignResp, err := http.Post(testServerURL+"/pullRequest/reassign", "application/json", strings.NewReader(reassignBody))
	require.NoError(t, err)
	defer reassignResp.Body.Close()

	assert.Equal(t, http.StatusConflict, reassignResp.StatusCode, "Should not be able to reassign on a merged PR")

	var errResp APIErrorResponse
	err = json.NewDecoder(reassignResp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "PR_MERGED", errResp.Error.Code)
}

func TestPullRequestHandler_E2E_ReassignSuccess(t *testing.T) {
	setupE2ETestData(t)

	createPRBody := `{"pull_request_id": "pr-e2e-reassign", "pull_request_name": "Reassign Test", "author_id": "author-e2e"}`
	createResp, err := http.Post(testServerURL+"/pullRequest/create", "application/json", strings.NewReader(createPRBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		PR PullRequestResponse `json:"pr"`
	}
	json.NewDecoder(createResp.Body).Decode(&createRespBody)
	createResp.Body.Close()

	oldReviewer := createRespBody.PR.AssignedReviewers[0]

	reassignBody := `{"pull_request_id": "pr-e2e-reassign", "old_user_id": "` + oldReviewer + `"}`

	reassignResp, err := http.Post(testServerURL+"/pullRequest/reassign", "application/json", strings.NewReader(reassignBody))
	require.NoError(t, err)
	defer reassignResp.Body.Close()

	assert.Equal(t, http.StatusOK, reassignResp.StatusCode, "Reassign should succeed")

	var reassignRespBody struct {
		PR         PullRequestResponse `json:"pr"`
		ReplacedBy string              `json:"replaced_by"`
	}
	err = json.NewDecoder(reassignResp.Body).Decode(&reassignRespBody)
	require.NoError(t, err)

	assert.NotContains(t, reassignRespBody.PR.AssignedReviewers, oldReviewer)
	assert.Contains(t, reassignRespBody.PR.AssignedReviewers, reassignRespBody.ReplacedBy)
	assert.NotEqual(t, oldReviewer, reassignRespBody.ReplacedBy)
}
