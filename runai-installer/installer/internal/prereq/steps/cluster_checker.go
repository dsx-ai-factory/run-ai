/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"regexp"

	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateComponentNamespace(ctx context.Context, kbClient client.Client, component, componentBaseName, expectedNamespace string) error {
	componentNameRegex := regexp.MustCompile(fmt.Sprintf("%s.*%s", component, componentBaseName))
	componentDisplayName := fmt.Sprintf("%s %s", component, componentBaseName)
	logging.Infof("[K8s] Validating component '%s' in namespace '%s'...", componentDisplayName, expectedNamespace)
	_, err := getDeploymentNamespace(ctx, kbClient, componentNameRegex, componentDisplayName, expectedNamespace)
	if err == nil {
		logging.Infof("[K8s] Found deployment for '%s'", componentDisplayName)
		return nil
	}
	_, err = getDeamonsetsNamespace(ctx, kbClient, componentNameRegex, componentDisplayName, expectedNamespace)
	if err == nil {
		logging.Infof("[K8s] Found daemonset for '%s'", componentDisplayName)
		return nil
	}
	logging.Infof("[K8s] Component '%s' not found in cluster", componentDisplayName)
	return fmt.Errorf("component %s not found in the cluster", componentDisplayName)
}

func getDeploymentNamespace(ctx context.Context, k8sReader client.Reader, componentNameRegex *regexp.Regexp, componentDisplayName, expectedNamespace string) (string, error) {
	namespacesSet := map[string]struct{}{}

	deployments := []appsv1.Deployment{}
	limit := 100
	continueToken := ""

	for {
		deploymentsList := &appsv1.DeploymentList{}
		err := k8sReader.List(ctx, deploymentsList,
			client.Limit(limit),
			client.Continue(continueToken))
		if err != nil {
			return "", err
		}

		deployments = append(deployments, deploymentsList.Items...)

		// Break if there are no more pages
		if deploymentsList.Continue == "" {
			break
		}
		continueToken = deploymentsList.Continue
	}

	for _, deployment := range deployments {
		if checValidResource(deployment.ObjectMeta, componentNameRegex, expectedNamespace) {
			namespacesSet[deployment.Namespace] = struct{}{}
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

func getDeamonsetsNamespace(ctx context.Context, k8sReader client.Reader, componentNameRegex *regexp.Regexp, componentDisplayName, expectedNamespace string) (string, error) {
	namespacesSet := map[string]struct{}{}

	daemonSets := []appsv1.DaemonSet{}
	limit := 100
	continueToken := ""

	for {
		daemonSetsList := &appsv1.DaemonSetList{}
		err := k8sReader.List(ctx, daemonSetsList,
			client.Limit(limit),
			client.Continue(continueToken))
		if err != nil {
			return "", err
		}

		daemonSets = append(daemonSets, daemonSetsList.Items...)

		// Break if there are no more pages
		if daemonSetsList.Continue == "" {
			break
		}
		continueToken = daemonSetsList.Continue
	}

	for _, daemonSet := range daemonSets {
		if checValidResource(daemonSet.ObjectMeta, componentNameRegex, expectedNamespace) {
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

func checValidResource(resource metav1.ObjectMeta, componentNameRegex *regexp.Regexp, expectedNamespace string) bool {
	if expectedNamespace != "" && expectedNamespace != resource.Namespace {
		return false
	}
	if componentNameRegex.MatchString(resource.Name) {
		return true
	}
	return false
}

func IsCrdsInstalledInstalled(cfg *rest.Config, crdName string) bool {
	logging.Infof("[K8s] Checking if CRD '%s' exists...", crdName)
	crdClient, err := apiextclient.NewForConfig(cfg)
	if err != nil {
		logging.Errorf("[K8s] Failed to create CRD client: %v", err)
		return false
	}

	_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), crdName, metav1.GetOptions{})
	if err != nil {
		logging.Infof("[K8s] CRD '%s' not found", crdName)
		return false
	}
	logging.Infof("[K8s] CRD '%s' exists", crdName)
	return true
}

func IsNamespacePresent(k8sClient client.Client, namespace string) (bool, error) {
	logging.Infof("[K8s] Checking if namespace '%s' exists...", namespace)
	ns := &corev1.Namespace{}

	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[K8s] Namespace '%s' does not exist", namespace)
			return false, nil
		}
		logging.Errorf("[K8s] Error checking namespace '%s': %v", namespace, err)
		return false, fmt.Errorf("error checking namespace %q: %v", namespace, err)
	}

	logging.Infof("[K8s] Namespace '%s' exists", namespace)
	return true, nil
}

func HasCustomResources(cl client.Client, gvk schema.GroupVersionKind) (bool, error) {
	logging.Infof("[K8s] Checking for custom resources of type %s/%s/%s...", gvk.Group, gvk.Version, gvk.Kind)
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(gvk)
	err := cl.List(context.Background(), list)
	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[K8s] No custom resources of type %s found", gvk.Kind)
			return false, nil
		}
		logging.Errorf("[K8s] Error listing custom resources: %v", err)
		return false, err
	}
	logging.Infof("[K8s] Found %d custom resources of type %s", len(list.Items), gvk.Kind)
	return len(list.Items) > 0, nil
}
