/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RunaiInferenceSecretName = "runai-cluster-inference-tls"
	KnativeServingNamespace  = "knative-serving"
	KnativeServingStepName   = "Installing Knative Serving"
)

var (
	requiredServingDeployments = []string{
		"controller",
		"webhook",
		"autoscaler",
		"activator",
		// Note: "autoscaler-hpa" is now installed separately by knative_serving_hpa_step.go
		"net-kourier-controller",
		"3scale-kourier-gateway",
	}
)

type KnativeServingInstallStep struct {
	k8sClient  k8s.K8sClients
	crdChecker CRDChecker
	version    string
	*step.RunAIInstallStepImpl
}

func NewKnativeServingInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	k := &KnativeServingInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(KnativeServingStepName),
		crdChecker:           &DefaultCRDChecker{},
		k8sClient:            k8sClient,
		version:              version,
	}
	k.SetStatusEvaluator(k.EvaluateStatus)
	return k
}

func (k *KnativeServingInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if k.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("[Prereq] Checking if Knative Serving Service CRD exists...")

	timeout := k.GetTimeout()
	installed := k.crdChecker.IsCRDInstalled(k.k8sClient.GetConfig(), "services.serving.knative.dev")
	if !installed {
		elapsed := k.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] Knative Serving CRD not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Knative Serving CRD not found")
		}
		return
	}

	// Check if Knative Serving custom resources are available
	gvk := schema.GroupVersionKind{
		Group:   "serving.knative.dev",
		Version: "v1",
		Kind:    "Service",
	}
	_, crErr := HasCustomResources(k.k8sClient.GetClient(), gvk)
	if crErr != nil {
		elapsed := k.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] Knative serving custom resources not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Error checking Knative serving custom resources: %v", crErr)
		}
		return
	}

	logging.Infof("[Prereq] Checking if Knative Serving deployments are installed...")
	installed, err := isKnativeServingFullyUp(context.Background(), k.k8sClient.GetClient())
	if err != nil {
		elapsed := k.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] Knative Serving not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Knative Serving not found in cluster")
		}
		return
	}
	if !installed {
		elapsed := k.GetElapsedTime()
		if elapsed > 0 {
			logging.Infof("[Prereq] Knative Serving deployments not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Knative Serving deployments not found in cluster")
		}
		return
	}
	logging.Infof("[Prereq] Knative Serving is installed")
	k.SetStatus(model.Installed, nil)
}

func (k *KnativeServingInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing Knative Serving %s...", k.version)
	k.SetStatus(model.Installing, nil)
	wildcardFQDN := runaiData.GetInputValue(configinputs.InstallInputWildcardFQDN)
	// Create dynamic client once
	logging.Infof("[Prereq] Creating dynamic client for Knative resources...")
	dyn, err := dynamic.NewForConfig(k.k8sClient.GetConfig())
	if err != nil {
		fullError := fmt.Errorf("failed to create dynamic client: %v", err)
		logging.Errorf("[Prereq - %s ]  %v", KnativeServingStepName, fullError)
		k.SetStatus(model.Failed, fullError)
		return fullError
	}

	// Step 1: Apply KnativeServing configuration
	logging.Infof("[Prereq] Step 1: Applying KnativeServing configuration with wildcard FQDN: %s...", wildcardFQDN)
	if err := applyKnativeServing(dyn, wildcardFQDN); err != nil {
		fullError := fmt.Errorf("failed to apply KnativeServing: %v", err)
		logging.Errorf("[Prereq - %s ]  %v", KnativeServingStepName, fullError)
		k.SetStatus(model.Failed, fullError)
		return fullError
	}

	// Step 2: Apply Ingress configuration
	logging.Infof("[Prereq] Step 2: Applying KnativeServing Ingress")
	if err := applyKnativeIngressWithClient(dyn, wildcardFQDN); err != nil {
		fullError := fmt.Errorf("failed to apply Knative Ingress: %v", err)
		logging.Errorf("[Prereq - %s ]  %v", KnativeServingStepName, fullError)
		k.SetStatus(model.Failed, fullError)
		return fullError
	}
	logging.Infof("[Prereq - %s] KnativeServing configuration applied, waiting for readiness", KnativeServingStepName)
	return nil
}

