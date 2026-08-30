package consts

const (
	NL = "\n"
)

// Flags validation consts
const (
	ErrorOutputType         = "cannot set output type, use `runai config set --output-type=<TYPE> to set a valid type"
	ErrorNameOrUuidIsNeeded = "one of the attributes must be set name|uuid"
	ErrorInvalidColumns     = "invalid column names: %s"
)

// Groups consts
const (
	GroupWorkloadID    = "workload"
	GroupWorkloadTitle = "Workloads Management:"

	GroupApplicationID    = "application"
	GroupApplicationTitle = "AI Applications Management:"

	GroupDeprecatedID    = "deprecated"
	GroupDeprecatedTitle = "Deprecated Commands:"

	GroupResourcesID    = "resources"
	GroupResourcesTitle = "Resources:"

	GroupOrganizationID    = "organization"
	GroupOrganizationTitle = "Organization:"

	GroupUserManagementID    = "userManagement"
	GroupUserManagementTitle = "User Management:"

	GroupConfigID    = "config"
	GroupConfigTitle = "Configuration and Information:"

	GroupTrainingID    = "training"
	GroupTrainingTitle = "Training Frameworks:"

	GroupTrainingBaseID    = "trainingBase"
	GroupTrainingBaseTitle = "Training General Commands:"

	GroupTrainingStdID    = "trainingStandard"
	GroupTrainingStdTitle = "Training Standard Commands:"

	GroupInferenceID    = "inferenceType"
	GroupInferenceTitle = "Inference Architecture:"

	GroupInferenceBaseID    = "inferenceBase"
	GroupInferenceBaseTitle = "Inference General Commands:"

	GroupInferenceStdID    = "inferenceStandard"
	GroupInferenceStdTitle = "Inference Standard Commands:"

	GroupAssetsID    = "assets"
	GroupAssetsTitle = "Assets:"
)

// CheckErrorsPrint
const (
	BadCpUrl          = "Error: Control Plane URL is missing. %v\nSet it by running 'runai config set --cp-url=<URL>'"
	BadAuthUrl        = "Error: Authentication failed. Authentication URL is missing. %v\nSet it by running 'runai config set --auth-url <auth_url>'"
	BadClusterUrl     = "Error: Cluster URL is missing. %v\nSet it by running 'runai cluster set <CLUSTER_NAME>'"
	EmptyProjectValue = "Error: project is missing. %v.\n" +
		"Set a default project by running 'runai project set <project-name>' or use flag -p <project-name>"
	BadWorkloadName         = "Error: invalid workload name. %v."
	TokenError              = "Error: %v"
	TokenErrorIfMissingAuth = "You are not logged in" //nolint:gosec // no token is used
	ReLoginMsg              = "Log in by running 'runai login'"
	BadUrlMsg               = "Error: URL is incorrect or does not exist. DNS lookup failed for '%s': %s\nSet a valid URL by running 'runai config set --cp-url=<CP_URL>'"
	ClusterTimedOutMsg      = "please check cluster availability or change the default timeout duration with the 'runai config set --status-timeout-duration' . Get \"%s\": %s"
	PermissionDenied        = "You have insufficient permissions. Ensure you have the required role/permissions to run this command"
	VersionCheckError       = "Error: version validation issue. %v"
)

// Root (no command) const string
const (
	AppName                            = "runai"
	InteractiveArgumentUsage           = "set interactive mode (enabled|disabled)"
	StatusTimeoutDurationArgumentUsage = "set cluster status call timeout duration value, the default is 3 second (\"3s\")"
	RootShortMsg                       = "Run:ai Command-line Interface"
	RootLongMsg                        = `runai - The Run:ai Researcher Command Line Interface
	
Description:  
  A tool for managing Run:ai workloads and monitoring available resources.
  It provides researchers with comprehensive control over their AI development environment.
`
)

// Config const strings
const (
	ConfigShortMsg = "configuration management"

	SelectionInputWithVisibleDefault = "%s (%s)"
	SelectionCPUrl                   = "Control Plane URL"
	SelectionAuthUrl                 = "Authentication URL"
	SelectionStatusDurationTimeout   = "Cluster Status Check Timeout Duration"

	ConfigInteractiveExecutionError = "Config interactive"
	ConfigFileGenerationFailed      = "Config generation failed"
	ConfigGenerateFailed            = "config generation failed"
	ConfigSetFailed                 = "Configuration set failed, attribute: %s"
	ConfigFileGenerated             = "Config file was generated successfully"

	ErrorOneAttributeRequired = "at least of the attributes is required json|yaml|file"
	ConfigSetSuccess          = "Config set successfully"

	ConfigSetExample = "runai config set --status-timeout-duration 5s\n" +
		"runai config set --status-timeout-duration 300ms"
)

