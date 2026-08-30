package dynamic_schema_upgrades

// NewRunaiConfigEnableNodepoolController enables the runai nodepool-controller
// (go-operator scales it back up).
func NewRunaiConfigEnableNodepoolController() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newOperandEnableUpgrader(true, "nodepool-controller"),
	}}
}

// NewRunaiConfigDisableNodepoolController disables the runai nodepool-controller
// (go-operator scales it down).
func NewRunaiConfigDisableNodepoolController() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newOperandEnableUpgrader(false, "nodepool-controller"),
	}}
}

// NewRunaiConfigEnableProjectController
// enables the runai project-controller (go-operator scales it back up).
func NewRunaiConfigEnableProjectController() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newOperandEnableUpgrader(true, "project-controller"),
	}}
}

// NewRunaiConfigDisableProjectController
// disables the runai project-controller (go-operator scales it down).
func NewRunaiConfigDisableProjectController() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newOperandEnableUpgrader(false, "project-controller"),
	}}
}

// NewRunaiConfigEnableAgent
// enables the runai agent (go-operator scales it back up).
func NewRunaiConfigEnableAgent() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newOperandEnableUpgrader(true, "agent"),
	}}
}

// NewRunaiConfigDisableAgent
// disables the runai agent (go-operator scales it down).
func NewRunaiConfigDisableAgent() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newOperandEnableUpgrader(false, "agent"),
	}}
}
