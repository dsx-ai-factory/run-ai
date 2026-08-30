package consts

// API Error Messages
const (
	ErrForbiddenDocument = "Forbidden manual document name: %s"
	ErrDocumentNotExist  = "Manual document does not exist: %s"
	ErrBinaryNotExist    = "Binary does not exist for OS %s, architecture %s"
	ErrVersionNotExist   = "version file does not exist: %s"

	ErrMissingToken       = "error getting token from context"
	ErrorGettingCpURL     = "error getting control-plane URL from context"
	ErrorUserAgent        = "missing user-agent header"
	ErrInternalCliVersion = "invalid internal CLI version"
	ErrClientCliVersion   = "invalid User-Agent Header; runai-cli version is invalid"
	ErrCliVersionMismatch = "versions mismatch, client SDK version: %s is bigger than server SDK version: %s; " +
		"please update your CLI version"
	ErrParsingInstallerTemplate = "failed to parse installer template file "
	ErrExecInstallerTemplate    = "failed to execute installer template "
)

// File Paths
const (
	DocumentPath = "/app/docs/"
)

const (
	VersionPrefix   = "v%s"
	LocalhostPrefix = "localhost"
)

// API Response Messages
const (
	WindowsInstallationCommands = "MSI installation will be supported soon"
	UnixInstallationCommands    = "bash -c \"$(curl -fsSL %s/api/v1/cli-exposer/installer/%s -H 'Authorization: Bearer %s')\""
)

type ContextValue string

const (
	ContextHostUrl  ContextValue = "host-url"
	UserAgentHeader ContextValue = "User-Agent"
)
