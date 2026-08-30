package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"k8s.io/utils/ptr"

	authBusiness "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/authorization"
	business "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/filters"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	models "github.com/run-ai/runai/control-plane/identity-manager/pkg/models/users"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	resetPasswordExpirationTimeSeconds = 7 * 24 * 60 * 60 // 7 days in seconds
)

type privilegeEscalationAction string

const (
	privilegeEscalationActionDelete        privilegeEscalationAction = "delete"
	privilegeEscalationActionResetPassword privilegeEscalationAction = "reset password for"
)

type UsersApiService struct {
	authzClient          auth_client.IAuthorizationClient
	iamServices          iam.IamServiceContainer
	config               config.Iam
	notifyConfig         config.Notify
	tenantDomainTemplate string
	authManager          authBusiness.AuthManagerInterface
}

func NewUsersApiService(iamServices iam.IamServiceContainer, configParam *config.Iam, notifyConfig *config.Notify, tenantDomainTemplate string, authManager authBusiness.AuthManagerInterface) api.UsersAPIServicer {
	client := auth_client.NewAuthorizationClient()
	return &UsersApiService{
		authzClient:          client,
		iamServices:          iamServices,
		config:               *configParam,
		notifyConfig:         *notifyConfig,
		tenantDomainTemplate: tenantDomainTemplate,
		authManager:          authManager,
	}
}

func (s *UsersApiService) CreateUser(ctx context.Context, userCreationRequest api.UserCreationRequest) (api.ImplResponse, error) {
	email := userCreationRequest.Email
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)

	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeUser).WithEntityName(email)

	user, err := s.createSingleUser(ctx, realm, userCreationRequest, *createdBy, nil)
	if err != nil {
		errMsg := fmt.Sprintf(`failed creating user: %s`, email)
		return utils.HandleError(ctx, utils.User, utils.EMAIL, email, err, errMsg)
	}

	audit.AuditFromCtx(ctx).WithEntityId(user.Id)
	return api.Response(http.StatusCreated, user), nil
}

func (s *UsersApiService) PatchUserById(ctx context.Context, userId string, req api.UserPatchRequest) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeUser).WithEntityId(userId)

	existing, err := s.iamServices.UserService.GetUserByID(ctx, realm, userId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed to patch user by id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}

	if existing.IsLocal == nil || !*existing.IsLocal {
		errMsg := fmt.Sprintf(`cannot update SSO user with id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, utils.ErrForbidden, errMsg)
	}

	patch := &iamModels.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	err = s.iamServices.UserService.UpdateUser(ctx, realm, userId, patch)
	if err != nil {
		errMsg := fmt.Sprintf(`failed to update user with id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}

	updated, err := s.iamServices.UserService.GetUserByID(ctx, realm, userId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed to get updated user by id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}
	return api.Response(http.StatusOK, models.BuildUserResponse(*updated)), nil
}

func (s *UsersApiService) DeleteUserById(ctx context.Context, userId string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeUser).WithEntityId(userId)

	if _, implResp, err := s.validateUserPrivilegeEscalation(ctx, realm, userId, privilegeEscalationActionDelete); err != nil {
		return implResp, err
	}

	user, err := s.iamServices.UserService.GetUserByID(ctx, realm, userId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed deleting user. failed to get user by id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}
	if user.Email == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Can't delete access rules and service accounts for user with id: %s. Missing email on user", userId)
	} else {
		audit.AuditFromCtx(ctx).WithEntityName(*user.Email)
	}
	if err := s.deleteUserResources(ctx, realm, userId, user.Email); err != nil {
		errMsg := fmt.Sprintf(`failed deleting user by id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *UsersApiService) GetUserById(ctx context.Context, userId string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	user, err := s.iamServices.UserService.GetUserByID(ctx, realm, userId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting user by id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}
	userResponse := models.BuildUserResponse(*user)
	return api.Response(http.StatusOK, userResponse), nil
}

func (s *UsersApiService) GetInternalUserById(ctx context.Context, userId string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if realm == "" {
		errMsg := "missing realm context for internal user lookup"
		return utils.HandleError(ctx, utils.User, utils.ID, userId, utils.ErrInvalidInput, errMsg)
	}

	user, err := s.iamServices.UserService.GetUserByID(ctx, realm, userId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting user by id: %s`, userId)
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
	}
	userResponse := models.BuildUserResponse(*user)
	return api.Response(http.StatusOK, userResponse), nil
}

