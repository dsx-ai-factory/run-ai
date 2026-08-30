package credential

import mapset "github.com/deckarep/golang-set/v2"

const (
	ShortMsg         = "Credential asset management"
	ProjectFlagUsage = "Specify the project. Higher-scope credentials in the same tenant are included. Use 'runai project set <project>' to set the default."
	TypeFlag         = "type"
	TypeFlagUsage    = "Filter by credential type (accessKey, dockerRegistry, password, genericSecret, ngcApiKey)."
)

var allowedCredentialTypes = mapset.NewSet[string]("accessKey", "dockerRegistry", "genericSecret", "ngcApiKey", "password")

const (
	ListShortMsg                = "list credential assets"
	FailedToListCredentials     = "Failed to list credential assets"
	ListAllFlagUsage            = "Show resources from all projects"
	DefaultCredentialColumnsSet = "Default credential columns set to: %s\n"
	ListExample                 = `# List all credential assets
runai credential list -A

# List credential assets in a specific project
runai credential list -p <project_name>

# List credential assets filtered by type
runai credential list --type <type>

# List credential assets with a specific output format
runai credential list --yaml`
)
