package builders

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dinfv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func TestServingPortBuildService_NimLabelWithLeaderLabel(t *testing.T) {
	b := NewServingPortBuilder(initClientWithResources(t, namespace), false)

	workload := &dinfv1.DistributedInferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dinfw-example",
			Namespace: "runai-team-a",
			Labels: map[string]string{
				lwsv1.SetNameLabelKey:      "lws-name",
				constants.LabelNIMWorkload: "true",
			},
		},
	}

	servingPort := &flds.ServingPortItemValue{Port: ptr.To(8080)}

	svc, err := b.buildService(workload, servingPort)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Expect NIM label to be set to "true"
	val, exists := svc.ObjectMeta.Labels[constants.LabelNIMWorkload]
	assert.True(t, exists)
	assert.Equal(t, "true", val)
}

func TestServingPortBuildService_NoNimLabelWithLeaderLabel(t *testing.T) {
	b := NewServingPortBuilder(initClientWithResources(t, namespace), false)

	workload := &dinfv1.DistributedInferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dinfw-example",
			Namespace: "runai-team-a",
			Labels: map[string]string{
				lwsv1.SetNameLabelKey: "lws-name",
			},
		},
	}

	servingPort := &flds.ServingPortItemValue{Port: ptr.To(8080)}

	svc, err := b.buildService(workload, servingPort)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// NIM label absents: should not be added
	_, exists := svc.ObjectMeta.Labels[constants.LabelNIMWorkload]
	assert.False(t, exists)
}

func TestServingPortBuildService_ServicePortNameIsCorrect(t *testing.T) {
	b := NewServingPortBuilder(initClientWithResources(t, namespace), false)

	workload := &dinfv1.DistributedInferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dinfw-example",
			Namespace: "runai-team-a",
			Labels: map[string]string{
				lwsv1.SetNameLabelKey: "lws-name",
			},
		},
	}

	servingPort := &flds.ServingPortItemValue{Port: ptr.To(8080)}

	svc, err := b.buildService(workload, servingPort)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Verify that the service has exactly one port
	assert.Len(t, svc.Spec.Ports, 1, "Service should have exactly one port")

	// Verify that the service port name is set to servingPortName constant
	servicePort := svc.Spec.Ports[0]
	assert.Equal(t, "serving-port", servicePort.Name, "Service port name should be 'serving-port'")
}
