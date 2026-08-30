package credentials

import "fmt"

type LoginType string

const (
	AccessKeyIDLogin   LoginType = "AccessKeyIDLogin"
	ApplicationIDLogin LoginType = "ApplicationIDLogin"
)

func (lt LoginType) GetDisplayName() string {
	switch lt {
	case AccessKeyIDLogin, ApplicationIDLogin:
		return "Access Key"
	default:
		return string(lt)
	}
}

type Type interface {
	GetType() LoginType
	GetID() string
	GetSecret() string
}

func GetCredentialTypeFactory(loginType LoginType) (Type, error) {
	if loginType == AccessKeyIDLogin || loginType == ApplicationIDLogin {
		return &AccessKeyAuth{}, nil
	}

	return nil, fmt.Errorf("cannot create credential type of: %s", loginType)
}

func GetDisplayNameFromString(typeStr string) string {
	return LoginType(typeStr).GetDisplayName()
}
