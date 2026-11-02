/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"net/http"

	clustersapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/dinf_to_lws"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/inferenceendpoint"
	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

const inferenceRoutingTestCluster = "inference-routing-cluster"

// inferenceIntegrationClusterForTest returns a cluster that declares the
// InferenceEndpoint integration with ready operands and an administrator Gateway.
func inferenceIntegrationClusterForTest() *clustersapi.DisplayedCluster {
	return &clustersapi.DisplayedCluster{
		Status: *clustersapi.NewNullableClusterDisplayedStatus(&clustersapi.ClusterDisplayedStatus{
			Config: *clustersapi.NewNullableClusterReportedConfig(&clustersapi.ClusterReportedConfig{
				InferenceEndpointSchemaVersion: *clustersapi.NewNullableString(ptr.To("v1")),
				GatewayAPIEnabled:              *clustersapi.NewNullableBool(ptr.To(true)),
				GatewayName:                    *clustersapi.NewNullableString(ptr.To("inference")),
				GatewayNamespace:               *clustersapi.NewNullableString(ptr.To("gateways")),
			}),
			Operands: &map[string]clustersapi.ClusterReportedStatusOperandsValue{
				"runai-agent": {Ready: true}, "anyworkload-controller": {Ready: true}, "inference-endpoint-controller": {Ready: true},
			},
		}),
	}
}

// composeDistributedInferenceForTest runs the native distributed inference composer
// and returns its LWS as a manifest and its networking as the persisted configuration.
func composeDistributedInferenceForTest(
	authorizationType *workloadsapi.DistributedInferenceServingPortAccessAuthorizationTypeEnum,
) (map[string]interface{}, *models.WorkloadConfiguration) {
	config := &composer.StaticClusterStorageConfig{}
	config.SetIngressEnabled(false)
	config.SetGatewayAPIEnabled(true)
	config.SetGatewayName("inference")
	config.SetGatewayNamespace("gateways")
	config.SetClusterDomain("example.com")
	spec := workloadsapi.NewDistributedInferenceSpecSpec()
	leader := workloadsapi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("example.com/model:1")
	spec.SetLeader(*leader)
	port := workloadsapi.NewDistributedInferenceServingPort()
	port.SetPort(8000)
	port.SetExposeExternally(true)
	port.SetExposedUrl("https://model.example.com")
	if authorizationType != nil {
		port.SetAuthorizationType(*authorizationType)
		config.SetOauth2ProxyImage("example.com/oauth2-proxy:1")
		config.SetOauth2ProxyOidcIssuerUrl("https://auth.example.com")
	}
	spec.SetServingPort(*port)
	builder := dinf_to_lws.NewDistributedInferenceToLWSComposer("model", "project", "runai-project", "user", nil, nil, composer.FieldsComposerOptions{}, nil).
		WithLeaderWorkerSetName("model-1234").WithClusterStorageConfig(config)
	target := dinf_to_lws.NewMinimalLeaderWorkerSetTarget()
	Expect(builder.Compose(context.Background(), spec, target)()).To(BeNil())

	rawManifest, err := json.Marshal(target.LeaderWorkerSet)
	Expect(err).NotTo(HaveOccurred())
	var manifest map[string]interface{}
	Expect(json.Unmarshal(rawManifest, &manifest)).To(Succeed())

	rawNetworking, err := json.Marshal(target.Config.Networking)
	Expect(err).NotTo(HaveOccurred())
	var networking models.NetworkingConfiguration
	Expect(json.Unmarshal(rawNetworking, &networking)).To(Succeed())
	return manifest, &models.WorkloadConfiguration{Networking: &networking}
}

