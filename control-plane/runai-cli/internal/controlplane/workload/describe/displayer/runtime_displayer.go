package displayer

// Type definitions
type Runtime struct {
	Preemptible      *bool    `json:"preemptible,omitempty"`
	Images           []string `json:"images,omitempty"`
	ImagePullPolicy  *string  `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	WorkingDir       *string  `json:"workingDir,omitempty"`
	NodeType         *string  `json:"nodeType,omitempty"`
	RunAsUser        *int64   `json:"runAsUser,omitempty"`
	RunAsGroup       *int64   `json:"runAsGroup,omitempty"`
	Capabilities     []string `json:"capabilities,omitempty"`
	Tolerations      []string `json:"tolerations,omitempty"`
	ReadinessProbe   *string  `json:"readinessProbe,omitempty"`
}