func (s *UsersApiService) DeleteInternalUserByEmail(ctx context.Context, realm string, email string) (api.ImplResponse, error) {
	user, err := s.iamServices.UserService.GetUserByEmail(ctx, realm, email)
	if err != nil {
		errMsg := fmt.Sprintf(`failed deleting user. failed to get user by email: %s`, email)
		return utils.HandleError(ctx, utils.User, utils.EMAIL, email, err, errMsg)
	}
	if user == nil || user.ID == nil {
		errMsg := "failed deleting user: user or user.ID is nil"
		return utils.HandleError(ctx, utils.User, utils.EMAIL, email, utils.ErrInternal, errMsg)
	}
	if _, hasSubjectID := ctx.Value(runaictx.CfSubjectID).(string); !hasSubjectID {
		if err := s.iamServices.UserService.DeleteUser(ctx, realm, *user.ID); err != nil {
			errMsg := fmt.Sprintf(`failed deleting user by email: %s`, email)
			return utils.HandleError(ctx, utils.User, utils.EMAIL, email, err, errMsg)
		}
		return api.Response(http.StatusOK, nil), nil
	}

	if err := s.deleteUserResources(ctx, realm, *user.ID, user.Email); err != nil {
		errMsg := fmt.Sprintf(`failed deleting user by email: %s`, email)
		return utils.HandleError(ctx, utils.User, utils.EMAIL, email, err, errMsg)
	}
	return api.Response(http.StatusOK, nil), nil
}

func (s *UsersApiService) deleteUserResources(ctx context.Context, realm, userId string, email *string) error {
	if email != nil {
		if err := s.deleteUserAccessRules(ctx, *email); err != nil {
			return fmt.Errorf("failed deleting access rules for user. err: %w", err)
		}
		if err := s.iamServices.ClientService.DeleteAccessKeysByEmail(ctx, realm, *email); err != nil {
			return fmt.Errorf("failed deleting access keys for user. err: %w", err)
		}
	}
	if err := s.iamServices.UserService.DeleteUser(ctx, realm, userId); err != nil {
		return fmt.Errorf("failed deleting user by id: %s. err: %w", userId, err)
	}
	return nil
}

func (s *UsersApiService) GetUsers(ctx context.Context, filterBy []string, sortBy api.UsersFilterSortFields, sortOrder string, offset int32, limit int32, search string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	filters, err := business.ParseFilters(filterBy, sortBy, sortOrder, offset, limit, search)
	if err != nil {
		errMsg := `failed to get users. failed to parse filters`
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`%s. err: %v`, errMsg, err)
		return api.Response(http.StatusBadRequest, err), utils.Error{Code: http.StatusBadRequest, Message: errMsg}
	}
	users, err := s.iamServices.UserService.GetUsers(ctx, realm, *filters)
	if err != nil {
		errMsg := `failed to get users`
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s. err: %v", errMsg, err.Error())
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: errMsg}
	}
	usersArray := processUsers(users)
	return api.Response(http.StatusOK, usersArray), nil
}