var _ = Describe("inference routing", func() {
	dinfLabels := map[string]string{
		ConvertedFromNativeLabel:    TrueAsString,
		constants.LabelWorkloadKind: composer.DistributedInference,
	}
	defaults := &inferenceendpoint.Defaults{GatewayName: "inference", GatewayNamespace: "gateways", EPPProfile: "gie-v1.5"}

	DescribeTable("opts in only through the spec fields",
		func(routing *api.WorkloadV2ConfigurationInferenceRouting, expected bool) {
			Expect(inferenceRoutingOptedIn(routing)).To(Equal(expected))
		},
		Entry("no routing fields", nil, false),
		Entry("loadAwareRouting false", &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(false)}, false),
		Entry("authorization type only", &api.WorkloadV2ConfigurationInferenceRouting{
			ServingPortAuthorizationType: ptr.To(api.SERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHENTICATED_USERS),
		}, false),
		Entry("empty modelRewrites", &api.WorkloadV2ConfigurationInferenceRouting{ModelRewrites: &[]api.ModelRewrite{}}, false),
		Entry("loadAwareRouting true", &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}, true),
		Entry("objectivePriority 0", &api.WorkloadV2ConfigurationInferenceRouting{ObjectivePriority: ptr.To(int32(0))}, true),
		Entry("modelRewrites", &api.WorkloadV2ConfigurationInferenceRouting{ModelRewrites: &[]api.ModelRewrite{
			{Model: "chat", Targets: []api.ModelRewriteTarget{{Model: "chat-v1", Weight: 1}}},
		}}, true),
	)

	It("accepts inferenceRouting only from the trusted internal caller", func() {
		configuration := &api.WorkloadV2Configuration{InferenceRouting: &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}}
		Expect(validateInferenceRoutingOrigin(configuration, true)).To(BeNil())

		err := validateInferenceRoutingOrigin(configuration, false)
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
		Expect(err.Message).To(Equal(inferenceRoutingNotSubmittableMessage))

		Expect(validateInferenceRoutingOrigin(&api.WorkloadV2Configuration{}, false)).To(BeNil())
		Expect(validateInferenceRoutingOrigin(nil, false)).To(BeNil())
	})

	Context("resolving cluster defaults", func() {
		var clusterMock *clusterservice.SimpleMock
		optedIn := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}

		BeforeEach(func() {
			clusterMock = clusterservice.NewSimpleMock()
			clusters.SetMock(clusterMock)
		})

		It("does not read the cluster when the workload did not opt in", func() {
			// The mock has no cluster, so any lookup would fail.
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(false)},
				inferenceRoutingTestCluster, true, dinfLabels)
			Expect(err).To(BeNil())
			Expect(resolved).To(BeNil())
		})

		It("resolves the administrator defaults on a cluster with the integration", func() {
			clusterMock.WithCluster(inferenceRoutingTestCluster, inferenceIntegrationClusterForTest())
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, true, dinfLabels)
			Expect(err).To(BeNil())
			Expect(resolved).To(Equal(defaults))
		})

		It("rejects an opt-in on a cluster that does not declare the integration", func() {
			clusterMock.WithCluster(inferenceRoutingTestCluster, &clustersapi.DisplayedCluster{})
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, true, dinfLabels)
			Expect(resolved).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(Equal(loadAwareRoutingUnavailableMessage))
		})

		It("rejects an opt-in with the integration error", func() {
			cluster := inferenceIntegrationClusterForTest()
			(*cluster.Status.Get().Operands)["inference-endpoint-controller"] = clustersapi.ClusterReportedStatusOperandsValue{Ready: false}
			clusterMock.WithCluster(inferenceRoutingTestCluster, cluster)
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, true, dinfLabels)
			Expect(resolved).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(ContainSubstring("requires ready operand inference-endpoint-controller"))
		})

		DescribeTable("rejects an opt-in on workloads that are not trusted distributed inference",
			func(trusted bool, labels map[string]string) {
				clusterMock.WithCluster(inferenceRoutingTestCluster, inferenceIntegrationClusterForTest())
				resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, trusted, labels)
				Expect(resolved).To(BeNil())
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("supported only for distributed inference workloads"))
			},
			Entry("untrusted origin", false, dinfLabels),
			Entry("another workload kind", true, map[string]string{ConvertedFromNativeLabel: TrueAsString, constants.LabelWorkloadKind: "Inference"}),
		)
	})

	Context("applying the intent", func() {
		It("leaves the configuration unchanged when the workload did not opt in", func() {
			manifest, configuration := composeDistributedInferenceForTest(nil)
			before, err := json.Marshal(configuration)
			Expect(err).NotTo(HaveOccurred())

			Expect(applyInferenceRouting(nil, nil, manifest, configuration, "model")).To(BeNil())

			after, err := json.Marshal(configuration)
			Expect(err).NotTo(HaveOccurred())
			Expect(after).To(MatchJSON(before))
			Expect((*configuration.Networking)[0].ExternalAccess.Inference).To(BeNil())
		})

		DescribeTable("sets the inference block on the serving endpoint",
			func(routing *api.WorkloadV2ConfigurationInferenceRouting, expected string) {
				manifest, configuration := composeDistributedInferenceForTest(nil)
				Expect(applyInferenceRouting(routing, defaults, manifest, configuration, "model")).To(BeNil())

				access := (*configuration.Networking)[0].ExternalAccess
				Expect(access.HttpRoute).NotTo(BeNil())
				raw, err := json.Marshal(access.Inference)
				Expect(err).NotTo(HaveOccurred())
				Expect(raw).To(MatchJSON(expected))
			},
			Entry("loadAwareRouting", &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)},
				`{"epp":{"profile":"gie-v1.5"}}`),
			Entry("objectivePriority 0 names the lane after the workload", &api.WorkloadV2ConfigurationInferenceRouting{ObjectivePriority: ptr.To(int32(0))},
				`{"epp":{"profile":"gie-v1.5"},"objectives":[{"model":"model","priority":0}]}`),
			Entry("modelRewrites without objectivePriority", &api.WorkloadV2ConfigurationInferenceRouting{ModelRewrites: &[]api.ModelRewrite{
				{Model: "chat", Targets: []api.ModelRewriteTarget{{Model: "chat-v1", Weight: 90}, {Model: "chat-v2", Weight: 10}}},
			}}, `{"epp":{"profile":"gie-v1.5"},"rewrites":[{"model":"chat","targets":[{"model":"chat-v1","weight":90},{"model":"chat-v2","weight":10}]}]}`),
		)

		It("targets the auth proxy port when the serving port is protected", func() {
			authenticated := workloadsapi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHENTICATED_USERS
			manifest, configuration := composeDistributedInferenceForTest(&authenticated)
			Expect(*(*(*configuration.Networking)[0].Services)[0].Spec.Ports[0].TargetPort).To(Equal(composer.Oauth2ContainerDefaultPort))

			protected := &api.WorkloadV2ConfigurationInferenceRouting{
				LoadAwareRouting:             ptr.To(true),
				ServingPortAuthorizationType: ptr.To(api.SERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHENTICATED_USERS),
			}
			Expect(applyInferenceRouting(protected, defaults, manifest, configuration, "model")).To(BeNil())
			Expect((*configuration.Networking)[0].ExternalAccess.Inference).NotTo(BeNil())

			// A public authorization type passes protected=false, which must not route to the proxy port.
			manifest, configuration = composeDistributedInferenceForTest(&authenticated)
			public := &api.WorkloadV2ConfigurationInferenceRouting{
				LoadAwareRouting:             ptr.To(true),
				ServingPortAuthorizationType: ptr.To(api.SERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC),
			}
			err := applyInferenceRouting(public, defaults, manifest, configuration, "model")
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("public distributed inference must not target the auth proxy port"))
		})

		It("rejects a protected type when the composed Service targets the model port", func() {
			manifest, configuration := composeDistributedInferenceForTest(nil)
			routing := &api.WorkloadV2ConfigurationInferenceRouting{
				LoadAwareRouting:             ptr.To(true),
				ServingPortAuthorizationType: ptr.To(api.SERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHORIZED_USERS),
			}
			err := applyInferenceRouting(routing, defaults, manifest, configuration, "model")
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("protected distributed inference must target the auth proxy port"))
		})

		It("rejects a workload without one serving endpoint", func() {
			manifest, _ := composeDistributedInferenceForTest(nil)
			routing := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}
			err := applyInferenceRouting(routing, defaults, manifest, &models.WorkloadConfiguration{}, "model")
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(ContainSubstring("requires one serving port endpoint"))
		})
	})
})
