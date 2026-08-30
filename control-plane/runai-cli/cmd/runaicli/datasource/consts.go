package datasource

const (
	// Datasource command messages
	ShortMsg         = "Data Source asset management"
	ProjectFlagUsage = "Specify the project. Higher-scope data-sources in the same tenant are included. Use 'runai project set <project>' to set the default."
	TypeFlag         = "type"
	TypeFlagUsage    = "Filter by data source type (nfs, pvc, s3, git, hostPath, configMap, secretVolume, dataVolume)"

	// Describe command messages
	DescribeShortMsg           = "describe a data source asset"
	FailedToDescribeDatasource = "failed to describe data source"
	DescribeExample            = `# Describe a data source asset with a default project
runai datasource describe training-data --type pvc

# Describe a data source asset in a specific project
runai datasource describe training-data --type pvc -p <project_name>

# Describe a data source asset with specific output format
runai datasource describe training-data --type nfs -o json`

	// List command
	ListShortMsg                = "list data source assets"
	FailedToListDatasource      = "Failed to list data source assets"
	ListAllFlagUsage            = "Show resources from all projects"
	DefaultDatasourceColumnsSet = "Default data source columns set to: %s\n"
	ListExample                 = `# List all data source assets
runai datasource list -A

# List data source assets in a specific project
runai datasource list -p <project_name>

# List data source assets with a specific output format
runai datasource list --yaml`
)
