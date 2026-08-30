package resources

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// CleanupDisabledNetworkingResources deletes networking resources that belong to the
// given workload but are of a type that is currently disabled. This handles mode
// transitions (e.g., Ingress -> Gateway API) where orphaned resources of the old
// type would otherwise persist until the workload is deleted.
//
// Only resources matching the provided labels (typically LabelRelease: workloadName)
// in the given namespace are considered. Resources of enabled types are never touched.
func CleanupDisabledNetworkingResources(ctx context.Context, k8sClient client.Client, namespace string, labels map[string]string, isOpenshift bool) error {
	if !config.IngressEnabled() {
		if isOpenshift {
			if err := deleteRoutesByLabels(ctx, k8sClient, namespace, labels); err != nil {
				return fmt.Errorf("failed to cleanup orphaned Routes: %w", err)
			}
		} else {
			if err := deleteIngressesByLabels(ctx, k8sClient, namespace, labels); err != nil {
				return fmt.Errorf("failed to cleanup orphaned Ingresses: %w", err)
			}
		}
	}

	if !config.GatewayEnabled() {
		if err := deleteHTTPRoutesByLabels(ctx, k8sClient, namespace, labels); err != nil {
			return fmt.Errorf("failed to cleanup orphaned HTTPRoutes: %w", err)
		}
	}

	return nil
}

func deleteIngressesByLabels(ctx context.Context, k8sClient client.Client, namespace string, labels map[string]string) error {
	log := logf.FromContext(ctx)
	list := &netv1.IngressList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		return err
	}
	for i := range list.Items {
		log.Info("deleting orphaned networking resource", "kind", constants.KindIngress, "name", list.Items[i].Name)
		if err := k8sClient.Delete(ctx, &list.Items[i]); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func deleteRoutesByLabels(ctx context.Context, k8sClient client.Client, namespace string, labels map[string]string) error {
	log := logf.FromContext(ctx)
	list := &routev1.RouteList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		return err
	}
	for i := range list.Items {
		log.Info("deleting orphaned networking resource", "kind", constants.KindRoute, "name", list.Items[i].Name)
		if err := k8sClient.Delete(ctx, &list.Items[i]); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func deleteHTTPRoutesByLabels(ctx context.Context, k8sClient client.Client, namespace string, labels map[string]string) error {
	log := logf.FromContext(ctx)
	list := &gatewayv1.HTTPRouteList{}
	if err := k8sClient.List(ctx, list, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		if meta.IsNoMatchError(err) {
			log.V(2).Info("HTTPRoute CRD not installed, skipping HTTPRoute cleanup")
			return nil
		}
		return err
	}
	for i := range list.Items {
		log.Info("deleting orphaned networking resource", "kind", constants.KindHTTPRoute, "name", list.Items[i].Name)
		if err := k8sClient.Delete(ctx, &list.Items[i]); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}