func applyKnativeServing(dynamicClient dynamic.Interface, wildcardFQDN string) error {
	// Use provided wildcard FQDN or default fallback
	if wildcardFQDN == "" {
		return fmt.Errorf("wildcard FQDN is empty")
	}
	wildcardFQDNForServingDefinition := strings.ReplaceAll(wildcardFQDN, "*.", "")
	// Create KnativeServing resource
	knativeServing := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.knative.dev/v1beta1",
			"kind":       "KnativeServing",
			"metadata": map[string]interface{}{
				"name":      "knative-serving",
				"namespace": KnativeServingNamespace,
			},
			"spec": map[string]interface{}{
				"config": map[string]interface{}{
					"config-autoscaler": map[string]interface{}{
						"enable-scale-to-zero": "true",
					},
					"config-features": map[string]interface{}{
						"kubernetes.podspec-affinity":                "enabled",
						"kubernetes.podspec-init-containers":         "enabled",
						"kubernetes.podspec-persistent-volume-claim": "enabled",
						"kubernetes.podspec-persistent-volume-write": "enabled",
						"kubernetes.podspec-schedulername":           "enabled",
						"kubernetes.podspec-securitycontext":         "enabled",
						"kubernetes.podspec-tolerations":             "enabled",
						"kubernetes.podspec-volumes-emptydir":        "enabled",
						"multi-container":                            "enabled",
					},
					"domain": map[string]interface{}{
						wildcardFQDNForServingDefinition: "",
					},
					"network": map[string]interface{}{
						"domainTemplate":          "{{.Name}}{{.Namespace}}.{{.Domain}}",
						"ingress-class":           "kourier.ingress.networking.knative.dev",
						"default-external-scheme": "https",
					},
				},
				"high-availability": map[string]interface{}{
					"replicas": int64(1),
				},
				"ingress": map[string]interface{}{
					"kourier": map[string]interface{}{
						"enabled":      true,
						"service-type": "ClusterIP",
					},
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "operator.knative.dev",
		Version:  "v1beta1",
		Resource: "knativeservings",
	}

	_, err := dynamicClient.Resource(gvr).Namespace(KnativeServingNamespace).Create(context.Background(), knativeServing, metav1.CreateOptions{})
	return err
}

// applyKnativeIngressWithClient allows injecting a fake dynamic client in tests.
func applyKnativeIngressWithClient(dynamicClient dynamic.Interface, wildcardFQDN string) error {
	// Use provided wildcard FQDN or default fallback
	if wildcardFQDN == "" {
		return fmt.Errorf("wildcard FQDN is empty")
	}

	// Create Ingress resource
	ingress := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "Ingress",
			"metadata": map[string]interface{}{
				"name":      "knative-serving",
				"namespace": KnativeServingNamespace,
			},
			"spec": map[string]interface{}{
				"ingressClassName": "nginx",
				"rules": []interface{}{
					map[string]interface{}{
						"host": wildcardFQDN,
						"http": map[string]interface{}{
							"paths": []interface{}{
								map[string]interface{}{
									"path":     "/",
									"pathType": "Prefix",
									"backend": map[string]interface{}{
										"service": map[string]interface{}{
											"name": "kourier",
											"port": map[string]interface{}{
												"number": int64(80),
											},
										},
									},
								},
							},
						},
					},
				},
				"tls": []interface{}{
					map[string]interface{}{
						"hosts": []interface{}{
							wildcardFQDN,
						},
						"secretName": RunaiInferenceSecretName,
					},
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	_, err := dynamicClient.Resource(gvr).Namespace(KnativeServingNamespace).Create(context.Background(), ingress, metav1.CreateOptions{})
	return err
}

func isKnativeServingFullyUp(ctx context.Context, c client.Client) (bool, error) {
	for _, deployName := range requiredServingDeployments {
		var deploy appsv1.Deployment
		if err := c.Get(ctx, types.NamespacedName{
			Namespace: KnativeServingNamespace,
			Name:      deployName,
		}, &deploy); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("failed to get deployment %s: %w", deployName, err)
		}
		ready := IsDeploymentAvailable(&deploy)
		if !ready {
			return false, nil
		}
		if err := checkPods(ctx, c, KnativeServingNamespace, &deploy); err != nil {
			return false, nil
		}
	}
	return true, nil
}

func IsDeploymentAvailable(d *appsv1.Deployment) bool {
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable &&
			cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func checkPods(ctx context.Context, c client.Client, namespace string, deploy *appsv1.Deployment) error {
	podList := &corev1.PodList{}

	selector, err := labels.ValidatedSelectorFromSet(deploy.Spec.Selector.MatchLabels)
	if err != nil {
		return err
	}

	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: selector},
	}

	if err := c.List(ctx, podList, listOpts...); err != nil {
		return fmt.Errorf("failed to list pods for deployment %s: %w", deploy.Name, err)
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found for deployment %s/%s", namespace, deploy.Name)
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("pod %s is in phase %s, expected Running", pod.Name, pod.Status.Phase)
		}

		isReady := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}
		if !isReady {
			return fmt.Errorf("pod %s is not reporting Ready: True (Ready Probe failed or waiting)", pod.Name)
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 0 {
				return fmt.Errorf("pod %s (container %s) is unstable with %d restarts (k9s red)",
					pod.Name, containerStatus.Name, containerStatus.RestartCount)
			}
		}
	}
	return nil
}
