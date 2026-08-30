package displayer

// Builder
type WorkloadRuntimeConfigBuilder struct {
	config WorkloadRuntimeConfig
}

func NewWorkloadRuntimeConfigBuilder() *WorkloadRuntimeConfigBuilder {
	return &WorkloadRuntimeConfigBuilder{config: WorkloadRuntimeConfig{}}
}

func (b *WorkloadRuntimeConfigBuilder) Image(image string) *WorkloadRuntimeConfigBuilder {
	b.config.Image = &image
	return b
}

func (b *WorkloadRuntimeConfigBuilder) ImagePullPolicy(policy string) *WorkloadRuntimeConfigBuilder {
	b.config.ImagePullPolicy = &policy
	return b
}

func (b *WorkloadRuntimeConfigBuilder) ImagePullSecrets(secrets []string) *WorkloadRuntimeConfigBuilder {
	b.config.ImagePullSecrets = secrets
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Command(command string) *WorkloadRuntimeConfigBuilder {
	b.config.Command = &command
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Arguments(arguments string) *WorkloadRuntimeConfigBuilder {
	b.config.Arguments = &arguments
	return b
}

func (b *WorkloadRuntimeConfigBuilder) WorkingDir(workingDir string) *WorkloadRuntimeConfigBuilder {
	b.config.WorkingDir = &workingDir
	return b
}

func (b *WorkloadRuntimeConfigBuilder) EnvironmentVariables(envVars map[string]string) *WorkloadRuntimeConfigBuilder {
	b.config.EnvironmentVariables = envVars
	return b
}

func (b *WorkloadRuntimeConfigBuilder) NodeType(nodeType string) *WorkloadRuntimeConfigBuilder {
	b.config.NodeType = &nodeType
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Tolerations(tolerations []string) *WorkloadRuntimeConfigBuilder {
	b.config.Tolerations = tolerations
	return b
}

func (b *WorkloadRuntimeConfigBuilder) RunAsUser(uid int64) *WorkloadRuntimeConfigBuilder {
	b.config.RunAsUser = &uid
	return b
}

func (b *WorkloadRuntimeConfigBuilder) RunAsGroup(gid int64) *WorkloadRuntimeConfigBuilder {
	b.config.RunAsGroup = &gid
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Capabilities(capabilities []string) *WorkloadRuntimeConfigBuilder {
	b.config.Capabilities = capabilities
	return b
}

func (b *WorkloadRuntimeConfigBuilder) ReadinessProbe(probe string) *WorkloadRuntimeConfigBuilder {
	b.config.ReadinessProbe = &probe
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Compute(compute *Compute) *WorkloadRuntimeConfigBuilder {
	b.config.Compute = compute
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Storage(storage *Storage) *WorkloadRuntimeConfigBuilder {
	b.config.Storage = storage
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Labels(labels map[string]string) *WorkloadRuntimeConfigBuilder {
	b.config.Labels = labels
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Preemptible(preemptible bool) *WorkloadRuntimeConfigBuilder {
	b.config.Preemptible = &preemptible
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Set(config WorkloadRuntimeConfig) *WorkloadRuntimeConfigBuilder {
	b.config = config
	return b
}

func (b *WorkloadRuntimeConfigBuilder) Build() WorkloadRuntimeConfig {
	return b.config
}

func (b *WorkloadRuntimeConfigBuilder) BuildPtr() *WorkloadRuntimeConfig {
	config := b.config
	return &config
}
