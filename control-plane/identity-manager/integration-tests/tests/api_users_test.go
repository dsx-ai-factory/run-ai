package tests

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

type ctxKey string

const ctxKeyAuth ctxKey = "Authorization"

func TestGetUsers(t *testing.T) {
	ctx := newTestContext()

	initialUsers, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, initialUsers)
	initialCount := len(initialUsers)

	const numTestUsers = 3
	createdUserIDs := createTestUsers(t, ctx, numTestUsers)
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	users, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, users)

	finalCount := len(users)
	expectedCount := initialCount + numTestUsers
	require.Equal(t, expectedCount, finalCount, "Expected %d users (initial: %d + created: %d), but got %d", expectedCount, initialCount, numTestUsers, finalCount)
}

func TestGetUsers_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCreateLocalUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	reset := ptr.To(false)
	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(reset),
	}
	user, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, user)
	defer userCleanup(t, ctx, user.Id)

	require.Equal(t, email, user.Username)
	require.NotEmpty(t, user.TempPassword)
	require.NotEmpty(t, user.Id)
}

func TestCreateLocalUser_InvalidEmail(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	req := openapi.UserCreationRequest{
		Email:         "invalid-email",
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
	}
	_, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateLocalUser_DuplicateEmail(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
	}

	user, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer userCleanup(t, ctx, user.Id)

	_, resp, err = OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateLocalUser_EmptyEmail(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	req := openapi.UserCreationRequest{
		Email:         "",
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
	}
	_, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateLocalUser_XSSPayload(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	req := openapi.UserCreationRequest{
		Email:         "<script>alert(1)</script>@evil.com",
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
	}
	_, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateLocalUser_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := openapi.UserCreationRequest{
		Email:         GenerateUsername(),
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
	}
	_, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCountUsers(t *testing.T) {
	// t.Parallel()
	ctx := newTestContext()

	initialCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, initialCount)

	const numTestUsers = 2
	createdUserIDs := createTestUsers(t, ctx, numTestUsers)
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	finalCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, finalCount)

	countIncrease := finalCount.Count - initialCount.Count
	require.Equal(t, int64(numTestUsers), countIncrease, "Expected count to increase by %d users, but it increased by %d", numTestUsers, countIncrease)
}

func TestCountUsers_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetUsers_FilterBySearch(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	userIDs, usernames, _ := createTestUsersWithDetails(t, ctx, 3)
	defer cleanupTestUsers(t, ctx, userIDs)

	searchTerm := usernames[0][:5]
	users, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Search(searchTerm).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, users)

	found := false
	for _, user := range users {
		if user.Username == usernames[0] {
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find user %s in search results for term %s", usernames[0], searchTerm)
}

func TestGetUsers_FilterByUsername(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	userIDs, usernames, _ := createTestUsersWithDetails(t, ctx, 2)
	defer cleanupTestUsers(t, ctx, userIDs)

	filterBy := []string{"username=@" + usernames[0][:8]}
	users, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).FilterBy(filterBy).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, users)

	found := false
	for _, user := range users {
		if user.Username == usernames[0] {
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find user %s in filterBy results", usernames[0])
}

func TestGetUsers_FilterByIsLocal(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	userIDs, _, _ := createTestUsersWithDetails(t, ctx, 2)
	defer cleanupTestUsers(t, ctx, userIDs)

	filterBy := []string{"isLocal==true"}
	users, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).FilterBy(filterBy).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, users)

	foundCreatedUsers := 0
	for _, user := range users {
		for _, createdUserID := range userIDs {
			if user.Id == createdUserID {
				foundCreatedUsers++
				require.True(t, *user.IsLocal.Get(), "Expected user %s to be local", user.Username)
			}
		}
	}
	require.Equal(t, len(userIDs), foundCreatedUsers, "Expected to find all %d created users in isLocal filter results", len(userIDs))
}

