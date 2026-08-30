package utils

import (
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"k8s.io/utils/ptr"
)

const (
	USER_ID       = "id"
	USER_EMAIL    = "test@run.ai"
	CREATED_BY    = "admin@run.ai"
	USER_PASSWORD = "password"
)

func GetUser(email *string) *iamModels.User {
	return &iamModels.User{
		Username:  email,
		Email:     email,
		ID:        ptr.To(USER_ID),
		CreatedBy: ptr.To("CREATED_BY"),
	}
}

func GetUserChangePasswordRequest() api.UserChangePasswordRequest {
	return api.UserChangePasswordRequest{
		CurrentPassword: USER_PASSWORD,
		NewPassword:     USER_PASSWORD + "1",
	}
}

func GetUsers() []*iamModels.User {
	return []*iamModels.User{
		{
			Username:  ptr.To(USER_EMAIL),
			Email:     ptr.To(USER_EMAIL),
			ID:        ptr.To(USER_ID),
			CreatedBy: ptr.To(CREATED_BY),
		},
		{
			Username:  ptr.To(USER_EMAIL),
			Email:     ptr.To(USER_EMAIL),
			ID:        ptr.To(USER_ID),
			CreatedBy: ptr.To(CREATED_BY),
		},
	}
}

func GetNoUsers() []*iamModels.User {
	return []*iamModels.User{}
}

func GetUserWithGroups() *iamModels.User {
	list := []string{"nottest"}
	return &iamModels.User{
		Username:  ptr.To(USER_EMAIL),
		Email:     ptr.To(USER_EMAIL),
		ID:        ptr.To(USER_ID),
		CreatedBy: ptr.To(CREATED_BY),
		Groups:    &list,
	}
}

func GetLocalAndSSOUsers() []*iamModels.User {
	i := 0
	usersNum := 5
	usersArray := []*iamModels.User{}
	for i < usersNum {
		var isLocalUser bool
		if i%2 == 0 {
			isLocalUser = false
		} else {
			isLocalUser = true
		}
		user := &iamModels.User{
			Username:  ptr.To(USER_EMAIL),
			Email:     ptr.To(USER_EMAIL),
			ID:        ptr.To(USER_ID),
			CreatedBy: ptr.To(CREATED_BY),
			IsLocal:   ptr.To(isLocalUser),
		}
		usersArray = append(usersArray, user)
		i++
	}
	return usersArray
}
