package enrichers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
)

const (
	KubeMetadataEnricherName = "kube_metadata"
)

type objectMetadata struct {
	labels      map[string]string
	annotations map[string]string
}

type KubeMetadataEnricherConfig struct {
	// Maps kubernetes object labels to event properties
	Labels map[string]string `mapstructure:"labels" json:"labels"`
	// Maps kubernetes object annotations to event properties
	Annotations map[string]string `mapstructure:"annotations" json:"annotations"`
}

type KubeMetadataEnricher struct {
	kubeRClient client.Reader
	config      KubeMetadataEnricherConfig
}

func NewKubeMetadataEnricher(kubeRClient client.Reader, config KubeMetadataEnricherConfig) *KubeMetadataEnricher {
	return &KubeMetadataEnricher{
		kubeRClient: kubeRClient,
		config:      config,
	}
}

func (ce *KubeMetadataEnricher) Enrich(ctx context.Context, event *events.Event) error {
	kubeEvent, ok := event.RawEvent.(*eventsv1.Event)
	if !ok {
		return nil
	}

	regardingObjMetadata, err := ce.getObjectMetadata(ctx, kubeEvent.Regarding)
	if err != nil {
		return fmt.Errorf("enricher %v failed to get labels of regarding object %v: %w", KubeMetadataEnricherName, kubeEvent.Regarding, err)
	}

	addPropertiesToEvent(event, ce.config.Annotations, regardingObjMetadata.annotations)
	addPropertiesToEvent(event, ce.config.Labels, regardingObjMetadata.labels)

	return nil
}

func (ce *KubeMetadataEnricher) getObjectMetadata(ctx context.Context, objectRef corev1.ObjectReference) (*objectMetadata, error) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": objectRef.APIVersion,
			"kind":       objectRef.Kind,
			"metadata": map[string]interface{}{
				"name":      objectRef.Name,
				"namespace": objectRef.Namespace,
			},
		},
	}

	err := ce.kubeRClient.Get(ctx, client.ObjectKey{
		Namespace: objectRef.Namespace,
		Name:      objectRef.Name,
	}, obj)
	if err != nil {
		return nil, err
	}

	return &objectMetadata{
		labels:      obj.GetLabels(),
		annotations: obj.GetAnnotations(),
	}, nil
}

func addPropertiesToEvent(event *events.Event, requiredProps map[string]string, props map[string]string) {
	for requirementKey, propKey := range requiredProps {
		if value, ok := props[requirementKey]; ok {
			event.Properties[propKey] = value
		}
	}
}