func TestCountUsers_FilterBySearch(t *testing.T) {
	ctx := newTestContext()

	initialCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, initialCount)

	userIDs, usernames, _ := createTestUsersWithDetails(t, ctx, 3)
	defer cleanupTestUsers(t, ctx, userIDs)

	searchTerm := usernames[0][:5]
	searchCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).Search(searchTerm).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, searchCount)

	require.Greater(t, searchCount.Count, initialCount.Count, "Search count should be greater than initial count")
}

func TestCountUsers_FilterByIsLocal(t *testing.T) {
	ctx := newTestContext()

	initialIsLocalTrueCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).FilterBy([]string{"isLocal==true"}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, initialIsLocalTrueCount)

	initialIsLocalFalseCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).FilterBy([]string{"isLocal==false"}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, initialIsLocalFalseCount)

	const numTestUsers = 2
	userIDs, _, _ := createTestUsersWithDetails(t, ctx, numTestUsers)
	defer cleanupTestUsers(t, ctx, userIDs)

	finalIsLocalTrueCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).FilterBy([]string{"isLocal==true"}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, finalIsLocalTrueCount)

	countIncrease := finalIsLocalTrueCount.Count - initialIsLocalTrueCount.Count
	require.Equal(t, int64(numTestUsers), countIncrease, "Expected isLocal==true count to increase by %d users, but it increased by %d", numTestUsers, countIncrease)

	// Local users must not affect the isLocal==false count
	finalIsLocalFalseCount, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).FilterBy([]string{"isLocal==false"}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, initialIsLocalFalseCount.Count, finalIsLocalFalseCount.Count, "isLocal==false count should not change after creating local users")
}

func TestGetUsers_Pagination(t *testing.T) {
	ctx := newTestContext()

	const numTestUsers = 5
	createdUserIDs := createTestUsers(t, ctx, numTestUsers)
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	firstPageUsers, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Limit(2).Offset(0).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, firstPageUsers)
	require.LessOrEqual(t, len(firstPageUsers), 2)

	secondPageUsers, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Limit(2).Offset(2).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, secondPageUsers)

	if len(firstPageUsers) > 0 && len(secondPageUsers) > 0 {
		require.NotEqual(t, firstPageUsers[0].Id, secondPageUsers[0].Id, "First page and second page should contain different users")
	}

	allUsers, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).Limit(500).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, allUsers)
	require.GreaterOrEqual(t, len(allUsers), numTestUsers)
}

func TestLogoutUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	resp, err := OpenapiClient.UsersAPI.LogoutUser(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogoutUser_NonExistentUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, err := OpenapiClient.UsersAPI.LogoutUser(ctx, "non-existent-user").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestLogoutUser_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	resp, err := OpenapiClient.UsersAPI.LogoutUser(ctx, userID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestResetUserPassword(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, _, originalPassword, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	_, resp, err := OpenapiClient.UsersAPI.ResetUserPassword(ctx, userID).Execute()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	tempPassword := strings.Trim(strings.TrimSpace(string(bodyBytes)), "\"")
	require.NotEmpty(t, tempPassword)
	require.NotEqual(t, originalPassword, tempPassword, "New temporary password should be different from the original password")
}

func TestResetUserPassword_NonExistentUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	_, resp, err := OpenapiClient.UsersAPI.ResetUserPassword(ctx, "non-existent-user").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestResetUserPassword_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	_, resp, err := OpenapiClient.UsersAPI.ResetUserPassword(ctx, userID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetUserById(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, email, tempPassword, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	user, resp, err := OpenapiClient.UsersAPI.GetUserById(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, user)

	require.Equal(t, userID, user.Id)
	require.Equal(t, email, user.Username)
	require.Equal(t, TestUserName, user.CreatedBy)
	require.True(t, user.CreatedAt.IsSet())
	require.True(t, user.UpdatedAt.IsSet())
	require.True(t, user.IsLocal.IsSet())
	require.True(t, *user.IsLocal.Get())

	_, _, err = OpenapiClient.TokensAPI.GrantToken(ctx).
		TokenRequest(getUserTokenRequest(email, tempPassword)).Execute()
	require.NoError(t, err)

	user, resp, err = OpenapiClient.UsersAPI.GetUserById(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, user.LastLogin.IsSet(), "lastLogin should be set after user logs in")
	require.NotNil(t, user.LastLogin.Get(), "lastLogin should not be nil after user logs in")
}

func TestGetUserById_NonExistentUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	_, resp, err := OpenapiClient.UsersAPI.GetUserById(ctx, "non-existent-user").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetUserById_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	_, resp, err := OpenapiClient.UsersAPI.GetUserById(ctx, userID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteUserById(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteUserById_NonExistentUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, "non-existent-user").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteUserById_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func getAuthenticatedUserContext(t *testing.T) (context.Context, string, string) {
	ctx := newTestContext()
	userID, username, tempPassword, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		userCleanup(t, ctx, userID)
	})

	token, _, err := OpenapiClient.TokensAPI.GrantToken(ctx).
		TokenRequest(getUserTokenRequest(username, tempPassword)).Execute()
	require.NoError(t, err)
	require.NotNil(t, token)

	userCtx := TestStepsClient.SetContextValuesWithUser(context.Background(), BaseTenantId, username)
	userCtx = context.WithValue(userCtx, ctxKeyAuth, "Bearer "+*token.AccessToken)
	return userCtx, username, tempPassword
}

func TestChangeUserPassword(t *testing.T) {
	t.Parallel()
	userCtx, _, tempPassword := getAuthenticatedUserContext(t)

	req := openapi.UserChangePasswordRequest{
		CurrentPassword: tempPassword,
		NewPassword:     "NewPassword123!",
	}
	_, resp, err := OpenapiClient.UsersAPI.ChangeUserPassword(userCtx).UserChangePasswordRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestChangeUserPassword_InvalidCurrentPassword(t *testing.T) {
	t.Parallel()
	userCtx, _, _ := getAuthenticatedUserContext(t)

	req := openapi.UserChangePasswordRequest{
		CurrentPassword: "WrongPassword123!",
		NewPassword:     "NewPassword123!",
	}
	_, resp, err := OpenapiClient.UsersAPI.ChangeUserPassword(userCtx).UserChangePasswordRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestChangeUserPassword_WeakNewPassword(t *testing.T) {
	t.Parallel()
	userCtx, _, tempPassword := getAuthenticatedUserContext(t)

	req := openapi.UserChangePasswordRequest{
		CurrentPassword: tempPassword,
		NewPassword:     "weak",
	}
	_, resp, err := OpenapiClient.UsersAPI.ChangeUserPassword(userCtx).UserChangePasswordRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeUserPassword_EmptyNewPassword(t *testing.T) {
	t.Parallel()
	userCtx, _, tempPassword := getAuthenticatedUserContext(t)

	req := openapi.UserChangePasswordRequest{
		CurrentPassword: tempPassword,
		NewPassword:     "",
	}
	_, resp, err := OpenapiClient.UsersAPI.ChangeUserPassword(userCtx).UserChangePasswordRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeUserPassword_SameAsCurrentPassword(t *testing.T) {
	t.Parallel()
	userCtx, _, tempPassword := getAuthenticatedUserContext(t)

	req := openapi.UserChangePasswordRequest{
		CurrentPassword: tempPassword,
		NewPassword:     tempPassword,
	}
	_, resp, err := OpenapiClient.UsersAPI.ChangeUserPassword(userCtx).UserChangePasswordRequest(req).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func createTestUsers(t *testing.T, ctx context.Context, count int) []string {
	var userIDs []string
	for i := 0; i < count; i++ {
		userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
		require.NoError(t, err)
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func cleanupTestUsers(t *testing.T, ctx context.Context, userIDs []string) {
	for _, userID := range userIDs {
		userCleanup(t, ctx, userID)
	}
}

func createTestUsersWithDetails(t *testing.T, ctx context.Context, count int) ([]string, []string, []string) {
	var userIDs, usernames, passwords []string
	for i := 0; i < count; i++ {
		userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
		require.NoError(t, err)
		userIDs = append(userIDs, userID)
		usernames = append(usernames, username)
		passwords = append(passwords, password)
	}
	return userIDs, usernames, passwords
}

func userCleanup(t *testing.T, ctx context.Context, userID string) {
	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	if err != nil && resp != nil && resp.StatusCode != http.StatusNotFound {
		t.Logf("Warning: Failed to cleanup user %s: %v", userID, err)
	}
}

func TestBatchCreateUsers_SuccessfulCreation(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(true)),
		},
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	require.Equal(t, int32(3), *response.TotalRequested)
	require.Equal(t, int32(3), *response.TotalSuccessful)
	require.Nil(t, response.TotalFailed)
	require.Len(t, response.Results, 3)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	for i, result := range response.Results {
		require.True(t, result.Success, "User %d creation should succeed", i)
		require.NotNil(t, result.User, "User %d should be present in result", i)
		require.NotEmpty(t, result.User.Id)
		require.Equal(t, users[i].Email, result.User.Username)
		require.NotEmpty(t, result.User.TempPassword)
		require.False(t, result.Error.IsSet(), "Error should not be set for successful creation")
	}
}

func TestBatchCreateUsers_PartialSuccess(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	existingEmail := GenerateUsername()
	existingUser, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).
		UserCreationRequest(openapi.UserCreationRequest{
			Email:         existingEmail,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		}).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer userCleanup(t, ctx, existingUser.Id)

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         existingEmail,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	require.Equal(t, int32(2), *response.TotalRequested)
	require.Equal(t, int32(1), *response.TotalSuccessful)
	require.Equal(t, int32(1), *response.TotalFailed)
	require.Len(t, response.Results, 2)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	require.True(t, response.Results[0].Success)
	require.NotNil(t, response.Results[0].User)
	require.False(t, response.Results[0].Error.IsSet())

	require.False(t, response.Results[1].Success)
	require.NotNil(t, response.Results[1].User)
	require.Empty(t, response.Results[1].User.Id)
	require.Equal(t, existingEmail, response.Results[1].User.Username)
	require.Empty(t, response.Results[1].User.TempPassword)
	require.True(t, response.Results[1].Error.IsSet())
	require.Equal(t, int32(http.StatusConflict), *response.Results[1].Error.Get().Code)
}

func TestBatchCreateUsers_AllFail(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	existingEmail1 := GenerateUsername()
	existingUser1, resp1, err1 := OpenapiClient.UsersAPI.CreateUser(ctx).
		UserCreationRequest(openapi.UserCreationRequest{
			Email:         existingEmail1,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		}).Execute()

	require.NoError(t, err1)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)
	defer userCleanup(t, ctx, existingUser1.Id)

	existingEmail2 := GenerateUsername()
	existingUser2, resp2, err2 := OpenapiClient.UsersAPI.CreateUser(ctx).
		UserCreationRequest(openapi.UserCreationRequest{
			Email:         existingEmail2,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		}).Execute()

	require.NoError(t, err2)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)
	defer userCleanup(t, ctx, existingUser2.Id)

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         existingEmail1,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         existingEmail2,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Nil(t, response)
}

func TestBatchCreateUsers_InvalidEmail(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         "invalid-email",
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Nil(t, response)
}

func TestBatchCreateUsers_EmptyRequest(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: []openapi.UserCreationRequestBulk{},
	}

	_, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	if openapiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		errMsg := string(openapiErr.Body())
		require.Contains(t, errMsg, "value for Users must have at least 1 items")
	}
}

func TestBatchCreateUsers_ExceedsMaxLimit(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := make([]openapi.UserCreationRequestBulk, 11)
	for i := 0; i < 11; i++ {
		users[i] = openapi.UserCreationRequestBulk{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		}
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	_, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	if openapiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		errMsg := string(openapiErr.Body())
		require.Contains(t, errMsg, "value for Users must have at most 10 items")
	}
}

func TestBatchCreateUsers_MaxLimitSuccess(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := make([]openapi.UserCreationRequestBulk, 10)
	for i := 0; i < 10; i++ {
		users[i] = openapi.UserCreationRequestBulk{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		}
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	require.Equal(t, int32(10), *response.TotalRequested)
	require.Equal(t, int32(10), *response.TotalSuccessful)
	require.Nil(t, response.TotalFailed)
	require.Len(t, response.Results, 10)
}

func TestBatchCreateUsers_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	_, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCreateLocalUser_WithNotifyTrue(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	reset := ptr.To(false)
	notify := ptr.To(true)

	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(reset),
		Notify:        *openapi.NewNullableBool(notify),
	}

	user, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, user)
	defer userCleanup(t, ctx, user.Id)

	require.Equal(t, email, user.Username)
	require.NotEmpty(t, user.TempPassword)
	require.NotEmpty(t, user.Id)
}

func TestCreateLocalUser_WithNotifyFalse(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	reset := ptr.To(false)
	notify := ptr.To(false)

	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(reset),
		Notify:        *openapi.NewNullableBool(notify),
	}

	user, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, user)
	defer userCleanup(t, ctx, user.Id)

	require.Equal(t, email, user.Username)
	require.NotEmpty(t, user.TempPassword)
	require.NotEmpty(t, user.Id)
}

