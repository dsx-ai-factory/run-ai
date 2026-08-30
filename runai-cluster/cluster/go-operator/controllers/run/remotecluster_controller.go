package run

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/installation/util"
	cu "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kube_client "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"gomodules.xyz/jsonpatch/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RemoteClusterReconciler reconciles a RemoteCluster object
// This is a stub for a new resource reconciler, modeled after RunAIConfigReconciler.
type RemoteClusterReconciler struct {
	runaiReconciler

	// remoteClusterDiscoveryClient is the discovery client for the remote cluster this reconciler's manager watches
	remoteClusterDiscoveryClient discovery.DiscoveryInterface
}

// NewRunAIConfigReconciler creates a new RunAIConfigReconciler
func NewRemoteClusterReconciler(localClusterClient kube_client.HybridClient,
	remoteClusterDiscoveryClient discovery.DiscoveryInterface, verbosity int) *RemoteClusterReconciler {
	reconciler := &RemoteClusterReconciler{
		runaiReconciler: runaiReconciler{
			verbosity:          verbosity,
			localClusterClient: localClusterClient,
		},
		remoteClusterDiscoveryClient: remoteClusterDiscoveryClient,
	}

	return reconciler
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO: Implement reconciliation logic for RemoteCluster.
func (r *RemoteClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).V(r.verbosity)

	// No matter what kind of event we get, we need to trigger a reconciliation in the configuration cluster
	// to ensure that the remote cluster is in sync with the configuration cluster.
	logger.Info("triggering reconciliation for RunaiConfig")
	if err := triggerReconciliationForRunaiConfig(ctx, r.localClusterClient); err != nil {
		logger.Error(err, "failed triggering reconciliation for RunaiConfig")
		return ctrl.Result{}, err
	}
	logger.Info("reconciliation for RunaiConfig triggered")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// TODO: Register the controller for RemoteCluster and set up any watches as needed.
func (r *RemoteClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr)

	// probe the remote cluster (the one this manager watches), not the local one
	isOpenshift, err := util.IsOpenshift(context.Background(), mgr.GetAPIReader())
	if err != nil {
		return fmt.Errorf("remote cluster reconciler setup: failed to determine if running on openshift or not: %w", err)
	}

	vapSupported, err := isVAPSupported(r.remoteClusterDiscoveryClient)
	if err != nil {
		return fmt.Errorf("remote cluster reconciler setup: failed to check ValidatingAdmissionPolicy support: %w", err)
	}

	addWatchesToBuilder(builder, nil, append(watchedResources(vapSupported), ownedResources(isOpenshift)...))

	return builder.Named("RemoteClusterReconciler").Complete(r)
}

func mapResourceToRunaiConfigEvent(_ context.Context, obj client.Object) []reconcile.Request {
	if !isOwnedByDefaultRunaiConfig(obj) {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      runaiv1.DefaultName,
				Namespace: runaiv1.DefaultNamespace,
			},
		},
	}
}

func isOwnedByDefaultRunaiConfig(obj client.Object) bool {
	return cu.IsOwnerOf(metav1.GroupKind{
		Group: runaiv1.GroupVersion.Group,
		Kind:  runaiv1.RunaiConfigKind,
	}, runaiv1.DefaultNamespace, runaiv1.DefaultName, obj.GetAnnotations())
}

func triggerReconciliationForRunaiConfig(ctx context.Context, k8s client.Client) error {
	runaiConfigToTrigger := &runaiv1.RunaiConfig{}
	err := k8s.Get(ctx, types.NamespacedName{
		Namespace: runaiv1.DefaultNamespace,
		Name:      runaiv1.DefaultName,
	}, runaiConfigToTrigger)
	if err != nil {
		return err
	}

	if runaiConfigToTrigger.Annotations == nil {
		runaiConfigToTrigger.Annotations = map[string]string{}
	}

	runaiConfigToTrigger.Annotations[runaiv1.ReconciliationTriggerRequestKey] = uuid.New().String()

	patch := jsonpatch.NewOperation("add", "/metadata/annotations", runaiConfigToTrigger.Annotations)

	return k8s.Patch(ctx, runaiConfigToTrigger, client.RawPatch(types.JSONPatchType,
		[]byte(fmt.Sprintf("[%s]", patch.Json()))))
}
