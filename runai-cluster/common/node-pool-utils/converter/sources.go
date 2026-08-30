package converter

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AffinityConversionErrorHandling bool

const (
	IgnoreAffinityConversionErrors AffinityConversionErrorHandling = true
	ReturnAffinityConversionErrors AffinityConversionErrorHandling = false
)

type AffinitySource struct {
	Affinity               *corev1.Affinity
	ReaderClient           client.Reader
	IgnoreConversionErrors bool
}

func (s *AffinitySource) HasValue() bool {
	return s.Affinity != nil && s.ReaderClient != nil
}

// NodePoolsSources holds every place a requested node pool list can come from.
type NodePoolsSources struct {
	PodAnnotations map[string]string
	AffinitySource AffinitySource
	Project        *kaiv1alpha1.Project
}

func NewNodePoolsSources(affinity *corev1.Affinity, readerClient client.Reader, affinityConversionErrorsHandling AffinityConversionErrorHandling, podAnnotations map[string]string, project *kaiv1alpha1.Project) NodePoolsSources {
	return NodePoolsSources{
		PodAnnotations: podAnnotations,
		AffinitySource: AffinitySource{
			Affinity:               affinity,
			ReaderClient:           readerClient,
			IgnoreConversionErrors: bool(affinityConversionErrorsHandling),
		},
		Project: project,
	}
}