func TestCreateLocalUser_WithFirstLastName(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		FirstName:     ptr.To("Alice"),
		LastName:      ptr.To("Smith"),
	}

	created, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, created)
	defer userCleanup(t, ctx, created.Id)

	user, resp, err := OpenapiClient.UsersAPI.GetUserById(ctx, created.Id).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "Alice", user.GetFirstName())
	require.Equal(t, "Smith", user.GetLastName())
}

func TestPatchLocalUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	patchReq := openapi.UserPatchRequest{
		FirstName: *openapi.NewNullableString(ptr.To("Bob")),
		LastName:  *openapi.NewNullableString(ptr.To("Jones")),
	}
	user, resp, err := OpenapiClient.UsersAPI.PatchUserById(ctx, userID).UserPatchRequest(patchReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, user)

	require.Equal(t, "Bob", user.GetFirstName())
	require.Equal(t, "Jones", user.GetLastName())
}

func TestPatchLocalUser_PartialUpdate(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		FirstName:     ptr.To("Carol"),
		LastName:      ptr.To("White"),
	}
	created, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer userCleanup(t, ctx, created.Id)

	// Patch only firstName, lastName should remain unchanged
	patchReq := openapi.UserPatchRequest{
		FirstName: *openapi.NewNullableString(ptr.To("Carol-Updated")),
	}
	updated, resp, err := OpenapiClient.UsersAPI.PatchUserById(ctx, created.Id).UserPatchRequest(patchReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, updated)

	require.Equal(t, "Carol-Updated", updated.GetFirstName())
	require.Equal(t, "White", updated.GetLastName())
}

