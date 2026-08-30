package controller_utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"go.uber.org/multierr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/installation/util"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
)

// ValidateComponentNamespace checks that the given component exists in the expected namespace.
func ValidateComponentNamespace(ctx context.Context, kbClient client.Client, component, componentBaseName, expectedNamespace string) error {
	componentNameRegex := regexp.MustCompile(fmt.Sprintf("%s.*%s", component, componentBaseName))
	componentDisplayName := fmt.Sprintf("%s %s", component, componentBaseName)
	_, err := GetComponentNamespace(ctx, kbClient, componentNameRegex, componentDisplayName, expectedNamespace)
	return err
}

func GetComponentNamespace(ctx context.Context, k8sReader client.Reader, componentNameRegex *regexp.Regexp, componentDisplayName, expectedNamespace string) (string, error) {
	namespacesSet := map[string]struct{}{}

	var listOptions []client.ListOption
	if expectedNamespace != "" {
		listOptions = append(listOptions, client.InNamespace(expectedNamespace))
	}

	// A component gateway may be deployed either as a Deployment (Run:ai bundled layout)
	// or a DaemonSet (standard upstream layout, e.g. Kourier's 3scale-kourier-gateway).
	// Both are valid, so match the component name against both kinds.
	deploymentsList := &appsv1.DeploymentList{}
	if err := k8sReader.List(ctx, deploymentsList, listOptions...); err != nil {
		return "", err
	}
	for _, deployment := range deploymentsList.Items {
		if componentMatches(deployment.Name, deployment.Namespace, componentNameRegex, expectedNamespace) {
			namespacesSet[deployment.Namespace] = struct{}{}
		}
	}

	daemonSetsList := &appsv1.DaemonSetList{}
	if err := k8sReader.List(ctx, daemonSetsList, listOptions...); err != nil {
		return "", err
	}
	for _, daemonSet := range daemonSetsList.Items {
		if componentMatches(daemonSet.Name, daemonSet.Namespace, componentNameRegex, expectedNamespace) {
			namespacesSet[daemonSet.Namespace] = struct{}{}
		}
	}

	if len(namespacesSet) == 0 {
		return "", fmt.Errorf("%s not found in the cluster", componentDisplayName)
	}

	if len(namespacesSet) > 1 {
		return "", fmt.Errorf("%s is installed in multiple namespaces: %v",
			componentDisplayName, namespacesSet)
	}

	// Return the only namespace in the set
	ns := ""
	for k := range namespacesSet {
		ns = k
		break
	}
	return ns, nil
}

func componentMatches(name, namespace string, componentNameRegex *regexp.Regexp, expectedNamespace string) bool {
	if expectedNamespace != "" && expectedNamespace != namespace {
		return false
	}
	return componentNameRegex.MatchString(name)
}

func GetDCGMExporterService(ctx context.Context, k8sReader client.Reader) (*v1.Service, error) {
	svcList := &v1.ServiceList{}
	var svcs []v1.Service

	err := k8sReader.List(ctx, svcList, client.MatchingLabels{"app": "nvidia-dcgm-exporter"})
	if err != nil {
		return nil, err
	}

	svcs = append(svcs, svcList.Items...)

	svcList = &v1.ServiceList{}
	err = k8sReader.List(ctx, svcList, client.MatchingLabels{"app": "dcgm-exporter"})
	if err != nil {
		return nil, err
	}

	svcs = append(svcs, svcList.Items...)

	svcList = &v1.ServiceList{}
	err = k8sReader.List(ctx, svcList, client.MatchingLabels{"app.kubernetes.io/name": "nvidia-dcgm-exporter"})
	if err != nil {
		return nil, err
	}

	svcs = append(svcs, svcList.Items...)

	svcList = &v1.ServiceList{}
	err = k8sReader.List(ctx, svcList, client.MatchingLabels{"app.kubernetes.io/name": "dcgm-exporter"})
	if err != nil {
		return nil, err
	}

	svcs = append(svcs, svcList.Items...)

	if len(svcs) == 0 {
		return nil, fmt.Errorf("dcgm-exporter service not found in the cluster")
	}

	if len(svcs) > 1 {
		return nil, fmt.Errorf("found more than one dcgm-exporter service in the cluster")
	}

	return &svcs[0], nil
}

