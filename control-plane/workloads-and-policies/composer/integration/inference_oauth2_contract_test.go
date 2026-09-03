/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integration_test

import (
	"context"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/inference_workloads"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("authenticated inference Knative contract", func() {
	DescribeTable("moves serving through oauth2-proxy without restoring the legacy authenticated Service ref",
		func(serviceSpecLabelsSupported bool, expectedNetworkingEndpoints int) {
			spec := authenticatedNIMInferenceSpec()
			composer := inference_workloads.NewInferenceToKnativeComposer(
				"my-llm",
				"team-a",
				"runai-team-a",
				"tester",
				nil,
				composer.FieldsComposerOptions{
					WorkloadV2ServiceSpecLabelsSupported: serviceSpecLabelsSupported,
				},
			).WithClusterStorageConfig(inferenceOAuth2ContractConfig())
			target := inference_workloads.NewMinimalKnativeServiceTarget()

			Expect(composer.Compose(context.Background(), convutils.Ref(spec), target)()).To(BeNil())

			podSpec := target.Service.Spec.Template.Spec.PodSpec
			Expect(podSpec.Containers).To(HaveLen(2))
			userContainer := podSpec.Containers[0]
			proxyContainer := podSpec.Containers[1]

			Expect(userContainer.Ports).To(BeEmpty())
			Expect(userContainer.ReadinessProbe).To(BeNil())
			Expect(proxyContainer.Name).To(Equal("runai-oauth2-proxy-0"))
			Expect(proxyContainer.Ports).To(Equal([]corev1.ContainerPort{{ContainerPort: 4180}}))
			Expect(proxyContainer.Args).To(ContainElement("--upstream=http://localhost:8080"))
			Expect(proxyContainer.ReadinessProbe).NotTo(BeNil())
			Expect(proxyContainer.ReadinessProbe.HTTPGet).NotTo(BeNil())
			Expect(proxyContainer.ReadinessProbe.HTTPGet.Path).To(Equal("/healthz"))
			Expect(proxyContainer.ReadinessProbe.HTTPGet.Port.IntVal).To(Equal(int32(8080)))

			Expect(target.Config.Networking).To(HaveLen(expectedNetworkingEndpoints))
			for _, endpoint := range target.Config.Networking {
				for _, service := range endpoint.Services {
					Expect(service.HasRef()).To(BeFalse(),
						"authenticated Knative serving must not emit the old NetworkingService ref")
					Expect(service.HasAuthorization()).To(BeFalse(),
						"oauth2 authorization belongs to the sidecar, not the metrics Service")
				}
			}

			if serviceSpecLabelsSupported {
				Expect(target.Config.Networking).To(HaveLen(1))
				Expect(target.Config.Networking[0].Services).To(HaveLen(1))
				metricsService := target.Config.Networking[0].Services[0]
				Expect(metricsService.HasSpec()).To(BeTrue())
				metricsSpec := metricsService.Spec.Get()
				Expect(metricsSpec).NotTo(BeNil())
				Expect(metricsSpec.GetName()).To(Equal("my-llm-metrics"))
				Expect(metricsSpec.GetLabels()).To(HaveKeyWithValue(
					"run.ai/workload-metrics-service", "true"))
			} else {
				Expect(target.Config.Networking).To(BeEmpty(),
					"clusters without Service spec labels cannot receive the metrics endpoint")
			}
		},
		Entry("metrics is the sole endpoint when Service spec labels are supported", true, 1),
		Entry("no endpoint is emitted when Service spec labels are unsupported", false, 0),
	)
})

func authenticatedNIMInferenceSpec() workloadsApi.InferenceSpecSpec {
	GinkgoHelper()
	const input = `
image: nvcr.io/nim/meta/llama3:latest
servingPort:
  container: 8080
  authorizationType: authenticatedUsers
probes:
  readiness:
    initialDelaySeconds: 5
    periodSeconds: 10
    handler:
      httpGet:
        path: /healthz
        port: 8080
`
	var spec workloadsApi.InferenceSpecSpec
	Expect(yaml.Unmarshal([]byte(input), &spec)).To(Succeed())
	return spec
}

func inferenceOAuth2ContractConfig() composer.ClusterStorageConfig {
	reported := clustersApi.ClusterReportedConfig{}
	reported.SetRunaiOauth2ProxyImage("test.registry/oauth2-proxy:latest")
	reported.SetRunaiOauth2ProxyWorkloadAuthSecretName("runai-auth-secret")
	reported.SetRunaiOauth2ProxySessionStoreType("redis")
	reported.SetRunaiOauth2ProxyRedisFQDN("redis.runai-backend.svc.cluster.local")
	reported.SetRunaiOauth2ProxyOidcIssuerUrl("https://test.run.ai/auth/realms/runai")
	reported.SetSubdomainEnabled(true)

	status := clustersApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster := clustersApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	cluster.SetDomain("https://test.run.ai")
	return composer.NewClusterStorageConfigFromCluster(&cluster)
}
