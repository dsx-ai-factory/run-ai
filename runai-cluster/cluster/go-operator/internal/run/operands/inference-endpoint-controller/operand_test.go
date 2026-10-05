/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_endpoint_controller

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/inference-endpoint-controller/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	discoveryfake "k8s.io/client-go/discovery/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type discoveryClient struct {
	kc.HybridClient
	discovery discovery.DiscoveryInterface
}

func (c discoveryClient) NewDiscoveryClient() (discovery.DiscoveryInterface, error) {
	return c.discovery, nil
}

func TestInferenceEndpoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inference endpoint operand")
}

var _ = Describe("Inference endpoint operand", func() {
	var config *runv1.RunaiConfig
	var reader discoveryClient
	BeforeEach(func() {
		config = &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "runai"}}
		config.SetDefaultsWhereNeeded()
		scheme := runtime.NewScheme()
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(monitorv1.AddToScheme(scheme)).To(Succeed())
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		reader = discoveryClient{HybridClient: kc.New(client, client, nil), discovery: &discoveryfake.FakeDiscovery{Fake: &clienttesting.Fake{}}}
	})

	It("reports missing API prerequisites without returning a discovery error", func() {
		status, err := Operand(reader, config).Status(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeFalse())
		Expect(status.Reasons).To(ContainElement(ContainSubstring("InferenceEndpoint")))
	})

	It("renders an HA controller deployment using configured image and probes", func() {
		settings := config.Spec.InferenceEndpointController
		settings.Image.Registry = ptr.To("mirror.example")
		settings.Image.Tag = ptr.To("test")
		settings.Replicas = ptr.To[int32](3)
		object, err := resources.NewDeploymentResource(Name, reader, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		deployment := object.(*appsv1.Deployment)
		Expect(*deployment.Spec.Replicas).To(Equal(int32(3)))
		Expect(deployment.Namespace).To(Equal("runai"))
		Expect(deployment.Spec.Template.Spec.ServiceAccountName).To(Equal(Name))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Image).To(Equal("mirror.example/inference-endpoint-controller:test"))
		Expect(container.Args).To(ContainElement("--leader-elect"))
		Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/readyz"))
		Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/healthz"))
		Expect(container.Resources).To(Equal(corev1.ResourceRequirements(*settings.Resources)))
	})

	DescribeTable("hardens the controller without assigning an OpenShift UID", func(openShift bool) {
		if openShift {
			config.Spec.Internal.Global.Platform.KubernetesFlavor = runv1internal.KubernetesFlavorOpenShift
		}
		object, err := resources.NewDeploymentResource(Name, reader, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		pod := object.(*appsv1.Deployment).Spec.Template.Spec
		Expect(pod.AutomountServiceAccountToken).To(Equal(ptr.To(true)))
		security := pod.Containers[0].SecurityContext
		Expect(security).NotTo(BeNil())
		Expect(security.RunAsNonRoot).To(Equal(ptr.To(true)))
		Expect(security.ReadOnlyRootFilesystem).To(Equal(ptr.To(true)))
		Expect(security.AllowPrivilegeEscalation).To(Equal(ptr.To(false)))
		Expect(security.Capabilities.Drop).To(ContainElement(corev1.Capability("ALL")))
		Expect(security.SeccompProfile.Type).To(Equal(corev1.SeccompProfileTypeRuntimeDefault))
		if openShift {
			Expect(security.RunAsUser).To(BeNil())
		}
	}, Entry("Kubernetes", false), Entry("OpenShift", true))

	It("preserves explicit global security settings without modifying them", func() {
		global := config.Spec.Global.SecurityContext
		global.RunAsUser = ptr.To[int64](12345)
		global.ReadOnlyRootFilesystem = ptr.To(false)
		global.Capabilities.Add = []corev1.Capability{"NET_BIND_SERVICE"}
		before := *global
		object, err := resources.NewDeploymentResource(Name, reader, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		security := object.(*appsv1.Deployment).Spec.Template.Spec.Containers[0].SecurityContext
		Expect(security.RunAsUser).To(Equal(ptr.To[int64](12345)))
		Expect(security.ReadOnlyRootFilesystem).To(Equal(ptr.To(false)))
		Expect(security.Capabilities.Add).To(Equal([]corev1.Capability{"NET_BIND_SERVICE"}))
		Expect(*global).To(Equal(before))
	})

	It("connects the metrics monitor to the controller without sending a service account token", func() {
		deploymentObject, err := resources.NewDeploymentResource(Name, reader, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		serviceObject, err := resources.NewServiceResource(Name, reader, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		monitorObject, err := resources.NewServiceMonitorResource(Name, reader, config).Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		deployment := deploymentObject.(*appsv1.Deployment)
		service := serviceObject.(*corev1.Service)
		monitor := monitorObject.(*monitorv1.ServiceMonitor)
		Expect(service.Spec.Selector).To(Equal(deployment.Spec.Selector.MatchLabels))
		Expect(monitor.Spec.Selector.MatchLabels).To(Equal(service.Labels))
		Expect(monitor.Spec.Endpoints).To(HaveLen(1))
		Expect(monitor.Spec.Endpoints[0].Port).To(Equal(service.Spec.Ports[0].Name))
		endpointJSON, err := json.Marshal(monitor.Spec.Endpoints[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(endpointJSON)).NotTo(ContainSubstring("bearerTokenFile"))
		Expect(monitor.Spec.Endpoints[0].Authorization).To(BeNil())
		Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(ContainElement("--metrics-secure=false"))
	})
})
