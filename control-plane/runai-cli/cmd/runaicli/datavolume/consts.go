package datavolume

const (
	ShortMsg         = "Data volume asset management"
	ProjectFlagUsage = "Specify the project. Higher-scope datavolumes in the same tenant are included. Use 'runai project set <project>' to set the default."

	ListShortMsg                = "list data volume assets"
	FailedToListDatavolume      = "Failed to list data volume assets"
	ListAllFlagUsage            = "Show resources from all projects"
	DefaultDatavolumeColumnsSet = "Default data volume columns set to: %s\n"
	ListExample                 = `# List all data volume assets
runai datavolume list -A

# List data volume assets in a specific project
runai datavolume list -p <project_name>

# List data volume assets with a specific output format
runai datavolume list --yaml`
)
