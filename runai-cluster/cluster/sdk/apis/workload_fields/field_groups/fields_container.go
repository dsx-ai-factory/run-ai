package field_groups

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type ContainerFields struct {

	// If set, overrides the image's entry point with the supplied command.
	Command *flds.StringField `json:"command,omitempty"`

	// When set,contains the arguments sent along with the command. These override the
	// entry point of the image in the created workload.
	Arguments *flds.StringField `json:"arguments,omitempty"`

	// Specifies the image to use when creating the container running the workload.
	Image *flds.StringField `json:"image,omitempty"`

	// Specifies the pull policy of the image when starting a container running the created workload. Options are:
	// always, ifNotPresent, or never. For more information see: https://kubernetes.io/docs/concepts/containers/images
	ImagePullPolicy *flds.StringField `json:"imagePullPolicy,omitempty"`

	// Specifies environment variables to be set in the container running the created workload.
	Environment *flds.DictionaryField `json:"environment,omitempty"`

	// Specifies labels to be set in the container running the created workload.
	Labels *flds.DictionaryField `json:"labels,omitempty"`

	// Specifies annotations to be set in the container that is running the created workload.
	Annotations *flds.DictionaryField `json:"annotations,omitempty"`

	// Instructs the system to keep stdin open for the container(s) running the
	// created workload, even if nothing is attached.
	Stdin *flds.BoolField `json:"stdin,omitempty"`

	// Instructs the system to allocate a pseudo-TTY for the created workload.
	Tty *flds.BoolField `json:"tty,omitempty"`

	// Limits the container running the created workload to run in the context of a specific non-root user.
	// The user id is provided by the runAsUid field. This would manifest itself in access to operating system resources,
	// in the ownership of new folders created under shared directories, etc. Alternatively, if your cluster is connected
	// to Run:ai via SAML, you can map the container to use the Linux UID/GID which is stored in the organization's
	// directory. For more information see the User Identity guide at https://docs.run.ai/admin/runai-setup/config/non-root-containers/
	RunAsUser *flds.BoolField `json:"runAsUser,omitempty"`

	// Specifies the Unix user id with which the container running the created workload should run.
	// Will be used only if runAsUser is set to true.
	RunAsUid *flds.IntField `json:"runAsUid,omitempty"`

	// Specifies the Unix group id with which the container should run.
	// Will be used only if runAsUser is set to true.
	RunAsGid *flds.IntField `json:"runAsGid,omitempty"`

	// ';' separated list of supplemental group IDs. Will be added to the security context of the
	// container running the created workload.
	SupplementalGroups *flds.StringField `json:"supplementalGroups,omitempty"`

	// Specifies a directory that will be used as the current directory when the container
	// running the created workload starts.
	WorkingDir *flds.StringField `json:"workingDir,omitempty"`

	// Instructs the system to create a temporary home directory for the user within the container.
	// Data stored in this directory will not be saved when the container exits.
	// When the runAsUser flag is set to true, this flag will default to true as well.
	CreateHomeDir *flds.BoolField `json:"createHomeDir,omitempty"`

	// Toleration rules which apply to the pods running the workload. Toleration rules guide (but do not require) the
	// system to which node each pod can be scheduled to or evicted from, based on matching between those rules
	// and the set of taints defined for each Kubernetes node.
	Tolerations *flds.TolerationsField `json:"tolerations,omitempty"`

	// The capabilities field allows adding a set of unix capabilities to the container running the workload.
	// Capabilities are Linux distinct privileges traditionally associated with superuser which can be independently enabled and
	// disabled. For more information see https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container
	Capabilities *flds.CapabilitiesField `json:"capabilities,omitempty"`

	// Indicates that the container must run as a non-root user.
	// If true, the Kubelet will validate the image at runtime to ensure that it
	// does not run as UID 0 (root) and fail to start the container if it does.
	// If unset or false, no such validation will be performed.
	RunAsNonRoot *flds.BoolField `json:"runAsNonRoot,omitempty"`

	// If true, mounts the container's root filesystem as read-only.
	ReadOnlyRootFilesystem *flds.BoolField `json:"readOnlyRootFilesystem,omitempty"`

	// Indicates which kind of seccomp profile will be applied to the container.
	// Valid options are:
	// RuntimeDefault - the container runtime default profile should be used.
	// Unconfined - no profile should be applied.
	// Localhost is not yet supported by Run:ai.
	SeccompProfileType *flds.StringField `json:"seccompProfileType,omitempty"`

	// Indicates whether to drop all capabilities or not. Default is false.
	DropAllCapabilities *flds.BoolField `json:"dropAllCapabilities,omitempty"`

	// Specifies the ReadinessProbe to use in order to determine if the container is ready to accept traffic.
	// For more information see https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes
	ReadinessProbe *flds.ProbeField `json:"readinessProbe,omitempty"`

	// Defines the restart policy for pods. If not specified, the default policy set by the framework is used.
	// For details, refer to: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	// +optional
	RestartPolicy *flds.StringField `json:"restartPolicy,omitempty"`

	//  An optional space separated list of references to secrets in the same namespace of the workload, to use for pulling any
	//  images in pods that reference private registries. For details, refer to:
	//     https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets *flds.StringField `json:"imagePullSecrets,omitempty"`
}

