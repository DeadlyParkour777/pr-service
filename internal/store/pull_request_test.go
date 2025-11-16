package store

import (
	"context"
	"testing"
	"time"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPRTestData(ctx context.Context, t *testing.T) {
	truncateTables(ctx)

	team := model.Team{Name: "test-team"}
	users := []model.User{
		{ID: "author-1", Username: "Author", IsActive: true},
		{ID: "reviewer-1", Username: "Reviewer 1", IsActive: true},
		{ID: "reviewer-2", Username: "Reviewer 2", IsActive: true},
		{ID: "new-reviewer", Username: "New Reviewer", IsActive: true},
	}

	_, err := testStore.Team().AddTeamWithMembers(ctx, team, users)
	require.NoError(t, err, "Failed to set up test data")
}

func TestPullRequestStore_Integration_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	prToCreate := model.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "author-1",
		AssignedReviewers: []string{"reviewer-1", "reviewer-2"},
	}

	err := s.Create(ctx, prToCreate)
	require.NoError(t, err, "Create should not return an error")

	fetchedPR, err := s.GetByID(ctx, "pr-1")
	require.NoError(t, err, "GetByID should not return an error")

	require.NotNil(t, fetchedPR)
	assert.Equal(t, prToCreate.ID, fetchedPR.ID)
	assert.Equal(t, prToCreate.Name, fetchedPR.Name)
	assert.Equal(t, prToCreate.AuthorID, fetchedPR.AuthorID)
	assert.Equal(t, model.StatusOpen, fetchedPR.Status)
	assert.ElementsMatch(t, prToCreate.AssignedReviewers, fetchedPR.AssignedReviewers)
	assert.NotZero(t, fetchedPR.CreatedAt)
}

func TestPullRequestStore_Integration_Merge(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	prToCreate := model.PullRequest{ID: "pr-to-merge", Name: "Merge Test", AuthorID: "author-1"}
	err := s.Create(ctx, prToCreate)
	require.NoError(t, err)

	err = s.Merge(ctx, "pr-to-merge")
	require.NoError(t, err)

	mergedPR, err := s.GetByID(ctx, "pr-to-merge")
	require.NoError(t, err)

	assert.Equal(t, model.StatusMerged, mergedPR.Status)
	require.NotNil(t, mergedPR.MergedAt)
	assert.WithinDuration(t, time.Now(), *mergedPR.MergedAt, 5*time.Second)
}

func TestPullRequestStore_Integration_Reassign(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	prToCreate := model.PullRequest{
		ID:                "pr-to-reassign",
		Name:              "Reassign Test",
		AuthorID:          "author-1",
		AssignedReviewers: []string{"reviewer-1"},
	}
	err := s.Create(ctx, prToCreate)
	require.NoError(t, err)

	err = s.ReassignReviewer(ctx, "pr-to-reassign", "reviewer-1", "new-reviewer")
	require.NoError(t, err)

	reassignedPR, err := s.GetByID(ctx, "pr-to-reassign")
	require.NoError(t, err)

	expectedReviewers := []string{"new-reviewer"}
	assert.Equal(t, expectedReviewers, reassignedPR.AssignedReviewers)
}

func TestPullRequestStore_Integration_Merge_NotFound(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	err := s.Merge(ctx, "non-existent-pr")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestPullRequestStore_Integration_Reassign_FailsIfOldReviewerNotAssigned(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	prToCreate := model.PullRequest{
		ID:                "pr-reassign-fail",
		AuthorID:          "author-1",
		AssignedReviewers: []string{"reviewer-1"},
	}
	err := s.Create(ctx, prToCreate)
	require.NoError(t, err)

	err = s.ReassignReviewer(ctx, "pr-reassign-fail", "reviewer-2", "new-reviewer")

	assert.Error(t, err)
}

func TestPullRequestStore_Integration_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	_, err := s.GetByID(ctx, "non-existent-pr")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestPullRequestStore_Integration_Create_FailsOnDuplicate(t *testing.T) {
	ctx := context.Background()
	setupPRTestData(ctx, t)

	s := testStore.PR()

	pr := model.PullRequest{ID: "duplicate-pr-id", AuthorID: "author-1"}
	err := s.Create(ctx, pr)
	require.NoError(t, err)

	err = s.Create(ctx, pr)

	assert.Error(t, err)
	assert.Equal(t, ErrPRExists, err)
}