func (s *UsersApiService) GetUsersMinimal(ctx context.Context, filterBy []string, search string, ids []string, limit int32) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	filters, bulkUsernames, err := business.ParseUsersMinimalFilters(filterBy, limit, search, ids)
	if err != nil {
		errMsg := `failed to get users minimal. failed to parse filters`
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`%s. err: %v`, errMsg, err)
		return api.Response(http.StatusBadRequest, err.Error()), utils.Error{Code: http.StatusBadRequest, Message: errMsg}
	}

	var users []*iamModels.User
	if search != "" || (len(filterBy) == 1 && len(bulkUsernames) == 0) {
		users, err = s.iamServices.UserService.GetUsers(ctx, realm, *filters)
		if err != nil {
			errMsg := `failed to get users minimal`
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s. err: %v", errMsg, err.Error())
			return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: errMsg}
		}
	}

	seenIDs := make(map[string]struct{}, len(users)+len(ids))
	for _, u := range users {
		if u != nil && u.ID != nil {
			seenIDs[*u.ID] = struct{}{}
		}
	}
	toFetch := make([]string, 0, len(ids))
	for _, userID := range ids {
		if _, dup := seenIDs[userID]; dup {
			continue
		}
		seenIDs[userID] = struct{}{}
		toFetch = append(toFetch, userID)
	}
	fetched := make([]*iamModels.User, len(toFetch))
	var wg sync.WaitGroup
	for i, userID := range toFetch {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			u, err := s.iamServices.UserService.GetUserByID(ctx, realm, id)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("users-minimal: skipping id %q (not found or error: %v)", id, err)
				return
			}
			fetched[i] = u
		}(i, userID)
	}
	wg.Wait()
	for _, u := range fetched {
		if u != nil {
			users = append(users, u)
		}
	}

	if len(bulkUsernames) > 0 {
		users = append(users, s.fetchUsersByUsernames(ctx, realm, bulkUsernames, users)...)
	}

	usersMinimal := make([]api.UserMinimal, 0, len(users))
	for _, user := range users {
		usersMinimal = append(usersMinimal, models.BuildUserMinimalResponse(*user))
	}
	return api.Response(http.StatusOK, usersMinimal), nil
}

func (s *UsersApiService) fetchUsersByUsernames(ctx context.Context, realm string, usernames []string, existing []*iamModels.User) []*iamModels.User {
	seen := make(map[string]struct{}, len(existing)+len(usernames))
	for _, u := range existing {
		if u != nil && u.Username != nil {
			seen[*u.Username] = struct{}{}
		}
	}
	toFetch := make([]string, 0, len(usernames))
	for _, name := range usernames {
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		toFetch = append(toFetch, name)
	}

	fetched := make([]*iamModels.User, len(toFetch))
	var wg sync.WaitGroup
	for i, name := range toFetch {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			exact := true
			results, err := s.iamServices.UserService.GetUsers(ctx, realm, iamModels.UserFilters{
				Username:      &name,
				UsernameExact: &exact,
			})
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("users-minimal: skipping username %q (error: %v)", name, err)
				return
			}
			for _, u := range results {
				if u != nil && u.Username != nil && *u.Username == name {
					fetched[i] = u
					return
				}
			}
		}(i, name)
	}
	wg.Wait()

	out := make([]*iamModels.User, 0, len(fetched))
	for _, u := range fetched {
		if u != nil {
			out = append(out, u)
		}
	}
	return out
}

func (s *UsersApiService) CountUsers(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	filters, err := business.ParseFilters(filterBy, "", "", 0, 0, search)
	if err != nil {
		errMsg := `failed to get users count. failed to parse filters`
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`%s. err: %v`, errMsg, err)
		return api.Response(http.StatusBadRequest, err), utils.Error{Code: http.StatusBadRequest, Message: errMsg}
	}
	users, err := s.iamServices.UserService.GetUsers(ctx, realm, *filters)
	if err != nil {
		return utils.HandleError(ctx, utils.User, "", "", err, `failed to count users`)
	}
	return api.Response(http.StatusOK, api.CountUsers200Response{
		Count: int64(len(users)),
	}), nil
}

func (s *UsersApiService) LogoutUser(ctx context.Context, userId string) (api.ImplResponse, error) {
	audit.AuditFromCtx(ctx).WithAction(utils.AuditActionLogout).WithEntityType(utils.AuditEntityTypeUser).WithEntityId(userId)
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	err := s.iamServices.UserService.LogoutUser(ctx, realm, userId)
	if err != nil {
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, fmt.Sprintf(`failed logging out user with id: %s`, userId))
	}
	return api.Response(http.StatusOK, nil), nil
}

