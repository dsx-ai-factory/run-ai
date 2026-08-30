package consts

// Project strings
const (
	DepartmentStr           = "department"
	DepartmentShortMsg      = "department management"
	FailedToListDepartments = "failed to list departments"
	ErrorNoDepartmentName   = "department name is missing"
	ListDepartmentsNotFound = "department name not found"
	ErrorNoDepartmentFound  = "department name: %s, not found"
	SetDepartmentProperties = "Department %s set successfully as the default department\n"

	ProjectStr               = "project"
	ProjectShort             = "p"
	ProjectShortMsg          = "project management"
	ProjectAll               = "all"
	FailedToListProject      = "failed to list project"
	SetProjectProperties     = "Project %s set successfully as the default project\n"
	ProjectAttributeSetUsage = "set default project name"
	FailedToProjectList      = "failed to generate project list"
	FailedToDefaultProject   = "failed to set default project"
	ErrorSelectionEmpty      = "selection returned empty"
	ErrorNoProjectFound      = "project name: %s, not found"

	ProjectAttributeExplanation = "Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default."
)

// AI Application strings
const (
	AIApplicationStr = "application"
)

// Nodes strings
const (
	NodeStr           = "node"
	NodeShortMsg      = "node management"
	FailedToListNodes = "Nodes list failed"
)

// Nodepools strings
const (
	NodePoolShortMsg      = "node pool management"
	FailedToListNodePools = "Nodepools list failed"
)
