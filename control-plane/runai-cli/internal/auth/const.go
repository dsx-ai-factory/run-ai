package auth

import "strings"

const (
	LoginWithApplicationCreds = "Authenticating with application credentials"
	FailedAuthAppMsg          = "authentication server issue"
	EOFAuthErr                = "connection issue, please check your connection to the authentication server and try again"
	FailedSaveTokenMsg        = "failed to save application token"
	AppLoginSuccessfullyMsg   = "Authentication was successful"
	SetClusterProperties      = "Cluster %s set successfully"

	AuthenticationSecretFileReadError = "failed to read application secret file"
)

const (
	IdTokenKey     = "id_token"
	AccessTokenKey = "access_token"
	CLIClientID    = "cli"
)

const (
	ApplicationStr = "Application"
	SsoStr         = "SSO"
	UserStr        = "User"

	ApplicationKey = "Access key"
	FilePath       = "File Path"
)

var (
	AppOption = []string{ApplicationKey, FilePath}
)

type Method int

const (
	Unknown Method = iota
	Application
	Browser
)

const (
	UnknownString     = "Unknown"
	ApplicationString = "Application"
	BrowserString     = "Browser"
)

var (
	lowerCaseMethodMap = map[string]Method{
		strings.ToLower(ApplicationString): Application,
		strings.ToLower(BrowserString):     Browser,
		strings.ToLower(UnknownString):     Unknown,
	}
)

func (m Method) String() string {
	return [...]string{UnknownString, ApplicationString, BrowserString}[m]
}

func MethodFromString(s string) Method {
	method, ok := lowerCaseMethodMap[strings.ToLower(s)]
	if !ok {
		return Unknown
	}
	return method
}
