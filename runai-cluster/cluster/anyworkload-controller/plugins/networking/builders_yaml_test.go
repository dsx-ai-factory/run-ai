/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"fmt"
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/testutil"
	testutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

type InputConfig struct {
	Runner RunnerInput                      `json:"runner"`
	Config *dromosv1alpha1.NetworkingConfig `json:"networking"`
}

type RunnerInput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

var _ = Describe("Builders YAML Tests", func() {
	var (
		plugin     *Plugin
		testRunner *testutils.YAMLTestRunner
	)

	BeforeEach(func() {
		plugin = NewPlugin(nil)
		testRunner = testutils.NewYAMLTestRunner("testing/inputs", "testing/expected")
	})

	Context("Builder Success Cases", func() {
		testCases := []testutils.YAMLTestConfig{
			{TestName: "simple_ingress"},
			{TestName: "use_existing_service"},
			{TestName: "openshift_route"},
			{TestName: "standalone_ingress"},
			{TestName: "internal_service"},
			{TestName: "non_root_path_auth"},
			{TestName: "haproxy_path_auth"},
		}

		for _, tc := range testCases {
			tc := tc
			It(fmt.Sprintf("should build resources for %s", tc.TestName), func() {
				runBuilderTest(plugin, testRunner, tc.TestName)
			})
		}
	})
})

func extractResourcesFromClient(ctx context.Context, k8sClient client.Client, namespace string) ([]client.Object, error) {
	var resources []client.Object

	svcList := &corev1.ServiceList{}
	if err := k8sClient.List(ctx, svcList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	for i := range svcList.Items {
		svcList.Items[i].SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
		resources = append(resources, &svcList.Items[i])
	}

	deployList := &appsv1.DeploymentList{}
	if err := k8sClient.List(ctx, deployList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	for i := range deployList.Items {
		deployList.Items[i].SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
		resources = append(resources, &deployList.Items[i])
	}

	ingressList := &networkingv1.IngressList{}
	if err := k8sClient.List(ctx, ingressList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	for i := range ingressList.Items {
		ingressList.Items[i].SetGroupVersionKind(networkingv1.SchemeGroupVersion.WithKind("Ingress"))
		resources = append(resources, &ingressList.Items[i])
	}

	routeList := &openshiftv1.RouteList{}
	if err := k8sClient.List(ctx, routeList, client.InNamespace(namespace)); err == nil {
		for i := range routeList.Items {
			routeList.Items[i].SetGroupVersionKind(openshiftv1.GroupVersion.WithKind("Route"))
			resources = append(resources, &routeList.Items[i])
		}
	}

	return resources, nil
}

func sortResources(resources []client.Object) {
	sort.Slice(resources, func(i, j int) bool {
		typeOrder := map[string]int{
			"Service":    1,
			"Deployment": 2,
			"Ingress":    3,
			"Route":      4,
		}

		iType := resources[i].GetObjectKind().GroupVersionKind().Kind
		jType := resources[j].GetObjectKind().GroupVersionKind().Kind

		if typeOrder[iType] != typeOrder[jType] {
			return typeOrder[iType] < typeOrder[jType]
		}

		return resources[i].GetName() < resources[j].GetName()
	})
}

func runBuilderTest(plugin *Plugin, testRunner *testutils.YAMLTestRunner, testName string) {
	ctx := context.Background()

	inputYaml, err := testRunner.LoadInputFile(testName)
	Expect(err).NotTo(HaveOccurred())

	var input InputConfig
	err = yaml.Unmarshal(inputYaml, &input)
	Expect(err).NotTo(HaveOccurred())

	runner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Runner.Name,
			Namespace: input.Runner.Namespace,
			UID:       "test-uid-12345",
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Instructions: &dromosv1alpha1.Instructions{
				Networking: input.Config,
			},
		},
	}

	scheme := runtime.NewScheme()
	Expect(dromosv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = openshiftv1.AddToScheme(scheme)

	k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(scheme).Build())

	err = plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
	Expect(err).NotTo(HaveOccurred())

	resources, err := extractResourcesFromClient(ctx, k8sClient, runner.Namespace)
	Expect(err).NotTo(HaveOccurred())

	sortResources(resources)

	actualYaml, err := testutils.ObjectsToYaml(resources)
	Expect(err).NotTo(HaveOccurred())

	expectedYaml, created, err := testRunner.LoadExpectedFile(testName, actualYaml)
	Expect(err).NotTo(HaveOccurred())

	if created {
		Skip(fmt.Sprintf("Expected file created. Review: %s", testName))
		return
	}

	Expect(actualYaml.String()).To(Equal(string(expectedYaml)))
}