func TestPatchLocalUser_NonExistentUser(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	patchReq := openapi.UserPatchRequest{
		FirstName: *openapi.NewNullableString(ptr.To("Ghost")),
	}
	_, resp, err := OpenapiClient.UsersAPI.PatchUserById(ctx, "non-existent-user").UserPatchRequest(patchReq).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPatchLocalUser_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	patchReq := openapi.UserPatchRequest{
		FirstName: *openapi.NewNullableString(ptr.To("Hacker")),
	}
	_, resp, err := OpenapiClient.UsersAPI.PatchUserById(ctx, userID).UserPatchRequest(patchReq).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetUsers_FilterByFirstName(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	email := GenerateUsername()
	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		FirstName:     ptr.To("Unique"),
		LastName:      ptr.To("Surname"),
	}
	user, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer userCleanup(t, ctx, user.Id)

	filterBy := []string{"firstName=@Unique"}
	users, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).FilterBy(filterBy).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	found := false
	for _, u := range users {
		if u.Id == user.Id {
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find user by firstName filter")
}

func TestCreateLocalUser_WithNotifyNil(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	email := GenerateUsername()
	reset := ptr.To(false)

	req := openapi.UserCreationRequest{
		Email:         email,
		ResetPassword: *openapi.NewNullableBool(reset),
	}

	user, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, user)
	defer userCleanup(t, ctx, user.Id)

	require.Equal(t, email, user.Username)
	require.NotEmpty(t, user.TempPassword)
	require.NotEmpty(t, user.Id)
}