func (s *UsersApiService) ResetUserPassword(ctx context.Context, userId string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeUser).WithAction(utils.AuditActionResetPassword).WithEntityId(userId)

	userEmail, implResp, err := s.validateUserPrivilegeEscalation(ctx, realm, userId, privilegeEscalationActionResetPassword)
	if err != nil {
		return implResp, err
	}

	tempPassword := utils.GenerateTemporaryUserPassword()
	err = s.iamServices.UserService.ResetUserPassword(ctx, realm, userId, tempPassword)
	if err != nil {
		return utils.HandleError(ctx, utils.User, utils.ID, userId, err, fmt.Sprintf(`failed resetting password for user with id %s`, userId))
	}
	err = s.iamServices.UserService.LogoutUser(ctx, realm, userId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`password was reset by admin. failed closing all sessions for user with id %s. err: %v`, userId, err)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully reset password for user %s (email: %s)", userId, *userEmail)
	return api.Response(http.StatusOK, tempPassword), nil
}

func (s *UsersApiService) validateUserPrivilegeEscalation(ctx context.Context, realm, userId string, action privilegeEscalationAction) (*string, api.ImplResponse, error) {
	user, err := s.iamServices.UserService.GetUserByID(ctx, realm, userId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting user for privilege escalation check. user id: %s`, userId)
		implResp, utilErr := utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
		return nil, implResp, utilErr
	}

	if user.Email == nil {
		errMsg := "User email is required for privilege escalation check"
		implResp, utilErr := utils.HandleError(ctx, utils.User, utils.ID, userId, utils.ErrInvalidInput, errMsg)
		return nil, implResp, utilErr
	}

	userEmail := *user.Email
	err = s.authManager.ValidatePrivilegeEscalation(ctx, s.authzClient, userEmail)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Privilege escalation check failed for %s %s (subjectId: %s): %v", utils.User, userId, userEmail, err)

		if utils.IsForbidden(err) {
			errMsg := fmt.Sprintf("You're not authorized to %s the selected user", action)
			implResp, utilErr := utils.HandleError(ctx, utils.User, utils.ID, userId, utils.ErrForbidden, errMsg)
			return nil, implResp, utilErr
		}

		errMsg := "Failed to validate permissions for this user"
		implResp, utilErr := utils.HandleError(ctx, utils.User, utils.ID, userId, err, errMsg)
		return nil, implResp, utilErr
	}

	return &userEmail, api.ImplResponse{}, nil
}

func (s *UsersApiService) BatchCreateUsers(ctx context.Context, bulkUserCreationRequest api.BulkUserCreationRequest) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)

	users := bulkUserCreationRequest.Users

	shouldNotify := s.shouldNotify(bulkUserCreationRequest.Notify)

	var wg sync.WaitGroup
	results := make([]api.BulkUserCreationResponseResultsInner, len(users))
	var totalSuccessful int64 = 0
	var totalFailed int64 = 0

	for i, userReq := range users {
		wg.Add(1)
		go func(i int, userReq api.UserCreationRequestBulk) {
			defer wg.Done()

			email := userReq.Email
			userCreationReq := api.UserCreationRequest{
				Email:         email,
				ResetPassword: userReq.ResetPassword,
				Notify:        ptr.To(shouldNotify),
			}
			user, err := s.createSingleUser(ctx, realm, userCreationReq, *createdBy, nil)

			result := api.BulkUserCreationResponseResultsInner{}
			if err != nil {
				result.Success = false
				result.User = user
				result.Error = s.getUserCreationError(ctx, email, err)
				atomic.AddInt64(&totalFailed, 1)
			} else {
				result.Success = true
				result.User = user
				atomic.AddInt64(&totalSuccessful, 1)
			}

			results[i] = result
		}(i, userReq)
	}

	wg.Wait()
	totalRequested := int32(len(users))

	response := api.BulkUserCreationResponse{
		Results:         results,
		TotalRequested:  totalRequested,
		TotalSuccessful: int32(totalSuccessful),
		TotalFailed:     int32(totalFailed),
	}
	if totalRequested == int32(totalFailed) {
		return api.Response(http.StatusInternalServerError, response), nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *UsersApiService) CreateIntenralUser(ctx context.Context, internalUserCreationRequest api.InternalUserCreationRequest) (api.ImplResponse, error) {
	email := internalUserCreationRequest.Email
	customPassword := utils.GetUserPassword(internalUserCreationRequest.Password)
	userCreationReq := api.UserCreationRequest{
		Email:         email,
		ResetPassword: ptr.To(internalUserCreationRequest.ResetPassword),
		Notify:        ptr.To(s.shouldNotify(internalUserCreationRequest.Notify)),
	}
	user, err := s.createSingleUser(ctx, internalUserCreationRequest.Realm, userCreationReq, "", &customPassword)
	if err != nil {
		response, _ := utils.HandleError(ctx, utils.User, utils.EMAIL, email, err, fmt.Sprintf(`failed creating internal user %s`, email))
		return detailedResponseFromError(response, err.Error()), nil
	}

	return api.Response(http.StatusCreated, user), nil
}

func (s *UsersApiService) ChangeUserPassword(ctx context.Context, userChangePasswordRequest api.UserChangePasswordRequest) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	userName := runaictx.GetString(ctx, runaictx.CfSubjectID)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeUser).WithAction(utils.AuditActionChangePassword).WithEntityName(*userName)
	user, err := s.iamServices.UserService.GetUserByEmail(ctx, realm, *userName)
	if err != nil {
		errMsg := fmt.Sprintf(`failed to change user password. failed to get user: %s. err: %v`, *userName, err)
		return utils.HandleError(ctx, utils.User, utils.ID, *userName, err, errMsg)
	}
	if user == nil || user.ID == nil {
		errMsg := fmt.Sprintf(`failed to change user password: user not found: %s`, *userName)
		return utils.HandleError(ctx, utils.User, utils.ID, *userName, utils.ErrInternal, errMsg)
	}
	audit.AuditFromCtx(ctx).WithEntityId(*user.ID)
	err = s.iamServices.UserService.ChangeUserPassword(ctx, realm, user, userChangePasswordRequest.CurrentPassword, userChangePasswordRequest.NewPassword)
	if err != nil {
		errMsg := fmt.Sprintf(`failed to change user password for user: %s`, *userName)
		response, _ := utils.HandleError(ctx, utils.User, utils.ID, *userName, err, errMsg)
		return detailedResponseFromError(response, err.Error()), nil
	}
	err = s.iamServices.UserService.LogoutUser(ctx, realm, *user.ID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`password was reset by the user. failed closing all sessions for user with id %s. err: %v`, *user.ID, err)
	}
	return api.Response(http.StatusOK, nil), nil
}

func (s *UsersApiService) GetUserAccessGroups(ctx context.Context, username string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)

	user, err := s.iamServices.UserService.GetUserByEmail(ctx, realm, username)
	if err != nil || user == nil {
		errMsg := fmt.Sprintf(`failed to get user's access groups. failed to get user with email: %s`, username)
		return utils.HandleError(ctx, utils.User, utils.EMAIL, username, err, errMsg)
	}
	if user.Groups == nil {
		return api.ImplResponse{
			Code: http.StatusOK,
			Body: []string{},
		}, nil
	}

	userGroups, err := s.authzClient.GetGroupsForTenant(ctx, *tenantId)
	if err != nil {
		errMsg := fmt.Sprintf(`failed to get user's access groups for user with email: %s`, username)
		return utils.HandleError(ctx, utils.GROUPS, utils.ID, *tenantId, err, errMsg)
	}
	intersection := utils.Intersection(userGroups, *user.Groups)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("found access groups: %v for user email: %s ", intersection, username)
	return api.ImplResponse{
		Code: http.StatusOK,
		Body: intersection,
	}, nil
}

