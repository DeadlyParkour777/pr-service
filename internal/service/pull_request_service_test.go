package service

import (
	"context"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/DeadlyParkour777/pr-service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_Reassign_FailsIfPRIsMerged(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	mergedPR := &model.PullRequest{
		ID:     "pr-1",
		Status: model.StatusMerged,
	}
	mockPRRepo.On("GetByID", context.Background(), "pr-1").Return(mergedPR, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	_, _, err := prService.Reassign(context.Background(), "pr-1", "old-reviewer-id")

	assert.Error(t, err)
	assert.Equal(t, ErrPRMerged, err)
}

func TestPullRequestService_Reassign_FailsIfReviewerNotAssigned(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	openPR := &model.PullRequest{
		ID:                "pr-1",
		Status:            model.StatusOpen,
		AssignedReviewers: []string{"user-B"},
	}

	mockPRRepo.On("GetByID", context.Background(), "pr-1").Return(openPR, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	_, _, err := prService.Reassign(context.Background(), "pr-1", "user-A")

	assert.Error(t, err)
	assert.Equal(t, ErrNotAssigned, err)
}

func TestPullRequestService_Create_Success(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	author := &model.FullUserInfo{User: model.User{ID: "author-1", TeamID: 123}}
	prToCreate := model.PullRequest{ID: "pr-1", AuthorID: "author-1"}

	candidates := []model.User{
		{ID: "user-A", TeamID: 123, IsActive: true},
		{ID: "user-B", TeamID: 123, IsActive: true},
		{ID: "user-C", TeamID: 123, IsActive: true},
	}

	mockUserRepo.On("GetByID", context.Background(), "author-1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembers", context.Background(), author.TeamID, author.ID).Return(candidates, nil)

	mockPRRepo.On("Create", context.Background(), mock.AnythingOfType("model.PullRequest")).Return(nil)

	finalPR := &model.PullRequest{
		ID:                "pr-1",
		AuthorID:          "author-1",
		AssignedReviewers: []string{"user-A", "user-C"},
	}
	mockPRRepo.On("GetByID", context.Background(), "pr-1").Return(finalPR, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	createdPR, err := prService.Create(context.Background(), prToCreate)

	assert.NoError(t, err)
	assert.NotNil(t, createdPR)
	assert.Len(t, createdPR.AssignedReviewers, 2)

	assert.Contains(t, []string{"user-A", "user-B", "user-C"}, createdPR.AssignedReviewers[0])
}

func TestPullRequestService_Create_AssignsOneReviewerIfOnlyOneCandidate(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	author := &model.FullUserInfo{User: model.User{ID: "author-1", TeamID: 123}}
	prToCreate := model.PullRequest{ID: "pr-1", AuthorID: "author-1"}

	candidates := []model.User{
		{ID: "user-A", TeamID: 123, IsActive: true},
	}

	mockUserRepo.On("GetByID", context.Background(), "author-1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembers", context.Background(), author.TeamID, author.ID).Return(candidates, nil)

	mockPRRepo.On("Create", context.Background(), mock.MatchedBy(func(pr model.PullRequest) bool {
		return len(pr.AssignedReviewers) == 1 && pr.AssignedReviewers[0] == "user-A"
	})).Return(nil)

	finalPR := &model.PullRequest{ID: "pr-1", AuthorID: "author-1", AssignedReviewers: []string{"user-A"}}
	mockPRRepo.On("GetByID", context.Background(), "pr-1").Return(finalPR, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	createdPR, err := prService.Create(context.Background(), prToCreate)

	assert.NoError(t, err)
	assert.NotNil(t, createdPR)
	assert.Len(t, createdPR.AssignedReviewers, 1)
	assert.Equal(t, "user-A", createdPR.AssignedReviewers[0])
}

func TestPullRequestService_Create_AssignsZeroReviewersIfNoCandidates(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	author := &model.FullUserInfo{User: model.User{ID: "author-1", TeamID: 123}}
	prToCreate := model.PullRequest{ID: "pr-1", AuthorID: "author-1"}

	candidates := []model.User{}

	mockUserRepo.On("GetByID", context.Background(), "author-1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembers", context.Background(), author.TeamID, author.ID).Return(candidates, nil)

	mockPRRepo.On("Create", context.Background(), mock.MatchedBy(func(pr model.PullRequest) bool {
		return len(pr.AssignedReviewers) == 0
	})).Return(nil)

	finalPR := &model.PullRequest{ID: "pr-1", AuthorID: "author-1", AssignedReviewers: []string{}}
	mockPRRepo.On("GetByID", context.Background(), "pr-1").Return(finalPR, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	createdPR, err := prService.Create(context.Background(), prToCreate)

	assert.NoError(t, err)
	assert.NotNil(t, createdPR)
	assert.Empty(t, createdPR.AssignedReviewers)
}

func TestPullRequestService_Reassign_FailsIfNoCandidatesAvailable(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	openPR := &model.PullRequest{
		ID:                "pr-1",
		AuthorID:          "author-1",
		Status:            model.StatusOpen,
		AssignedReviewers: []string{"another-reviewer", "old-reviewer"},
	}
	oldReviewer := &model.FullUserInfo{User: model.User{ID: "old-reviewer", TeamID: 123}}

	mockPRRepo.On("GetByID", context.Background(), "pr-1").Return(openPR, nil)
	mockUserRepo.On("GetByID", context.Background(), "old-reviewer").Return(oldReviewer, nil)

	mockUserRepo.On("GetActiveTeamMembers", context.Background(), oldReviewer.TeamID, "").Return([]model.User{}, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	_, _, err := prService.Reassign(context.Background(), "pr-1", "old-reviewer")

	assert.Error(t, err)
	assert.Equal(t, ErrNoCandidates, err)
}

func TestPullRequestService_Merge_Success(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	prID := "pr-1"

	openPR := &model.PullRequest{ID: prID, Status: model.StatusOpen}
	mergedPR := &model.PullRequest{ID: prID, Status: model.StatusMerged}

	mockPRRepo.On("GetByID", context.Background(), prID).Return(openPR, nil).Once()
	mockPRRepo.On("Merge", context.Background(), prID).Return(nil)
	mockPRRepo.On("GetByID", context.Background(), prID).Return(mergedPR, nil).Once()

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	resultPR, err := prService.Merge(context.Background(), prID)
	assert.NoError(t, err)
	assert.NotNil(t, resultPR)
	assert.Equal(t, model.StatusMerged, resultPR.Status)

	mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_Merge_IsIdempotent(t *testing.T) {
	mockPRRepo := mocks.NewPullRequestRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)

	prID := "pr-1"

	mergedPR := &model.PullRequest{ID: prID, Status: model.StatusMerged}

	mockPRRepo.On("GetByID", context.Background(), prID).Return(mergedPR, nil)

	prService := NewPullRequestService(mockPRRepo, mockUserRepo)

	resultPR, err := prService.Merge(context.Background(), prID)

	assert.NoError(t, err)
	assert.NotNil(t, resultPR)
	assert.Equal(t, model.StatusMerged, resultPR.Status)

	mockPRRepo.AssertNotCalled(t, "Merge", mock.Anything, mock.Anything)

	mockPRRepo.AssertExpectations(t)
}
