package constants

const (
	AnnotationsBackendPolicyCounter = "run.ai/policy-update-counter"
	PolicyLabel                     = "run.ai/policy" // this label is used to identify new policies in the cluster that were created by the agent from the backend

	AnnotationsBackendWorkloadCounter = "run.ai/workload-update-counter"
	WorkloadOwnerUIDAnnotation        = "run.ai/workload-owner-uid"
	WorkloadLabel                     = "run.ai/workload" // this label is used to identify new workloads in the cluster that were created by the agent from the backend

	AdditionalNIMWorkload  = "nimWorkload"
	AdditionalVLLMWorkload = "vllmWorkload"

	BoolTrueAsString      = "true"
	BoolFalseAsString     = "false"
	EmptyStorageClass     = "runai-none"
	AnnotationsPhase      = "run.ai/phase"
	AnnotationsReadyPhase = "Ready"
	TenantScope           = "tenant"
	DepartmentScope       = "department"
	ProjectScope          = "project"
	// CloneAnnotation is the annotation key to indication if there was any issue with cloning k8s object
	CloneAnnotation = "clone.run.ai"
	CloneError      = "ReplicationError"

	LabelName = "run.ai/name"

	// RunaiCASecretName is the name of the secret that holds a custom CA certificate
	RunaiCASecretName        = "runai-ca-cert"
	ManagedNodesConfigName   = "runai-managed-nodes-config"
	ExcludedNodepoolName     = "runai-excluded-nodes"
	ShouldBeExcludedLabelKey = "runai.ai/to-exclude"

	// RunaiLimitRangeName is the name of the LimitRange resource
	// project-controller creates in each project namespace.
	RunaiLimitRangeName = "runai-limit-range"

	RunaiResourceManualOverrideLabel = "run.ai/resource-override"

	RunaiAdsSecretName            = "runai-ads-key"
	RunaiAdsEncryptionKeyName     = "key"
	RunaiAdsEncryptionKeyEnvVar   = "RUNAI_ADS_ENCRYPTION_KEY"
	OAuth2ProxyResourcesEnvVar    = "RUNAI_OAUTH2_PROXY_RESOURCES"
	UsernameLoaderResourcesEnvVar = "RUNAI_USERNAME_LOADER_RESOURCES"
	GitSyncResourcesEnvVar        = "RUNAI_GIT_SYNC_RESOURCES"
	S3FileSystemResourcesEnvVar   = "RUNAI_S3_FILE_SYSTEM_RESOURCES"
	UserCredentialsLabel          = "run.ai/user-credentials"

	// Knative and LWS installation status environment variables
	KnativeInstalledEnvVar = "KNATIVE_INSTALLED"
	LwsInstalledEnvVar     = "LWS_INSTALLED"

	PriorityClassIdLabel     = "run.ai/priority-class-id"
	PriorityClassSourceLabel = "run.ai/priority-class-source"
	PriorityClassRunaiSource = "Run.ai"

	// GitAuthenticatingMethodPassword indicates the git authenticating method is by username/password
	GitAuthenticatingMethodPassword = "password"
	// GitAuthenticatingMethodSSHKey indicates the git authenticating method is by ssh key
	GitAuthenticatingMethodSSHKey = "ssh-key"

	EnforceSchedulerAnnotationName = "runai/enforce-scheduler-name"

	AnnotationRunaiApp = "run.ai/application"

	ProjectControllerRoleBindingCm = "runai-project-controller-rolebindings-plugin"

	GvkDeleteBlockersConfigMapName = "project-delete-blockers"

	InteractiveJobTimeLimitSecondsAnnotation                  = "run.ai/interactive-job-time-limit-seconds"
	InteractiveJobMaxIdleDurationSecondsAnnotation            = "run.ai/interactive-job-max-idle-duration-seconds"
	InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation = "run.ai/interactive-preemptible-job-max-idle-duration-seconds"
	TrainingJobTimeLimitSecondsAnnotation                     = "run.ai/training-job-time-limit-seconds"
	TrainingJobMaxIdleDurationSecondsAnnotation               = "run.ai/training-job-max-idle-duration-seconds"
	// Deprecated node-affinity fields, owned by the nodeType epic. JSON-encoded []string.
	NodeAffinityInteractiveAnnotation = "run.ai/node-affinity-interactive"
	NodeAffinityTrainAnnotation       = "run.ai/node-affinity-train"
	QueueDepartmentNameLabel          = "runai/department-name"

	SyncedDepartmentIdLabel = "runai/department-id"

	// ForceDeleteAnnotation carries the former deletionRequest.Force flag ("true"/"false").
	ForceDeleteAnnotation = "kai/force-delete"

	// UpdatedAtAnnotation carries the control-plane's last-modified timestamp (RFC3339) of
	// the org unit. It replaces the former Spec.UpdatedAt field on the legacy CRDs and lets
	// cluster-sync detect control-plane edits on the KAI Project/Department CRDs.
	UpdatedAtAnnotation = "run.ai/updated-at"

	// UseKaiNodePools selects which NodePool CRD the consumers read. When true they read the
	// KAI NodePool CRD (kai.resources/v1alpha1) instead of the legacy run.ai one.
	// Temporary dev-process toggle for the resource-management migration; flipped by the cutover release.
	UseKaiNodePools = true

	ReplicatedProjectNameLabel = "run.ai/replicated-project-name"
	ReplicatedProjectUIDLabel  = "run.ai/replicated-project-uid"
)
