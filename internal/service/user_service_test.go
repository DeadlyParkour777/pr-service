package service

import (
	"context"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/DeadlyParkour777/pr-service/internal/store"
	"github.com/DeadlyParkour777/pr-service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_GetReviewsForUser_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockPRRepo := mocks.NewPullRequestRepository(t)

	userID := "user-1"
	user := &model.FullUserInfo{User: model.User{ID: userID}}
	expectedPRs := []model.PullRequest{{ID: "pr-1"}}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockPRRepo.On("GetByReviewerID", mock.Anything, userID).Return(expectedPRs, nil)

	userService := NewUserService(mockUserRepo, mockPRRepo)

	resultPRs, err := userService.GetReviewsForUser(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPRs, resultPRs)
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestUserService_GetReviewsForUser_FailsIfUserNotFound(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockPRRepo := mocks.NewPullRequestRepository(t)

	userID := "non-existent-user"

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, store.ErrNotFound)

	userService := NewUserService(mockUserRepo, mockPRRepo)

	_, err := userService.GetReviewsForUser(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	mockPRRepo.AssertNotCalled(t, "GetByReviewerID", mock.Anything, mock.Anything)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_Success(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockPRRepo := mocks.NewPullRequestRepository(t)

	userID := "user-1"
	statusToSet := false

	updatedUser := &model.FullUserInfo{
		User:     model.User{ID: userID, IsActive: statusToSet},
		TeamName: "backend",
	}

	mockUserRepo.On("SetIsActive", mock.Anything, userID, statusToSet).Return(updatedUser, nil)

	userService := NewUserService(mockUserRepo, mockPRRepo)

	resultUser, err := userService.SetIsActive(context.Background(), userID, statusToSet)

	assert.NoError(t, err)
	assert.Equal(t, updatedUser, resultUser)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_FailsIfNotFound(t *testing.T) {
	mockUserRepo := mocks.NewUserRepository(t)
	mockPRRepo := mocks.NewPullRequestRepository(t)

	userID := "non-existent-user"

	mockUserRepo.On("SetIsActive", mock.Anything, userID, true).Return(nil, store.ErrNotFound)

	userService := NewUserService(mockUserRepo, mockPRRepo)
	_, err := userService.SetIsActive(context.Background(), userID, true)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	mockUserRepo.AssertExpectations(t)
}
