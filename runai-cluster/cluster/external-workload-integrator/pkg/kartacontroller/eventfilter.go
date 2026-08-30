package kartacontroller

import (
	"context"

	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (rc *KartaController) MapClusterRoleToKartaEvent(ctx context.Context, _ client.Object) []reconcile.Request {
	// for ClusterRole events, Reconcile all Kartas -
	// it will be pretty hard to not possible to know which Karta uses which GVKs
	kartas := &kartav1alpha1.KartaList{}
	err := rc.Client.List(ctx, kartas)
	if err != nil {
		log.Error().Msgf("Failed to list Kartas: %v", err)
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, 0, len(kartas.Items))
	for _, karta := range kartas.Items {
		requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKey{Name: karta.Name}})
	}
	return requests
}

func (rc *KartaController) ControllerPredicateFuncs() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return rc.filterEvent(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return rc.filterEvent(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return rc.filterEvent(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return rc.filterEvent(e.Object)
		},
	}
}

func (rc *KartaController) filterEvent(obj client.Object) bool {
	switch obj.(type) {
	case *rbacv1.ClusterRole:
		return rc.filterClusterRolesByName(obj.(*rbacv1.ClusterRole))
	}
	return true

}

func (rc *KartaController) filterClusterRolesByName(cr *rbacv1.ClusterRole) bool {
	return cr.Name == rc.clusterRoleReadName || cr.Name == rc.clusterRoleCrudName
}
