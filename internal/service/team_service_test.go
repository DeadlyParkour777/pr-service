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

func TestTeamService_Create_Success(t *testing.T) {
	mockTeamRepo := mocks.NewTeamRepository(t)

	teamToCreate := model.Team{Name: "backend"}
	membersToCreate := []model.User{{ID: "u1"}}

	createdTeamWithID := model.Team{ID: 1, Name: "backend"}
	mockTeamRepo.On("AddTeamWithMembers", mock.Anything, teamToCreate, membersToCreate).Return(&createdTeamWithID, nil)
	teamService := NewTeamService(mockTeamRepo)

	resultTeam, resultMembers, err := teamService.Create(context.Background(), teamToCreate, membersToCreate)

	assert.NoError(t, err)
	assert.Equal(t, &createdTeamWithID, resultTeam)
	assert.Equal(t, membersToCreate, resultMembers)
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_Create_FailsIfTeamExists(t *testing.T) {
	mockTeamRepo := mocks.NewTeamRepository(t)
	teamToCreate := model.Team{Name: "backend"}
	mockTeamRepo.On("AddTeamWithMembers", mock.Anything, teamToCreate, mock.Anything).Return(nil, store.ErrTeamExists)

	teamService := NewTeamService(mockTeamRepo)

	_, _, err := teamService.Create(context.Background(), teamToCreate, nil)

	assert.Error(t, err)
	assert.Equal(t, ErrTeamExists, err)
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_Get_SuccessWithNoMembers(t *testing.T) {
	mockTeamRepo := mocks.NewTeamRepository(t)

	teamName := "lonely-team"
	teamFromStore := &model.Team{ID: 2, Name: teamName}
	membersFromStore := []model.User{}

	mockTeamRepo.On("GetByName", mock.Anything, teamName).Return(teamFromStore, membersFromStore, nil)

	teamService := NewTeamService(mockTeamRepo)

	resultTeam, resultMembers, err := teamService.Get(context.Background(), teamName)

	assert.NoError(t, err)
	assert.Equal(t, teamFromStore, resultTeam)
	assert.NotNil(t, resultMembers)
	assert.Empty(t, resultMembers)
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_Get_FailsIfNotFound(t *testing.T) {
	mockTeamRepo := mocks.NewTeamRepository(t)
	teamName := "non-existent-team"
	mockTeamRepo.On("GetByName", mock.Anything, teamName).Return(nil, nil, store.ErrNotFound)

	teamService := NewTeamService(mockTeamRepo)

	_, _, err := teamService.Get(context.Background(), teamName)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_Get_SuccessWithMembers(t *testing.T) {
	mockTeamRepo := mocks.NewTeamRepository(t)

	teamName := "full-team"
	teamFromStore := &model.Team{ID: 3, Name: teamName}
	membersFromStore := []model.User{
		{ID: "u1", Username: "Alice"},
		{ID: "u2", Username: "Bob"},
	}

	mockTeamRepo.On("GetByName", mock.Anything, teamName).Return(teamFromStore, membersFromStore, nil)

	teamService := NewTeamService(mockTeamRepo)

	resultTeam, resultMembers, err := teamService.Get(context.Background(), teamName)

	assert.NoError(t, err)
	assert.Equal(t, teamFromStore, resultTeam)
	assert.Equal(t, membersFromStore, resultMembers)
	mockTeamRepo.AssertExpectations(t)
}