func TestBatchCreateUsers_WithNotifyTrue(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users:  users,
		Notify: *openapi.NewNullableBool(ptr.To(true)),
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	require.Equal(t, int32(2), *response.TotalRequested)
	require.Equal(t, int32(2), *response.TotalSuccessful)
	require.Nil(t, response.TotalFailed)
	require.Len(t, response.Results, 2)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	for i, result := range response.Results {
		require.True(t, result.Success, "User %d creation should succeed", i)
		require.NotNil(t, result.User, "User %d should be present in result", i)
		require.NotEmpty(t, result.User.Id)
		require.Equal(t, users[i].Email, result.User.Username)
		require.NotEmpty(t, result.User.TempPassword)
		require.False(t, result.Error.IsSet(), "Error should not be set for successful creation")
	}
}

func TestBatchCreateUsers_WithNotifyFalse(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users:  users,
		Notify: *openapi.NewNullableBool(ptr.To(false)),
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	require.Equal(t, int32(2), *response.TotalRequested)
	require.Equal(t, int32(2), *response.TotalSuccessful)
	require.Nil(t, response.TotalFailed)
	require.Len(t, response.Results, 2)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	for i, result := range response.Results {
		require.True(t, result.Success, "User %d creation should succeed", i)
		require.NotNil(t, result.User, "User %d should be present in result", i)
		require.NotEmpty(t, result.User.Id)
		require.Equal(t, users[i].Email, result.User.Username)
		require.NotEmpty(t, result.User.TempPassword)
		require.False(t, result.Error.IsSet(), "Error should not be set for successful creation")
	}
}

func TestBatchCreateUsers_WithNotifyNil(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users: users,
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	require.Equal(t, int32(2), *response.TotalRequested)
	require.Equal(t, int32(2), *response.TotalSuccessful)
	require.Nil(t, response.TotalFailed)
	require.Len(t, response.Results, 2)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	for i, result := range response.Results {
		require.True(t, result.Success, "User %d creation should succeed", i)
		require.NotNil(t, result.User, "User %d should be present in result", i)
		require.NotEmpty(t, result.User.Id)
		require.Equal(t, users[i].Email, result.User.Username)
		require.NotEmpty(t, result.User.TempPassword)
		require.False(t, result.Error.IsSet(), "Error should not be set for successful creation")
	}
}

func TestBatchCreateUsers_WithNotifyTrue_PartialSuccess(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	existingEmail := GenerateUsername()
	existingUser, resp, err := OpenapiClient.UsersAPI.CreateUser(ctx).
		UserCreationRequest(openapi.UserCreationRequest{
			Email:         existingEmail,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		}).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer userCleanup(t, ctx, existingUser.Id)

	users := []openapi.UserCreationRequestBulk{
		{
			Email:         GenerateUsername(),
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
		{
			Email:         existingEmail,
			ResetPassword: *openapi.NewNullableBool(ptr.To(false)),
		},
	}

	bulkRequest := openapi.BulkUserCreationRequest{
		Users:  users,
		Notify: *openapi.NewNullableBool(ptr.To(true)),
	}

	response, resp, err := OpenapiClient.UsersAPI.BatchCreateUsers(ctx).BulkUserCreationRequest(bulkRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, response)

	require.Equal(t, int32(2), *response.TotalRequested)
	require.Equal(t, int32(1), *response.TotalSuccessful)
	require.Equal(t, int32(1), *response.TotalFailed)
	require.Len(t, response.Results, 2)

	var createdUserIDs []string
	for _, result := range response.Results {
		if result.Success && result.User != nil {
			createdUserIDs = append(createdUserIDs, result.User.Id)
		}
	}
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	require.True(t, response.Results[0].Success)
	require.NotNil(t, response.Results[0].User)
	require.False(t, response.Results[0].Error.IsSet())

	require.False(t, response.Results[1].Success)
	require.NotNil(t, response.Results[1].User)
	require.Empty(t, response.Results[1].User.Id)
	require.Equal(t, existingEmail, response.Results[1].User.Username)
	require.Empty(t, response.Results[1].User.TempPassword)
	require.True(t, response.Results[1].Error.IsSet())
	require.Equal(t, int32(http.StatusConflict), *response.Results[1].Error.Get().Code)
}

func TestGetUsersMinimal_NoParams_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	// Listing every user is no longer allowed; at least one of search, id, or filterBy must be provided.
	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUsersMinimal_WithUsernameFilter(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	email := GenerateUsername()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &email)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).FilterBy([]string{"username==" + email}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, users, 1)
	require.Equal(t, email, users[0].Username)
	require.Equal(t, userID, users[0].Id)
}

