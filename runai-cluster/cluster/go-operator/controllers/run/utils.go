package run

import (
	"context"
	"fmt"
	"net/http"
	"time"

	kube_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/common/utils/kubernetes"

	version_utils "github.com/run-ai/runai/runai-cluster/cluster/common/utils/version"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kube_client "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	rss "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type resources struct {
	CreateOrUpdate []operands.Resource
	Delete         []operands.Resource
}

const (
	minimumKubernetesVersion           = "1.30.0"
	minimumKubernetesVersionConstraint = "1.30.0-0"
)

// noRequeue is a shorthand function
func noRequeue() ctrl.Result {
	return requeueAfter(time.Minute)
}

// requeue is a shorthand function
func requeue() ctrl.Result {
	return ctrl.Result{
		Requeue: true,
	}
}

// requeueAfter is a shorthand function
func requeueAfter(after time.Duration) ctrl.Result {
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: after,
	}
}

func ResourcesForOperands(ctx context.Context, operandsSlice []operands.Operand) (resources, error) {
	resourceSlices := resources{}
	errs := &me.MultiError{}

	type result struct {
		toCreateOrUpdate []operands.Resource
		toDelete         []operands.Resource
		errs             []error
	}

	resultsCh := make(chan result, len(operandsSlice))

	for _, operand := range operandsSlice {
		operand := operand
		go func() {
			var res result
			operandResources := operand.Resources()
			if operand.Enabled() {
				enabled := utils.GetEnabled(&operandResources)
				disabled := utils.GetDisabled(&operandResources)
				localErrs := &me.MultiError{}
				addResources(ctx, enabled, localErrs, &res.toCreateOrUpdate)
				addResources(ctx, disabled, localErrs, &res.toDelete)
				if err := localErrs.ErrorOrNil(); err != nil {
					res.errs = append(res.errs, err)
				}
			} else {
				localErrs := &me.MultiError{}
				addResources(ctx, operandResources, localErrs, &res.toDelete)
				if err := localErrs.ErrorOrNil(); err != nil {
					res.errs = append(res.errs, err)
				}
			}
			resultsCh <- res
		}()
	}

	for i := 0; i < len(operandsSlice); i++ {
		res := <-resultsCh
		resourceSlices.CreateOrUpdate = append(resourceSlices.CreateOrUpdate, res.toCreateOrUpdate...)
		resourceSlices.Delete = append(resourceSlices.Delete, res.toDelete...)
		for _, err := range res.errs {
			errs.Append(err)
		}
	}

	return resourceSlices, errs.ErrorOrNil()
}

func addResources(ctx context.Context, operandResources []operands.Resource, errs *me.MultiError, resourceList *[]operands.Resource) {
	if operandResources == nil {
		return
	}

	type res struct {
		resource operands.Resource
		err      error
	}

	resCh := make(chan res, len(operandResources))

	for _, resource := range operandResources {
		resource := resource
		go func() {
			obj, err := resource.Object(ctx)
			if obj != nil {
				resCh <- res{resource: resource, err: err}
			} else if err != nil {
				resCh <- res{resource: nil, err: err}
			} else {
				// neither obj nor err, do not send
				resCh <- res{resource: nil, err: nil}
			}
		}()
	}

	for i := 0; i < len(operandResources); i++ {
		r := <-resCh
		if r.err != nil {
			errs.Append(r.err)
		} else if r.resource != nil {
			*resourceList = append(*resourceList, r.resource)
		}
	}
}