func (s *UsersApiService) getUserCreationError(ctx context.Context, userEmail string, err error) *api.BulkUserCreationResponseResultsInnerError {
	errorCode := int32(http.StatusInternalServerError)
	errorMessage := err.Error()

	if utils.IsConflict(err) {
		errorCode = int32(http.StatusConflict)
		errorMessage = "user already exists"
	} else if utils.IsBadRequest(err) {
		errorCode = int32(http.StatusBadRequest)
		errorMessage = "bad request"
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(`batch user creation request fail for user: %s. err: %v`, userEmail, err)
	}
	return &api.BulkUserCreationResponseResultsInnerError{
		Code:    errorCode,
		Message: errorMessage,
	}
}

func (s *UsersApiService) deleteUserAccessRules(ctx context.Context, email string) error {
	request := s.authzClient.DeleteSubjectAccessRules(ctx)
	request = request.SubjectType(string(authorization_api.SUBJECTTYPE_USER)).SubjectId(email)
	_, err := s.authzClient.DeleteSubjectAccessRulesExecute(request)
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersApiService) createSingleUser(ctx context.Context, realm string, userCreationRequest api.UserCreationRequest, createdBy string, customPassword *string) (api.UserPostResponse, error) {
	password := utils.GenerateTemporaryUserPassword()
	if customPassword != nil {
		password = *customPassword
	}
	email := userCreationRequest.Email
	newUser := s.generateUser(email, true, userCreationRequest.ResetPassword, createdBy, password)
	if userCreationRequest.FirstName != "" {
		newUser.FirstName = ptr.To(userCreationRequest.FirstName)
	}
	if userCreationRequest.LastName != "" {
		newUser.LastName = ptr.To(userCreationRequest.LastName)
	}

	userId, err := s.iamServices.UserService.CreateUser(ctx, realm, newUser)
	if err != nil {
		return api.UserPostResponse{
			Id:           "",
			Username:     email,
			TempPassword: "",
		}, err
	}
	if s.shouldNotify(userCreationRequest.Notify) {
		redirectURI := s.getRedirectUriForEmailNotification(realm)
		err = s.iamServices.UserService.SendPasswordResetEmail(ctx, realm, userId, resetPasswordExpirationTimeSeconds, redirectURI, utils.RUNAI_ADMIN_UI_CLIENT_ID)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`failed sending reset password email to user: %s. err: %v`, email, err)
		}
	}
	return models.BuildUserCreationResponse(userId, email, password), nil
}

