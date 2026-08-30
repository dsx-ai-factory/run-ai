package config

import "fmt"

const (
	// defaultKAIPrefix is the single source of truth for the kai.resources
	// label/domain prefix used to construct project-controller's own
	// default-value consts below — distinct from the kai.scheduler/* keys the
	// scheduler owns (node-pool, queue).
	defaultKAIPrefix = "kai.resources"

	defaultProjectIdLabelKey              = defaultKAIPrefix + "/project-id"
	defaultNamespaceVersionLabelKey       = defaultKAIPrefix + "/namespace-version"
	defaultEnforceSchedulerAnnotationKey  = defaultKAIPrefix + "/enforce-scheduler-name"
	defaultFinalizerDomain                = defaultKAIPrefix
	defaultResourceManualOverrideLabelKey = defaultKAIPrefix + "/resource-manual-override"
	defaultLimitRangeName                 = "kai-limit-range"
	defaultQueueDepartmentNameLabelKey    = defaultKAIPrefix + "/department-name"

	defaultEnableProjectValidationWebhook    = true
	defaultEnableDepartmentValidationWebhook = true

	// namespacePrefixSuffix is appended to the configured ProjectNamePrefix
	// to compose the namespace prefix (e.g., "kai" + "-" = "kai-"). See
	// NamespacePrefix() for the composition.
	namespacePrefixSuffix = "-"

	// defaultProjectNamePrefix is the OSS default for --project-name-prefix.
	defaultProjectNamePrefix = "kai"

	// finalizerSuffix is appended to the configured FinalizerDomain to
	// compose the full finalizer string
	finalizerSuffix = ".finalizers.project"

	// departmentFinalizerSuffix is appended to the configured FinalizerDomain
	// to compose the department finalizer string
	departmentFinalizerSuffix = ".finalizers.department"
)