func RunaiContainerRules() ContainerFields {
	result := ContainerFields{
		Image: &flds.StringField{
			Rules: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatImageName),
			},
		},

		ImagePullPolicy: &flds.StringField{
			Value: ptr.To(constants.PullPolicyIfNotPresent), // RUNAI DEFAULT
			Rules: &flds.StringRules{
				Options: &[]flds.OptionRule{
					{
						Value: ptr.To(constants.PullPolicyAlways),
						Name:  ptr.To(constants.PullPolicyAlways),
					},
					{
						Value: ptr.To(constants.PullPolicyIfNotPresent),
						Name:  ptr.To(constants.PullPolicyIfNotPresent),
					},
					{
						Value: ptr.To(constants.PullPolicyNever),
						Name:  ptr.To(constants.PullPolicyNever),
					},
				},
			},
		},

		Command: &flds.StringField{
			Rules: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatShellCommand),
			},
		},

		Arguments: &flds.StringField{
			Rules: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatShellCommand),
			},
		},

		Environment: &flds.DictionaryField{
			KeyRules: &flds.StringRules{
				Format:   ptr.To(workload_fields.FormatEnvVar),
				Required: ptr.To(true),
			},
		},

		SupplementalGroups: &flds.StringField{
			Rules: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatGroups),
			},
		},

		RunAsUid: &flds.IntField{
			Rules: &flds.IntRules{
				Min: ptr.To(0),
			},
		},

		RunAsGid: &flds.IntField{
			Rules: &flds.IntRules{
				Min: ptr.To(0),
			},
		},

		Tolerations: &flds.TolerationsField{
			ItemRules: &flds.TolerationsItemRules{
				Operator: &flds.StringRules{
					Required: ptr.To(true),
					Options: &[]flds.OptionRule{
						{
							Value: ptr.To(constants.TolerationsEqual),
							Name:  ptr.To(constants.TolerationsEqual),
						},
						{
							Value: ptr.To(constants.TolerationsExists),
							Name:  ptr.To(constants.TolerationsExists),
						},
					},
				},
				Key: &flds.StringRules{
					Required: ptr.To(true),
				},
				Effect: &flds.StringRules{
					Required: ptr.To(false),
					Options: &[]flds.OptionRule{
						{
							Value: ptr.To(constants.TolerationsNoExecute),
							Name:  ptr.To(constants.TolerationsNoExecute),
						},
						{
							Value: ptr.To(constants.TolerationsNoSchedule),
							Name:  ptr.To(constants.TolerationsNoSchedule),
						},
						{
							Value: ptr.To(constants.TolerationsPreferNoSched),
							Name:  ptr.To(constants.TolerationsPreferNoSched),
						},
					},
				},
				Seconds: &flds.IntRules{
					Min: ptr.To(1),
				},
			},
		},

		Capabilities: &flds.CapabilitiesField{
			ItemRules: &flds.CapabilitiesItemRules{
				Capability: &flds.StringRules{
					Required: ptr.To(true),
					Options: &[]flds.OptionRule{
						capabilityOption("AUDIT_CONTROL"),
						capabilityOption("AUDIT_READ"),
						capabilityOption("AUDIT_WRITE"),
						capabilityOption("BLOCK_SUSPEND"),
						capabilityOption("CHOWN"),
						capabilityOption("DAC_OVERRIDE"),
						capabilityOption("DAC_READ_SEARCH"),
						capabilityOption("FOWNER"),
						capabilityOption("FSETID"),
						capabilityOption("IPC_LOCK"),
						capabilityOption("IPC_OWNER"),
						capabilityOption("KILL"),
						capabilityOption("LEASE"),
						capabilityOption("LINUX_IMMUTABLE"),
						capabilityOption("MAC_ADMIN"),
						capabilityOption("MAC_OVERRIDE"),
						capabilityOption("MKNOD"),
						capabilityOption("NET_ADMIN"),
						capabilityOption("NET_BIND_SERVICE"),
						capabilityOption("NET_BROADCAST"),
						capabilityOption("NET_RAW"),
						capabilityOption("SETGID"),
						capabilityOption("SETFCAP"),
						capabilityOption("SETPCAP"),
						capabilityOption("SETUID"),
						capabilityOption("SYS_ADMIN"),
						capabilityOption("SYS_BOOT"),
						capabilityOption("SYS_CHROOT"),
						capabilityOption("SYS_MODULE"),
						capabilityOption("SYS_NICE"),
						capabilityOption("SYS_PACCT"),
						capabilityOption("SYS_PTRACE"),
						capabilityOption("SYS_RAWIO"),
						capabilityOption("SYS_RESOURCE"),
						capabilityOption("SYS_TIME"),
						capabilityOption("SYS_TTY_CONFIG"),
						capabilityOption("SYSLOG"),
						capabilityOption("WAKE_ALARM"),
					},
				},
			},
		},

		SeccompProfileType: &flds.StringField{
			Value: ptr.To(constants.SeccompProfileTypeRuntimeDefault), // RUNAI DEFAULT
			Rules: &flds.StringRules{
				Options: &[]flds.OptionRule{
					{
						Value: ptr.To(constants.SeccompProfileTypeRuntimeDefault),
						Name:  ptr.To(constants.SeccompProfileTypeRuntimeDefault),
					},
					{
						Value: ptr.To(constants.SeccompProfileTypeUnconfined),
						Name:  ptr.To(constants.SeccompProfileTypeUnconfined),
					},
				},
			},
		},
	}

	return result
}

func capabilityOption(capability string) flds.OptionRule {
	return flds.OptionRule{
		Name:  ptr.To(capability),
		Value: ptr.To(capability),
	}
}