func (s *UsersApiService) generateUser(email string, enabled bool, resetPassword *bool, createdBy string, password string) *iamModels.User {
	return &iamModels.User{
		Username:      ptr.To(email),
		Email:         ptr.To(email),
		Enabled:       ptr.To(enabled),
		ResetPassword: resetPassword,
		CreatedBy:     ptr.To(createdBy),
		TempPassword:  ptr.To(password),
	}
}

func (s *UsersApiService) getRedirectUriForEmailNotification(realm string) string {
	baseURL, err := utils.TemplateBaseUrl(s.tenantDomainTemplate, realm)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s/login/callback?%s=%s", baseURL, utils.RedirectFlowQueryKey, utils.RedirectFlowInvite)
}

func (s *UsersApiService) shouldNotify(isNotify *bool) bool {
	return s.notifyConfig.Enabled && (isNotify != nil && *isNotify)
}

func processUsers(users []*iamModels.User) []api.User {
	var wg sync.WaitGroup
	usersArray := make([]api.User, len(users))
	mu := &sync.Mutex{}
	for i := range users {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			userResponse := models.BuildUserResponse(*users[i])
			mu.Lock()
			usersArray[i] = userResponse
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	return usersArray
}

func detailedResponseFromError(response api.ImplResponse, err string) api.ImplResponse {
	if response.Code == http.StatusBadRequest {
		startIdx := strings.Index(err, "Invalid password:")
		if startIdx != -1 {
			errMessage := err[startIdx:]
			return api.Response(http.StatusBadRequest, errMessage)
		}
		return api.Response(http.StatusBadRequest, nil)
	}
	return response
}
