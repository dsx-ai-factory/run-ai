package settings

import (
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type LogoutRedirectUri struct {
	Uri string `json:"uri"`
}

func NewLogoutRedirectUri() manager.Settinger {
	return &LogoutRedirectUri{}
}

func (l *LogoutRedirectUri) Category() string {
	return utils.SecurityCategory
}

func (l *LogoutRedirectUri) Description() string {
	return "The URI where users will be redirected after logging out of the application"
}

func (l *LogoutRedirectUri) DisplayName() string {
	return "Logout Redirect URI"
}

func (l *LogoutRedirectUri) Key() string {
	return string(api.SETTINGSKEYENUM_LOGOUT_REDIRECT_URI)
}

func (l *LogoutRedirectUri) SetDefault() {
	l.Uri = ""
}

func (l *LogoutRedirectUri) Stage() manager.Stage {
	return manager.GA
}

func (l *LogoutRedirectUri) Validate(value interface{}) bool {
	return true
}