func TestGetUsersMinimal_WithSearch(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	uniquePrefix := "mintest" + GenerateRandomString("")
	email := uniquePrefix + "@run.ai"
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &email)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Search(uniquePrefix).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	found := false
	for _, u := range users {
		if u.Id == userID {
			found = true
			break
		}
	}
	require.True(t, found, "created user should appear in search results")
}

func TestGetUsersMinimal_SearchDoesNotMatchCreatedBy(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	uniquePrefix := "nocb" + GenerateRandomString("")
	email := uniquePrefix + "@run.ai"
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &email)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	// Search by the admin's username (= createdBy value for all test-created users).
	// Minimal API must not search in createdBy, so our user must not appear.
	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Search(TestUserName).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	for _, u := range users {
		require.NotEqual(t, userID, u.Id, "minimal search must not match on createdBy field")
	}
}

func TestGetUsersMinimal_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetUsersMinimal_InvalidFilter_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).FilterBy([]string{"firstName=@john"}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUsersMinimal_WithLimit(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	const limit = int32(1)
	createdUserIDs := createTestUsers(t, ctx, 2)
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Search("test-").Limit(limit).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.LessOrEqual(t, len(users), int(limit))
}

func TestGetUsersMinimal_ById(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	email := GenerateUsername()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &email)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Id([]string{userID}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, users, 1)
	require.Equal(t, userID, users[0].Id)
	require.Equal(t, email, users[0].Username)
}

func TestGetUsersMinimal_ByMultipleIds(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	const numTestUsers = 3
	createdUserIDs := createTestUsers(t, ctx, numTestUsers)
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Id(createdUserIDs).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, users, numTestUsers)

	returnedIDs := make(map[string]bool)
	for _, u := range users {
		returnedIDs[u.Id] = true
	}
	for _, id := range createdUserIDs {
		require.True(t, returnedIDs[id], "created user %s should appear in response", id)
	}
}

func TestGetUsersMinimal_ByMultipleUsernames(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	emails := []string{GenerateUsername(), GenerateUsername(), GenerateUsername()}
	userIDs := make([]string, 0, len(emails))
	for _, email := range emails {
		uid, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &email)
		require.NoError(t, err)
		userIDs = append(userIDs, uid)
	}
	defer func() {
		for _, id := range userIDs {
			userCleanup(t, ctx, id)
		}
	}()

	filterBy := make([]string, 0, len(emails))
	for _, e := range emails {
		filterBy = append(filterBy, "username=="+e)
	}
	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).FilterBy(filterBy).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, users, len(emails))

	returned := make(map[string]bool, len(users))
	for _, u := range users {
		returned[u.Username] = true
	}
	for _, e := range emails {
		require.True(t, returned[e], "created user %s should appear in response", e)
	}
}