// Login const strings
const (
	LoginShortMsg    = "login to the control plane"
	LoginSSOShortMsg = "login using sso without browser"

	ChooseUsingExistingLoginSubject     = "Use Configured Authentication:"
	ChooseLoginViaByExistingLoginValues = "Use configured credentials: %s"
	ChooseReplaceExistingLoginValues    = "Set different configuration"
	ChooseErrorExitingWithoutSelecting  = "no login type was selected"

	ChooseErrorExitingCredLoadFailed = "failed to load credentials"

	FailedToAuthenticate = "Authenticate failed"
	SuccessfullyLoggedIn = "Authentication was successful"

	AccessKeyMethodSelection = "Access key login method selection:"
	AccessKeySuccessfullySet = "Access key successfully set"

	LoginAccessKeyCredentialSetError = "login access key credential set" //nolint:gosec // no cred is viewed
	LoginAccessKeyMissingAttributes  = "Error: access key login is missing attributes to log in\n" +
		"either use `runai login access-key --client-id=<CLIENT_ID> --secret=<SECRET>` or use `runai login access-key --help\n"
	ErrorLoginAccessKeySecretRequired = "one of the attributes must be set secret|secret-file"
	OpenBrowserError                  = "failed to open browser: %w\n please try to use other login methods, for remote browser login run 'runai login remote-browser'"
	InvalidAddress                    = "invalid address: %s, please set a valid address in the format: <host>:<port>"
	ListenPortUsage                   = "the port to listen on for the authentication callback (for browser mode only)"
	ListenHostUsage                   = "the host to listen on for the authentication callback (for browser mode only)"
	RemoteBrowserOpenUrlMsg           = "Open the following URL in a browser and follow the instructions: %s\nCopy the code from the browser and paste it here: "
	PrintUrlBrowserModeMsg            = "Open the following URL in a browser: %s\n"
	FailedCopyCodeFromUser            = "failed to read auth code"
	EmptyCodeError                    = "empty code"

	ExampleLogin = `
  # Login using browser
  runai login

  # Login using SSO with remote browser
  runai login sso
  runai login remote-browser

  # Login using username and password without browser
  runai login user -u <username> 

  # Login using browser with specific port and host
  runai login --listen-port=43121 --listen-host=localhost
`
	ExampleLoginUser = `
# Login with a username. the password will be prompted via stdin afterward (recommended)
runai login user -u <username>

# Login with a username and plain password (not recommended for security reasons)
runai login user --user=user --password=pass

# Login with a username and password (not recommended for security reasons)
runai login user -u=user -p=pass
`

	LoginFailed            = "login failed"
	LoginUserShortMsg      = "login for local user without browser"
	LoginUserLongMsg       = "Login to the control plane using a local user without browser"
	LoginUserUsernameUsage = "the username to login with"
	LoginUserPasswordUsage = "plaintext password of the given username. not recommended for security reasons"
)

// Cluster strings
const (
	ClusterStr          = "cluster"
	ClusterShortMsg     = "cluster management"
	ClusterListShortMsg = "cluster list command"
	ClusterSetShortMsg  = "set cluster context"

	FailedToListClusters             = "Cluster list failed"
	FailedToSetCluster               = "Cluster setting failed"
	MoreOrLessThanOneCluster         = "more/less than one cluster found; skip auto cluster selection"
	ClusterAlreadySet                = "cluster already set"
	ErrorClusterSetClusterIdWithName = "Set name argument or use --%s flag"
	SingleClusterDefined             = "Cluster %s was automatically defined successfully"
	FailedToAutoComplete             = "Failed auto-completion"
)

// KubeConfig strings
const (
	KubeConfigStr         = "kubeconfig"
	KubeConfigShortMsg    = "kubeconfig management"
	KubeConfigSetShortMsg = "kubeconfig set login token"

	FailedToValidateClusterKubeConfigMatch = "failed cluster domain matching"
	FailedToUpdateKubeConfigBadCFile       = "failed to set kubeconfig file. Missing or misconfigured kubeconfig file. verify that you have a working kubeconfig file on your local machine and that its settings are correct"
	KubeConfigSetSuccessfully              = "Kubeconfig token set successfully"
)

// Environment name and the default values
const (
	Home                    = "HOME"
	EnvRunaiCliConfigPath   = "RUNAI_CLI_CONFIG_PATH"
	EnvRunaiCliConfigFile   = "RUNAI_CLI_CONFIG_FILE"
	EnvRunaiCliSettingsFile = "RUNAI_CLI_SETTINGS_FILE"
	DefaultRunaiDir         = ".runai"
	DefaultConfigFile       = "config.json"
	DefaultInteractive      = "true"
	DefaultSettingsFile     = "settings.json"
)

