package displayer

import (
	"time"
)

// Builder
type PodBuilder struct {
	pod Pod
}

func NewPodBuilder() *PodBuilder {
	return &PodBuilder{pod: Pod{
		Containers: []Container{},
	}}
}

func (b *PodBuilder) Name(name string) *PodBuilder {
	b.pod.Name = name
	return b
}

func (b *PodBuilder) NodeName(nodeName string) *PodBuilder {
	b.pod.NodeName = &nodeName
	return b
}

func (b *PodBuilder) CompletedAt(completedAt time.Time) *PodBuilder {
	b.pod.CompletedAt = &completedAt
	return b
}

func (b *PodBuilder) CurrentNodePool(currentNodePool string) *PodBuilder {
	b.pod.CurrentNodePool = &currentNodePool
	return b
}

func (b *PodBuilder) K8sPhase(phase string) *PodBuilder {
	b.pod.K8sPhase = phase
	return b
}

func (b *PodBuilder) Containers(containers []Container) *PodBuilder {
	b.pod.Containers = containers
	return b
}

func (b *PodBuilder) AddContainer(container Container) *PodBuilder {
	b.pod.Containers = append(b.pod.Containers, container)
	return b
}

func (b *PodBuilder) Role(role string) *PodBuilder {
	b.pod.Role = &role
	return b
}

func (b *PodBuilder) Build() Pod {
	return b.pod
}
