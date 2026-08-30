package consts

// Legacy consts - deprecated commands
const (
	ListShortMsg       = "[Deprecated] display resource list. By default displays the job list"
	SubmitShortMsg     = "[Deprecated] Submit a new workload"
	SubmitDistShortMsg = "[Deprecated] Submit a new distributed job"

	ListClusterShortMsg = "[Deprecated] list all available clusters"
	ListClusterDepMsg   = "please use: `runai cluster list` instead"
	ListNodeShortMsg    = "[Deprecated] list all nodes"
	ListNodeDepMsg      = "please use: `runai node list` instead"
	ListProjectShortMsg = "[Deprecated] list all available projects"
	ListProjectDepMsg   = "please use: `runai project list` instead"
	ListJobShortMsg     = "[Deprecated] list all jobs"
	ListJobDepMsg       = "please use: `runai workload list` instead"
	DepConfigProjectMsg = "[Deprecated] Configure a default project"
	DepConfigProject    = "please use: `runai project set` instead"
	DepSubmitMsg        = "please use: `runai workspace submit` or `runai training submit` instead"
	DepSubmitDistMsg    = "please use: `runai training <framework> submit` instead"
	ListDepMsg          = "please use list in the relevant context: `runai project list`, `runai workload list`, `runai node list`, `runai cluster list`"

	// legacy port-forward
	PortForwardShortMsgDep = "[Deprecated] port forward"
	DepPortForwardMsg      = "please use: `runai <workload_type> port-forward` instead"

	// legacy logs
	LogsShortMsgDep = "[Deprecated] logs"
	DepLogsMsg      = "please use: `runai <workload_type> logs` instead"

	// legacy exec
	ExecShortMsgDep = "[Deprecated] exec"
	DepExecMsg      = "please use: `runai <workload_type> exec` instead"

	// legacy attach
	AttachShortMsgDep = "[Deprecated] attach"
	DepAttachMsg      = "please use: `runai <workload_type> attach` instead"

	// legacy describe
	DescribeShortMsgDep     = "[Deprecated] Display detailed information about resources"
	DepDescribeMsg          = "please use: `runai workload describe` instead"
	DescribeJobShortMsgDep  = "[Deprecated] Display details of a job"
	DepDescribeJobMsg       = "please use: `runai workload describe` instead"
	DescribeNodeShortMsgDep = "[Deprecated] Display detailed information about nodes in the cluster"
	DepDescribeNodeMsg      = "please use: `runai node list` instead"

	// legacy distributed training frameworks (TF, XGBoost, JAX)
	DepFrameworkMsg = "this framework is deprecated and scheduled for removal in a future release. Existing workloads can still be managed via 'runai workload list/delete/describe --framework <name>'"
)
