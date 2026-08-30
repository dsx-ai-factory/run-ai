/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package inference

import (
	"context"
	"fmt"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	loginfw = logger.Log.WithName("reconcile.infw")
)

// Reconciler reconciles a InferenceWorkload object
type Reconciler struct {
	CachedClient  client.Client
	APIReadClient client.Reader
	Scheme        *runtime.Scheme
	Config        *rest.Config
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconcile int, mnnvlSupported bool) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&infv2.InferenceWorkload{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&servingv1.Service{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(controller.InferencePodsFilter()),
		).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconcile,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay),
		}).
		WithEventFilter(controller.InferencePredicate())

	// Watch owned routing resources based on the active routing mode(s).
	// Both can be enabled simultaneously during migration from Ingress to Gateway API.
	if workloadconfig.IngressEnabled() && workloadconfig.IsOpenshift() {
		manager.Owns(&openshiftV1.Route{})
	}
	if workloadconfig.IngressEnabled() && !workloadconfig.IsOpenshift() {
		manager.Owns(&networkingv1.Ingress{})
	}
	if workloadconfig.GatewayEnabled() {
		manager.Owns(&gatewayv1.HTTPRoute{})
	}

	if mnnvlSupported {
		manager.Owns(&nvapi.ComputeDomain{})
	}

	return manager.Complete(r)
}

func NewSimpleReconcile(k8sClient client.Client, apiReadClient client.Reader, config *rest.Config) *Reconciler {
	return &Reconciler{
		CachedClient:  k8sClient,
		APIReadClient: apiReadClient,
		Scheme:        k8sClient.Scheme(),
		Config:        config,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger.IntoContext(ctx, loginfw)

	resourceId := fmt.Sprintf("%s:%s", req.Namespace, req.Name)
	loginfw.Info("reconcile started", "id", resourceId)

	//
	//    obtain the resource being reconciled
	//
	inferenceWorkload := &infv2.InferenceWorkload{}

	err := r.CachedClient.Get(ctx, req.NamespacedName, inferenceWorkload)
	if err != nil {
		//
		//   there was an error to get the resource from k8s
		//
		statusError := err.(*k8sErrors.StatusError)
		if statusError == nil || statusError.Status().Code != 404 {
			loginfw.Error(err, "failed to load the workload", "id", resourceId)
		} else {
			loginfw.Info("workload has been deleted", "id", resourceId)
		}

		//
		//    assume that the resource has been deleted with so just return normally
		//
		return ctrl.Result{}, nil
	}

	var jobWorkload jobs.JobWorkload
	inferenceWorkload.CopyTo(&jobWorkload)
	workloadControllerFactory := NewInferenceWorkloadControllerFactory(r.CachedClient, r.APIReadClient, workloadconfig.IsOpenshift())
	controller, err := workloadControllerFactory.CreateController()
	if err != nil {
		return ctrl.Result{}, err
	}
	result, err := controller.Reconcile(ctx, &jobWorkload, inferenceWorkload)
	if err != nil {
		return ctrl.Result{}, err
	}

	return result, nil
}
