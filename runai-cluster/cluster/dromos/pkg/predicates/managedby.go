package predicates

import (
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ManagedByPredicate filters WorkloadRunner events by managedBy
func ManagedByPredicate(want string) predicate.Predicate {
	isSelected := func(obj client.Object) bool {
		wlr, ok := obj.(*dromosv1alpha1.WorkloadRunner)
		if !ok {
			return true
		}
		return wlr.Spec.ManagedBy == want
	}
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool { return isSelected(e.Object) },
		UpdateFunc: func(e event.UpdateEvent) bool { return isSelected(e.ObjectNew) },
	}
}