func GetPrometheusOperatorNamespace(ctx context.Context, k8sReader client.Reader) (string, error) {
	ocp, err := util.IsOpenshift(ctx, k8sReader)
	if err != nil {
		return "", err
	} else if ocp {
		return consts.OpenshiftMonitoringNamespace, nil
	}

	return GetComponentNamespace(ctx, k8sReader, regexp.MustCompile("(prometheus|rancher-monitoring).*operator"), "Prometheus Operator", "")
}

func VerifyNvidiaStackExists(ctx context.Context, k8sReader client.Reader) error {
	_, err := GetDCGMExporterService(ctx, k8sReader)
	return err
}

func VerifyNvidiaStackExistsInRightState(svc *v1.Service) (error, string) {
	resBodyStr, err := getMetricsFromExporterService(svc)
	if err != nil {
		return err, ""
	}

	errorToReturn := getMissingMetrics(resBodyStr)
	if len(errorToReturn) > 0 {
		return nil, strings.Join(errorToReturn, ", ")
	}

	return nil, ""
}

func getMissingMetrics(resBodyStr string) []string {
	errorToReturn := []string{}
	if !strings.Contains(resBodyStr, "DCGM_FI_DEV_FB_USED") {
		errorToReturn = append(errorToReturn, "DCGM_FI_DEV_FB_USED")
	}
	if !strings.Contains(resBodyStr, "DCGM_FI_DEV_FB_FREE") {
		errorToReturn = append(errorToReturn, "DCGM_FI_DEV_FB_FREE")
	}
	if !strings.Contains(resBodyStr, "DCGM_FI_DEV_GPU_UTIL") {
		errorToReturn = append(errorToReturn, "DCGM_FI_DEV_GPU_UTIL")
	}
	return errorToReturn
}

func getMetricsFromExporterService(svc *v1.Service) (string, error) {
	res, err := http.Get("http://" + svc.Spec.ClusterIP + ":" + strconv.Itoa(int(svc.Spec.Ports[0].Port)) + "/metrics")
	if err != nil {
		return "", fmt.Errorf("failed to connect to dcgm-exporter service in the cluster")
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to parse dcgm-exporter service in the cluster metrics content")
	}

	return string(resBody), nil
}

func GetDependencyStatus(ctx context.Context, kbClient client.Client, component, componentBaseName, expectedNamespace string, gvks []schema.GroupVersionKind, runaiNs *string) runv1.PublicStatus {
	var errs error

	err := ValidateComponentNamespace(ctx, kbClient, component, componentBaseName, expectedNamespace)
	dependency := runv1.PublicStatus{
		Available: runv1.AvailableStatus(true),
	}
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	ns := v1.NamespaceDefault
	if runaiNs != nil {
		ns = *runaiNs
	}

	for _, gvk := range gvks {
		err := kbClient.Create(ctx, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": gvk.GroupVersion().String(),
				"kind":       gvk.Kind,
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": ns,
				},
			},
		}, &client.CreateOptions{
			DryRun: []string{metav1.DryRunAll},
		})

		if err != nil && !kerrors.IsBadRequest(err) && !kerrors.IsAlreadyExists(err) {
			// In case the gvk already exists on the client's RESTMappers, but there is no such gvk in the cluster (e.g. the CRD was deleted during the operator's runtime)
			// the server error is not a StatusError nor an UnexpectedObjectError, so we need to check the error message string directly.
			if meta.IsNoMatchError(err) || strings.Contains(err.Error(), "the server could not find the requested resource") {
				errs = multierr.Append(errs, fmt.Errorf("resource (%s) not found on the cluster", gvk.String()))
			} else {
				errs = multierr.Append(errs, err)
			}
		}
	}

	if errs != nil {
		dependency.Available = runv1.AvailableStatus(false)
		dependency.Reason = errs.Error()
	}

	return dependency
}

func CheckGpuOperatorDependency(ctx context.Context, kubeClient client.Client) error {
	if err := VerifyNvidiaStackExists(ctx, kubeClient); err != nil {
		return fmt.Errorf("could not find NVIDIA GPU Stack in the cluster: %v", err)
	}
	return nil
}