// Compute strings
const (
	ComputeStr = "compute"
)

// CLI attributes names
const (
	ConfigPath            = "config-path"
	ConfigFile            = "config-file"
	Interactive           = "interactive"
	StatusTimeoutDuration = "status-timeout-duration"
	Debug                 = "debug"
	Verbose               = "verbose"
	Quiet                 = "quiet"
)

const (
	MetricsPath = "/metrics"
)

// Upgrade consts
const (
	UpgradeCommand            = "upgrade"
	UpgradeShortMsg           = "upgrades the CLI to the latest version"
	FailedToCreateTmpDir      = "failed to create tmp directoy to download latest version: %w"
	FailedToRemoveTmpDir      = "failed to remove tmp directory %s, please remove it manually"
	FailedToDownloadInstaller = "failed to download installer: %w"
	FailedToClearCacheFile    = "\nfailed to clear cache file: %w"
	MsiNotFound               = "failed to install new version of CLI, msiexec not found, please install it manually"
	UpgradeMsiFailed          = "failed to install New version of CLI, MSI installation failed: %s"
	UpgradeScriptFailed       = "failed to install the new version, installation script failed: %s\n\n%s"
	FailedDownloadBinary      = "failed to download installer"
	VersionMismatchPrefix     = "versions mismatch"
	NoNewVersion              = "you are up to date with the latest version of the CLI"
	UpgradingCLIVersion       = "upgrading CLI to version"
	FailedToUpdateCLI         = "failed to upgrade CLI"
	FailedToClearUpdateCache  = "failed to clear flow condition cache after upgrade"
	UnsupportedOSArch         = "unsupported OS type with architecture"
	ForceUpgrade              = "force upgrade CLI"
	WindowsUpgradeMessage     = "Upgrading the Windows CLI is currently unsupported. Visit the Run:ai user interface to download the latest version."

	EnableAutoUpdate        = "To enable automatic version updates, run:\nrunai config set --auto-update enabled"
	NewVersionAvailableMsg  = "A new version (%s) is available.\nUpdating now...\n"
	UpdateCompleteMsg       = "Update complete. (%s --> %s)"
	ManualUpdateRequiredMsg = "Warning: You are currently using version %s." +
		" A newer version %s is available." +
		"\nTo update, run:\nrunai upgrade"
	CLIVersionTooHighMsg = "Your CLI version (%s) is higher than the Control Plane version." +
		"\nControl Plane URL: %s\nControl Plane-compatible CLI version:" +
		" %s\n\nTo download the compatible CLI version, run:\nrunai upgrade --force"
	FailedToUpdateMsg    = "Failed to update the CLI. try update manually using:\nrunai upgrade"
	FailedToCheckVersion = "Failed to check for CLI updates."
)

// WhoamiShortMsg whoami strings
const (
	WhoamiShortMsg = "show the current logged in user"
)

// VersionShortMsg version consts
const (
	VersionShortMsg = "show the current version of the CLI"
)

// Report consts
const (
	ReportShortMsg        = "[Experimental] report management"
	MetricsShortMsg       = "[Experimental] metrics management"
	MetricsConfigShortMsg = "metrics configuration"
	MetricsClearShortMsg  = "metrics logs deletion"
	MetricsOutputShortMsg = "metrics logs output"

	FailedToSetReportMetrics   = "failed to set report metrics"
	FailedToClearReportMetrics = "failed to clear metrics"
	FailedToPrintReportMetrics = "failed to output metrics"
)

// Flags groups consts
const (
	GroupNameGeneral              = "General"
	GroupNameContainer            = "Container"
	GroupNameCompute              = "Compute"
	GroupNameStorage              = "Storage"
	GroupNameNetwork              = "Network"
	GroupNameLifecycle            = "Lifecycle"
	GroupNameSecurity             = "Security"
	GroupNameScheduling           = "Scheduling"
	GroupNameDistributed          = "Distributed"
	GroupNameDistributedInference = "Distributed Inference"
	GroupNameInference            = "Inference"
	GroupNameNimConfiguration     = "NIM Configuration"
	GroupNameScaling              = "Scaling"
	GroupNameProbes               = "Probes"
)

// list common flags
const (
	AttributeLimitUsage  = "Maximum number of results to return. Used for pagination."
	AttributeOffsetUsage = "Offset for paginated results. Default is 0 (start from first result)."
)

// Environment asset
const (
	EnvironmentStr = "environment"
)

// Datasource asset
const (
	DatasourceStr = "datasource"
)

// Credential asset
const (
	CredentialStr = "credential"
)

// Datavolume asset
const (
	DatavolumeStr = "datavolume"
)
