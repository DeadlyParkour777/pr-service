package store

import (
	"context"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamStore_Integration_AddTeamWithMembers_And_GetByName(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	s := testStore.Team()

	teamToCreate := model.Team{Name: "backend-team"}
	membersToCreate := []model.User{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: false},
	}

	createdTeam, err := s.AddTeamWithMembers(ctx, teamToCreate, membersToCreate)
	require.NoError(t, err, "AddTeamWithMembers should not return an error")

	require.NotZero(t, createdTeam.ID, "Created team should have a non-zero ID")

	fetchedTeam, fetchedMembers, err := s.GetByName(ctx, "backend-team")
	require.NoError(t, err, "GetByName should not return an error")

	require.NotNil(t, fetchedTeam)
	assert.Equal(t, createdTeam.ID, fetchedTeam.ID, "Team ID should match")
	assert.Equal(t, "backend-team", fetchedTeam.Name, "Team name should match")

	require.Len(t, fetchedMembers, 2, "Should fetch 2 members")

	for i := range membersToCreate {
		membersToCreate[i].TeamID = createdTeam.ID
	}
	assert.ElementsMatch(t, membersToCreate, fetchedMembers, "Fetched members should match created members")
}

func TestTeamStore_Integration_CreateFailsOnDuplicate(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	s := testStore.Team()

	team := model.Team{Name: "duplicate-team"}

	_, err := s.AddTeamWithMembers(ctx, team, nil)
	require.NoError(t, err)

	_, err = s.AddTeamWithMembers(ctx, team, nil)

	assert.Error(t, err)
	assert.Equal(t, ErrTeamExists, err)
}

func TestTeamStore_Integration_GetByName_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	s := testStore.Team()

	_, _, err := s.GetByName(ctx, "non-existent-team")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestTeamStore_Integration_AddTeamWithMembers_NoMembers(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	s := testStore.Team()

	teamToCreate := model.Team{Name: "empty-team"}

	_, err := s.AddTeamWithMembers(ctx, teamToCreate, nil)
	require.NoError(t, err)

	fetchedTeam, fetchedMembers, err := s.GetByName(ctx, "empty-team")
	require.NoError(t, err)

	assert.NotNil(t, fetchedTeam)
	assert.Equal(t, "empty-team", fetchedTeam.Name)
	assert.Empty(t, fetchedMembers, "Should have no members")
}
