package settings

import (
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
	"k8s.io/utils/ptr"
)

type AutoRedirectSSO struct {
	Enabled  bool    `json:"enabled"`
	IdpAlias *string `json:"idpAlias"`
}

func NewAutoRedirectSSO() manager.Settinger {
	return &AutoRedirectSSO{}
}

func (a *AutoRedirectSSO) Category() string {
	return utils.SecurityCategory
}

func (a *AutoRedirectSSO) Description() string {
	return "Automatically redirect users to the SSO login page after logging in to the application"
}

func (a *AutoRedirectSSO) DisplayName() string {
	return "Auto Redirect SSO"
}

func (a *AutoRedirectSSO) Key() string {
	return string(api.SETTINGSKEYENUM_AUTO_REDIRECT_SSO)
}

func (a *AutoRedirectSSO) SetDefault() {
	a.Enabled = false
	a.IdpAlias = ptr.To("")
}

func (a *AutoRedirectSSO) Stage() manager.Stage {
	return manager.GA
}

func (a *AutoRedirectSSO) Validate(value interface{}) bool {
	var valueToUpdate AutoRedirectSSO

	switch v := value.(type) {
	case AutoRedirectSSO:
		valueToUpdate = v
	case *AutoRedirectSSO:
		valueToUpdate = *v
	default:
		return false
	}

	return a.isEnabledWithAlias(valueToUpdate) || a.isDisabled(valueToUpdate)
}

func (a *AutoRedirectSSO) isEnabledWithAlias(valueToUpdate AutoRedirectSSO) bool {
	enabled := valueToUpdate.Enabled
	idpAlias := valueToUpdate.IdpAlias
	return enabled && idpAlias != nil && *idpAlias != ""
}

func (a *AutoRedirectSSO) isDisabled(valueToUpdate AutoRedirectSSO) bool {
	return !valueToUpdate.Enabled
}
