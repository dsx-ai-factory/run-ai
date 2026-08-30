package users

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
)

const (
	GROUPS                 = "GROUPS"
	UPDATE_PASSWORD_ACTION = "UPDATE_PASSWORD"
	UserIDParam            = "uid"
)

func ToUserModel(kcUser gocloak.User) *models.User {
	email := getEmailWithFallback(kcUser)
	user := &models.User{
		Username:  &email,
		ID:        kcUser.ID,
		Email:     &email,
		CreatedBy: gocloak.StringP(getUserAttribute(kcUser.Attributes, utils.CreatedBy)),
		CreatedAt: utils.FormatDate(getUserAttribute(kcUser.Attributes, utils.CreationTime)),
		UpdatedAt: utils.FormatDate(getUserAttribute(kcUser.Attributes, utils.UpdateTime)),
		LastLogin: utils.FormatDate(getUserAttribute(kcUser.Attributes, utils.LastLoginTime)),
		IsLocal:   utils.StringToBool(getUserAttribute(kcUser.Attributes, utils.IsLocalUser)),
		Groups:    getUserAttributeSlice(kcUser.Attributes, GROUPS),
		FirstName: kcUser.FirstName,
		LastName:  kcUser.LastName,
	}
	return user
}

func ToUserModelFromCustomUser(customUser map[string]string) *models.User {
	userName := customUser["email"]
	user := &models.User{
		Username:  gocloak.StringP(userName),
		ID:        gocloak.StringP(customUser["id"]),
		Email:     gocloak.StringP(userName),
		CreatedBy: gocloak.StringP(customUser["createdBy"]),
		CreatedAt: utils.FormatDate(customUser["createdAt"]),
		UpdatedAt: utils.FormatDate(customUser["updatedAt"]),
		LastLogin: utils.FormatDate(customUser["lastLogin"]),
		IsLocal:   utils.StringToBool(customUser["isLocal"]),
		Groups:    utils.SplitCommaSeparated(customUser["groups"]),
	}
	if v := customUser["firstName"]; v != "" {
		user.FirstName = gocloak.StringP(v)
	}
	if v := customUser["lastName"]; v != "" {
		user.LastName = gocloak.StringP(v)
	}
	return user
}

func FromUserModel(user *models.User, internalUser *gocloak.User) *gocloak.User {
	internalUser.Enabled = user.Enabled
	internalUser.Username = user.Username
	internalUser.Email = user.Email
	internalUser.EmailVerified = gocloak.BoolP(true)
	internalUser.RequiredActions = getRequiredUserActions(user.ResetPassword)
	internalUser.Credentials = getCredentialRepresentation(user.TempPassword)
	internalUser.Attributes = generateUserAttributes(*user.CreatedBy)
	internalUser.FirstName = user.FirstName
	internalUser.LastName = user.LastName
	return internalUser
}

// ApplyUserPatch applies non-nil patch fields to an existing Keycloak user.
func ApplyUserPatch(existing *gocloak.User, patch *models.User) *gocloak.User {
	if patch.FirstName != nil {
		existing.FirstName = patch.FirstName
	}
	if patch.LastName != nil {
		existing.LastName = patch.LastName
	}
	return existing
}

func getRequiredUserActions(resetPassword *bool) *[]string {
	var requiredActions *[]string
	if resetPassword == nil || *resetPassword {
		requiredActions = &[]string{UPDATE_PASSWORD_ACTION}
	}
	return requiredActions
}

func getCredentialRepresentation(password *string) *[]gocloak.CredentialRepresentation {
	return &[]gocloak.CredentialRepresentation{
		{
			Type:  gocloak.StringP("password"),
			Value: password,
		},
	}
}

func generateUserAttributes(createdBy string) *map[string][]string {
	currentTime := utils.UnixTimestampToString()
	attributes := &map[string][]string{
		utils.IsLocalUser:  {"true"},
		utils.CreationTime: {currentTime},
		utils.UpdateTime:   {currentTime},
	}
	if createdBy != "" {
		(*attributes)[utils.CreatedBy] = []string{createdBy}
	}
	return attributes
}

func getEmailWithFallback(user gocloak.User) string {
	if user.Email == nil || *user.Email == "" {
		if user.Username == nil {
			return ""
		}
		return *user.Username
	}
	return *user.Email
}

func getUserAttribute(attributesMap *map[string][]string, attributeName string) string {
	attributeValues := getUserAttributeSlice(attributesMap, attributeName)
	if attributeValues != nil && len(*attributeValues) > 0 {
		return (*attributeValues)[0]
	}
	return ""
}

func getUserAttributeSlice(attributesMap *map[string][]string, attributeName string) *[]string {
	if attributesMap == nil {
		return nil
	}
	values, ok := (*attributesMap)[attributeName]
	if !ok || len(values) == 0 {
		return nil
	}
	return &values
}
