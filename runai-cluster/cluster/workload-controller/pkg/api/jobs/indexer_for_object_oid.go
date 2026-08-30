package jobs

import (
	"context"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func IndexInvolvedOid(mgr manager.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &corev1.Event{},
		workloadcommon.InvolvedObjectUidField,
		func(obj client.Object) []string {
			ev := obj.(*corev1.Event)
			return []string{string(ev.InvolvedObject.UID)}
		},
	)

	if err != nil {
		return err
	}

	return nil
}
