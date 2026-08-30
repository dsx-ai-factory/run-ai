package requested_nodepools_converter

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/common/node-pool-utils/converter"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RequestedNodePoolsConverter struct {
	Client      client.Client
	identifiers converter.NodePoolIdentifiers
}

func NewRequestedNodePoolsConverter(cachedClient client.Client, identifiers converter.NodePoolIdentifiers) *RequestedNodePoolsConverter {
	return &RequestedNodePoolsConverter{
		Client:      cachedClient,
		identifiers: identifiers,
	}
}

func (rnpc *RequestedNodePoolsConverter) GetRequestedNodePoolsForPodGroup(ctx context.Context, podGroupMeta metav1.ObjectMeta, podGroupPods *corev1.PodList) ([]string, error) {
	podFromPodGroup, err := rnpc.getPodFromPodGroup(podGroupMeta, podGroupPods)
	if err != nil {
		return []string{}, err
	}

	nodePoolsSources := converter.NewNodePoolsSources(podFromPodGroup.Spec.Affinity, rnpc.Client, converter.ReturnAffinityConversionErrors, podFromPodGroup.Annotations, nil)

	return converter.GetRequestedNodePools(ctx, podGroupMeta, "podGroup", rnpc.identifiers, nodePoolsSources)
}

func (rnpc *RequestedNodePoolsConverter) getPodFromPodGroup(podGroupMeta metav1.ObjectMeta, podGroupPods *corev1.PodList) (*corev1.Pod, error) {
	if podGroupPods == nil || len(podGroupPods.Items) == 0 {
		err := fmt.Errorf("didn't find any pods for pod group <%s/%s>", podGroupMeta.Namespace, podGroupMeta.Name)
		return nil, err
	}

	return &podGroupPods.Items[0], nil
}
