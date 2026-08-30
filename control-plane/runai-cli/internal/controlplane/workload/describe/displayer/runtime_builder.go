package displayer

// Builder
type RuntimeBuilder struct {
	runtime Runtime
}

func NewRuntimeBuilder() *RuntimeBuilder {
	return &RuntimeBuilder{runtime: Runtime{}}
}

func (b *RuntimeBuilder) Preemptible(preemptible bool) *RuntimeBuilder {
	b.runtime.Preemptible = &preemptible
	return b
}

func (b *RuntimeBuilder) Images(images []string) *RuntimeBuilder {
	b.runtime.Images = images
	return b
}

func (b *RuntimeBuilder) AddImage(image string) *RuntimeBuilder {
	b.runtime.Images = append(b.runtime.Images, image)
	return b
}

func (b *RuntimeBuilder) ImagePullPolicy(policy string) *RuntimeBuilder {
	b.runtime.ImagePullPolicy = &policy
	return b
}

func (b *RuntimeBuilder) ImagePullSecrets(secrets []string) *RuntimeBuilder {
	b.runtime.ImagePullSecrets = secrets
	return b
}

func (b *RuntimeBuilder) WorkingDir(workingDir string) *RuntimeBuilder {
	b.runtime.WorkingDir = &workingDir
	return b
}

func (b *RuntimeBuilder) NodeType(nodeType string) *RuntimeBuilder {
	b.runtime.NodeType = &nodeType
	return b
}

func (b *RuntimeBuilder) RunAsUser(uid int64) *RuntimeBuilder {
	b.runtime.RunAsUser = &uid
	return b
}

func (b *RuntimeBuilder) RunAsGroup(gid int64) *RuntimeBuilder {
	b.runtime.RunAsGroup = &gid
	return b
}

func (b *RuntimeBuilder) Capabilities(capabilities []string) *RuntimeBuilder {
	b.runtime.Capabilities = capabilities
	return b
}

func (b *RuntimeBuilder) Tolerations(tolerations []string) *RuntimeBuilder {
	b.runtime.Tolerations = tolerations
	return b
}

func (b *RuntimeBuilder) ReadinessProbe(probe string) *RuntimeBuilder {
	b.runtime.ReadinessProbe = &probe
	return b
}

func (b *RuntimeBuilder) Set(runtime Runtime) *RuntimeBuilder {
	b.runtime = runtime
	return b
}

func (b *RuntimeBuilder) Build() Runtime {
	return b.runtime
}