func checkDependencies(ctx context.Context, kubeClient kube_client.HybridClient, runaiConfig *runaiv1.RunaiConfig) error {
	format := "could not find %s in the cluster: %v"
	multiErr := &me.MultiError{}

	_, err := controller_utils.GetPrometheusOperatorNamespace(ctx, kubeClient)
	if err != nil {
		multiErr.Append(fmt.Errorf(format, consts.PrometheusOperatorDeploymentName, err))
	}

	err = controller_utils.CheckGpuOperatorDependency(ctx, kubeClient)
	if err != nil {
		multiErr.Append(fmt.Errorf(format, consts.GPUOperatorDeploymentName, err))
	}

	gatewayEnabled := runaiConfig.Spec.Global.GatewayAPI != nil &&
		runaiConfig.Spec.Global.GatewayAPI.Enabled != nil &&
		*runaiConfig.Spec.Global.GatewayAPI.Enabled

	ingressEnabled := runaiConfig.Spec.Global.Ingress != nil &&
		runaiConfig.Spec.Global.Ingress.Enabled != nil &&
		*runaiConfig.Spec.Global.Ingress.Enabled

	if !gatewayEnabled && !ingressEnabled {
		multiErr.Append(fmt.Errorf("no routing mechanism configured: both ingress and gatewayAPI are disabled"))
	}

	if gatewayEnabled {
		gwStatus, gwErr := gatewayAPIStatus(ctx, kubeClient.Uncached(), runaiConfig)
		if gwErr != nil {
			multiErr.Append(fmt.Errorf(format, "gateway-api", gwErr))
		}
		if !gwStatus.Ready {
			multiErr.Append(fmt.Errorf(format, "gateway-api", gwStatus.Reasons))
		}
	}

	if ingressEnabled {
		ingressStatus, ingressErr := rss.IngressStatus(ctx, kubeClient, runaiConfig)
		if ingressErr != nil {
			multiErr.Append(fmt.Errorf(format, "ingress-controller", ingressErr))
		}
		if !ingressStatus.Ready {
			multiErr.Append(fmt.Errorf(format, "ingress-controller", ingressStatus.Reasons))
		}
	}

	if !runaiConfig.Spec.Internal.Development.Headless {
		err = CheckConnectivity(runaiConfig)
		if err != nil {
			multiErr.Append(err)
		}
	}

	if multiErr.Empty() {
		return nil
	}

	return multiErr
}

func checkKubeVersion(dClient discovery.DiscoveryInterface) error {
	kubeVersion, err := kube_utils.GetKubernetesClusterVersion(dClient)
	if err != nil {
		return fmt.Errorf("failed to get k8s version: %w", err)
	}

	isKubeVersionSupported, err := version_utils.VersionGreaterEqualThan(kubeVersion, minimumKubernetesVersionConstraint)
	if err != nil {
		return fmt.Errorf("failed to check server version: %w", err)
	}

	if !isKubeVersionSupported {
		return fmt.Errorf("kubernetes version %s is not supported; minimum required version is %s",
			kubeVersion, minimumKubernetesVersion)
	}

	return nil
}

func checkKubeVersions(local, remote discovery.DiscoveryInterface) error {
	if err := checkKubeVersion(local); err != nil {
		return fmt.Errorf("local cluster: %w", err)
	}
	if err := checkKubeVersion(remote); err != nil {
		return fmt.Errorf("remote cluster: %w", err)
	}
	return nil
}

// isVAPSupported checks whether the cluster serves ValidatingAdmissionPolicy in
// admissionregistration.k8s.io/v1 (GA in 1.30). The group version alone is not
// enough — webhook configurations are served as v1 on older clusters too.
// Watching an unserved kind fails cache sync and crashes the manager.
func isVAPSupported(dClient discovery.DiscoveryInterface) (bool, error) {
	apiList, err := dClient.ServerResourcesForGroupVersion(admissionv1.SchemeGroupVersion.String())
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to discover %s resources: %w", admissionv1.SchemeGroupVersion, err)
	}
	for _, apiResource := range apiList.APIResources {
		if apiResource.Kind == "ValidatingAdmissionPolicy" {
			return true, nil
		}
	}
	return false, nil
}

func CheckConnectivity(runaiConfig *runaiv1.RunaiConfig) error {
	// TODO move to struct and mock the client for tests
	validate := func(addr, method string) error {
		req, err := http.NewRequest(method, addr, nil)
		if err != nil {
			return fmt.Errorf("connection failed to : %s err: %s", addr, err.Error())
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("connection failed to : %s err: %s", addr, err.Error())
		}
		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("connection failed to : %s status: %v", addr, resp.StatusCode)
		}
		return nil
	}

	finalError := &me.MultiError{}
	finalError.Append(validate(runaiConfig.Spec.Internal.Global.ControlPlane.URL, "HEAD"))

	if finalError.Empty() {
		log.Log.Info("All conectivity check passed succesfully")
		return nil
	}
	return finalError
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func gatewayAPIStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) (operands.Status, error) {
	global := runaiConfig.Spec.Global
	if global == nil || global.GatewayAPI == nil || global.GatewayAPI.Enabled == nil || !*global.GatewayAPI.Enabled {
		return operands.ReadyStatus(), nil
	}

	gw := global.GatewayAPI
	gateway := &gwv1.Gateway{}
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      gw.Name,
		Namespace: gw.Namespace,
	}, gateway)
	if err != nil {
		if meta.IsNoMatchError(err) {
			return operands.NotReadyStatus("Gateway API CRDs not installed"), nil
		}
		if errors.IsNotFound(err) {
			return operands.NotReadyStatus(fmt.Sprintf("Gateway %s/%s not found", gw.Namespace, gw.Name)), nil
		}
		return operands.NotReadyStatus(), err
	}

	return operands.ReadyStatus(), nil
}