type ProjectReconcilerConfig struct {
	CreateNamespaces        bool
	CreateRoleBindings      bool
	RoleBindingsCm          string
	RoleBindingsCmNamespace string
	ClusterWideSecrets      bool
	ClusterWidePvcs         bool
	ClusterWideConfigMaps   bool
	LimitRange              bool
	IsOpenshift             bool

	// NodePoolLabelKey is the label key used to mark Kubernetes resources
	// (PodGroups, Queues, NodePool selectors) with the node pool they belong
	// to. Flag-configurable (--nodepool-label-key), default: `kai.scheduler/node-pool`.
	NodePoolLabelKey string

	// DefaultNodepoolName is the name of the implicit default node pool. Queues
	// for this node pool carry no node-pool label. Flag-configurable
	// (--default-nodepool-name).
	DefaultNodepoolName string

	// QueueLabelKey is the label key written on PodGroups to identify the queue
	// a workload belongs to. Flag-configurable (--queue-label-key).
	QueueLabelKey string

	// NamespaceProjectLabelKey is the label key written on a project's namespace
	// (and read back to resolve a namespace's project) to identify the project
	// it belongs to. Flag-configurable (--namespace-project-label-key). Paired
	// with pod-group-assigner's flag of the same name — both must agree.
	NamespaceProjectLabelKey string

	// ProjectLabelKey is the label key used to identify the project a workload
	// belongs to (workload side). Flag-configurable (--project-label-key).
	ProjectLabelKey string

	// ProjectIdLabelKey is the label key used to identify the numeric project
	// ID on resources. Flag-configurable (--project-id-label-key). Distinct
	// from ProjectLabelKey: the latter holds the project NAME, this one holds
	// the project ID.
	ProjectIdLabelKey string

	// QueueDepartmentNameLabelKey is the label key written on queues to record
	// the department they belong to (a project queue's parent department, or a
	// department queue's own department name).
	// Flag-configurable (--queue-department-name-label-key).
	QueueDepartmentNameLabelKey string

	// NamespaceVersionLabelKey is the label key project-controller writes on
	// newly-created namespaces to track which schema version they were
	// initialized with. Flag-configurable (--namespace-version-label-key).
	NamespaceVersionLabelKey string

	// EnforceSchedulerAnnotationKey is the annotation key the namespace handler writes
	// on a project's namespace, carrying Project.Spec.EnforceKaiScheduler as "true"/"false".
	// It is the contract with whatever component enforces the scheduler on pods
	EnforceSchedulerAnnotationKey string

	// FinalizerDomain is the DNS-style domain prefix used to compose the
	// project finalizer string. See FinalizerName() for the composition.
	// Flag-configurable (--finalizer-domain).
	FinalizerDomain string

	// InstallNamespace is the namespace project-controller is installed in, and
	// from which it reads its replicated-resource sources (the limit-range source
	// ConfigMap, the cluster-wide-secret source) and watches them. Flag-configurable
	// (--install-namespace); the chart/operator inject their own namespace.
	InstallNamespace string

	// ProjectNamePrefix is the prefix used to name and identify the namespaces
	// project-controller manages (<prefix>-<project>). See NamespacePrefix().
	// Flag-configurable (--project-name-prefix).
	ProjectNamePrefix string

	ResourceManualOverrideLabelKey string

	// EnableProjectValidationWebhook gates registration of the validating
	// admission webhook handler for Project resources. Default off so a
	// deployment that ships no serving certificate is unaffected; the helm
	// chart that wires the webhook turns it on. Flag-configurable
	// (--enable-project-validation-webhook).
	EnableProjectValidationWebhook bool

	// EnableDepartmentValidationWebhook gates registration of the validating
	// admission webhook handler for Department resources. Flag-configurable
	// (--enable-department-validation-webhook).
	EnableDepartmentValidationWebhook bool

	// WebhookPort is the port the validating webhook server listens on.
	// Flag-configurable (--webhook-port).
	WebhookPort int

	// WebhookCertDir is the directory that contains the webhook server's TLS
	// key and certificate. Flag-configurable (--cert-dir).
	WebhookCertDir string

	EnableProfiler       bool
	ProfilerApiPort      string
	K8sClientConfigQPS   int
	K8sClientConfigBurst int

	LimitRangeName string

	GvkDeleteBlockersNamespace string
}

// current holds the parsed configuration. SetOptions assigns it during
// startup; SetForTest replaces it temporarily in tests. Get() returns it for
// read-only access from deep-stack code.
var current *ProjectReconcilerConfig

// Get returns the parsed configuration. Must be called after SetOptions has
// run (i.e., after flag.Parse() in main, or after SetForTest in tests).
func Get() *ProjectReconcilerConfig { return current }

// SetForTest replaces the global config and returns a restore closure.
// Tests only — production code goes through SetOptions.
func SetForTest(c *ProjectReconcilerConfig) func() {
	prev := current
	current = c
	return func() { current = prev }
}

// FinalizerName composes the project-finalizer string from the configured
// FinalizerDomain. default: "kai.scheduler.finalizers.project"
func FinalizerName() string {
	return fmt.Sprintf("%s%s", Get().FinalizerDomain, finalizerSuffix)
}

// DepartmentFinalizerName composes the department-finalizer string from the
// configured FinalizerDomain. default: "kai.scheduler.finalizers.department".
func DepartmentFinalizerName() string {
	return fmt.Sprintf("%s%s", Get().FinalizerDomain, departmentFinalizerSuffix)
}

// NamespacePrefix returns the prefix that identifies namespaces created by
// project-controller (e.g., `kai-team-foo`). Derived from ProjectNamePrefix.
func NamespacePrefix() string {
	return fmt.Sprintf("%s%s", Get().ProjectNamePrefix, namespacePrefixSuffix)
}

func NewProjectReconcilerConfig() *ProjectReconcilerConfig {
	return &ProjectReconcilerConfig{}
}
