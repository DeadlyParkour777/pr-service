package store

import (
	"context"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTestData(ctx context.Context, t *testing.T) {
	truncateTables(ctx)

	team := model.Team{Name: "user-test-team"}
	users := []model.User{
		{ID: "active-user-1", Username: "Alice", IsActive: true},
		{ID: "active-user-2", Username: "Charlie", IsActive: true},
		{ID: "inactive-user", Username: "Bob", IsActive: false},
	}

	_, err := testStore.Team().AddTeamWithMembers(ctx, team, users)
	require.NoError(t, err, "Failed to set up user test data")
}

func TestUserStore_Integration_GetByID(t *testing.T) {
	ctx := context.Background()
	setupUserTestData(ctx, t)

	s := testStore.User()

	user, err := s.GetByID(ctx, "active-user-1")
	require.NoError(t, err)

	require.NotNil(t, user)
	assert.Equal(t, "active-user-1", user.ID)
	assert.Equal(t, "Alice", user.Username)
	assert.True(t, user.IsActive)
	assert.Equal(t, "user-test-team", user.TeamName)
}

func TestUserStore_Integration_SetIsActive(t *testing.T) {
	ctx := context.Background()
	setupUserTestData(ctx, t)

	s := testStore.User()

	initialUser, err := s.GetByID(ctx, "active-user-1")
	require.NoError(t, err)
	require.True(t, initialUser.IsActive)

	updatedUser, err := s.SetIsActive(ctx, "active-user-1", false)
	require.NoError(t, err)

	require.NotNil(t, updatedUser)
	assert.False(t, updatedUser.IsActive)

	finalUser, err := s.GetByID(ctx, "active-user-1")
	require.NoError(t, err)
	assert.False(t, finalUser.IsActive)
}

func TestUserStore_Integration_GetActiveTeamMembers(t *testing.T) {
	ctx := context.Background()
	setupUserTestData(ctx, t)

	s := testStore.User()

	team, _, err := testStore.Team().GetByName(ctx, "user-test-team")
	require.NoError(t, err)

	members, err := s.GetActiveTeamMembers(ctx, team.ID, "active-user-1")
	require.NoError(t, err)

	require.Len(t, members, 1)
	assert.Equal(t, "active-user-2", members[0].ID)
}

func TestUserStore_Integration_GetActiveTeamMembers_AllActive(t *testing.T) {
	ctx := context.Background()
	truncateTables(ctx)

	s := testStore.User()

	team := model.Team{Name: "all-active-team"}
	users := []model.User{
		{ID: "u1", IsActive: true},
		{ID: "u2", IsActive: true},
		{ID: "u3", IsActive: true},
	}
	createdTeam, err := testStore.Team().AddTeamWithMembers(ctx, team, users)
	require.NoError(t, err)

	members, err := s.GetActiveTeamMembers(ctx, createdTeam.ID, "u1")
	require.NoError(t, err)

	require.Len(t, members, 2)
	memberIDs := []string{members[0].ID, members[1].ID}
	assert.ElementsMatch(t, []string{"u2", "u3"}, memberIDs)
}

func TestUserStore_Integration_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	setupUserTestData(ctx, t)

	s := testStore.User()

	_, err := s.GetByID(ctx, "non-existent-user")

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}