func TestGetUsersMinimal_MultipleUsernames_ContainsOperatorRejected(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	// =@ (contains) is only meaningful with one filter; multiple entries must use ==.
	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).
		FilterBy([]string{"username=@a", "username=@b"}).
		Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUsersMinimal_TooManyFilterBy_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	filterBy := make([]string, 26)
	for i := range filterBy {
		filterBy[i] = "username==u" + GenerateRandomString("")
	}
	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).FilterBy(filterBy).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUsersMinimal_ShortSearch_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	// `search` enforces a 2-char minimum at the OpenAPI layer; a single character must be rejected
	// so the autocomplete can't be used to fan out broad scans.
	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Search("a").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCountUsers_FilterByLastLogin(t *testing.T) {
	ctx := newTestContext()

	filterBy := []string{"lastLogin>=2020-01-01T00:00:00.000Z"}

	countBefore, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).FilterBy(filterBy).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	userID, username, tempPassword, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	defer userCleanup(t, ctx, userID)

	_, _, err = OpenapiClient.TokensAPI.GrantToken(ctx).
		TokenRequest(getUserTokenRequest(username, tempPassword)).Execute()
	require.NoError(t, err)

	countAfter, resp, err := OpenapiClient.UsersAPI.CountUsers(ctx).FilterBy(filterBy).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Greater(t, countAfter.Count, countBefore.Count, "lastLogin filter count should increase after a user logs in")
}

func TestGetUsers_FilterByIsLocalFalse(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userIDs, _, _ := createTestUsersWithDetails(t, ctx, 2)
	defer cleanupTestUsers(t, ctx, userIDs)

	users, resp, err := OpenapiClient.UsersAPI.GetUsers(ctx).FilterBy([]string{"isLocal==false"}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, users)

	createdIDSet := make(map[string]bool)
	for _, id := range userIDs {
		createdIDSet[id] = true
	}
	for _, user := range users {
		require.False(t, createdIDSet[user.Id], "local user %s should not appear in isLocal==false filter results", user.Id)
	}
}

func TestDeleteUserById_VerifyGone(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)

	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, resp, err = OpenapiClient.UsersAPI.GetUserById(ctx, userID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteUserById_VerifyCannotLogin(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, username, tempPassword, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)

	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, _, err = OpenapiClient.TokensAPI.GrantToken(ctx).
		TokenRequest(getUserTokenRequest(username, tempPassword)).Execute()
	require.Error(t, err, "deleted user should not be able to obtain a token")
}

func TestDeleteUserById_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	userID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)

	resp, err := OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp, err = OpenapiClient.UsersAPI.DeleteUserById(ctx, userID).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetUsersMinimal_TooManyIds_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	ids := make([]string, 26)
	for i := range ids {
		ids[i] = "00000000-0000-0000-0000-00000000000" + string(rune('0'+i%10))
	}
	_, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).Id(ids).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUsersMinimal_SearchAndIdUnion(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	// One user that matches the search prefix.
	uniquePrefix := "minunion" + GenerateRandomString("")
	searchEmail := uniquePrefix + "@run.ai"
	searchUserID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &searchEmail)
	require.NoError(t, err)
	defer userCleanup(t, ctx, searchUserID)

	// One user that does NOT match the search prefix; we'll request it by id.
	idOnlyEmail := GenerateUsername()
	idOnlyUserID, _, _, err := TestStepsClient.CreateUser(ctx, BaseRealmName, &idOnlyEmail)
	require.NoError(t, err)
	defer userCleanup(t, ctx, idOnlyUserID)

	users, resp, err := OpenapiClient.UsersAPI.GetUsersMinimal(ctx).
		Search(uniquePrefix).
		Id([]string{idOnlyUserID, searchUserID}). // searchUserID overlaps both filters; must be deduped.
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	returnedIDs := make(map[string]int)
	for _, u := range users {
		returnedIDs[u.Id]++
	}
	require.Equal(t, 1, returnedIDs[searchUserID], "search-matched user should appear exactly once")
	require.Equal(t, 1, returnedIDs[idOnlyUserID], "id-only user should appear exactly once")
}
