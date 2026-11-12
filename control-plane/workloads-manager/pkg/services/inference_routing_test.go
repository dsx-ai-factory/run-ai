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
				"agent": {Ready: true}, "anyworkload-controller": {Ready: true}, "inference-endpoint-controller": {Ready: true},
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
		Entry("loadAwareRouting true", &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}, true),
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
				inferenceRoutingTestCluster, inferenceRoutingDistributedInference)
			Expect(err).To(BeNil())
			Expect(resolved).To(BeNil())
		})

		It("resolves the administrator defaults on a cluster with the integration", func() {
			clusterMock.WithCluster(inferenceRoutingTestCluster, inferenceIntegrationClusterForTest())
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, inferenceRoutingDistributedInference)
			Expect(err).To(BeNil())
			Expect(resolved).To(Equal(defaults))
		})

		It("rejects an opt-in on a cluster that does not declare the integration", func() {
			clusterMock.WithCluster(inferenceRoutingTestCluster, &clustersapi.DisplayedCluster{})
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, inferenceRoutingDistributedInference)
			Expect(resolved).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(Equal(loadAwareRoutingUnavailableMessage))
		})

		It("rejects an opt-in with the integration error", func() {
			cluster := inferenceIntegrationClusterForTest()
			(*cluster.Status.Get().Operands)["inference-endpoint-controller"] = clustersapi.ClusterReportedStatusOperandsValue{Ready: false}
			clusterMock.WithCluster(inferenceRoutingTestCluster, cluster)
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, inferenceRoutingDistributedInference)
			Expect(resolved).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(ContainSubstring("requires ready operand inference-endpoint-controller"))
		})

		It("rejects an opt-in on a workload without a supported adapter", func() {
			clusterMock.WithCluster(inferenceRoutingTestCluster, inferenceIntegrationClusterForTest())
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, inferenceRoutingUnsupported)
			Expect(resolved).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(Equal(inferenceRoutingUnsupportedWorkloadMessage))
		})

		It("resolves the administrator defaults for a NIM service", func() {
			clusterMock.WithCluster(inferenceRoutingTestCluster, inferenceIntegrationClusterForTest())
			resolved, err := resolveInferenceRoutingDefaults(context.Background(), optedIn, inferenceRoutingTestCluster, inferenceRoutingNIMService)
			Expect(err).To(BeNil())
			Expect(resolved).To(Equal(defaults))
		})
	})

	DescribeTable("selects the adapter from the trusted origin",
		func(trustedConvertedFromNative, trustedNIMService bool, resource *ResourceInfo, expected inferenceRoutingAdapter) {
			Expect(inferenceRoutingAdapterFor(trustedConvertedFromNative, trustedNIMService, resource)).To(Equal(expected))
		},
		Entry("trusted distributed inference", true, false, &ResourceInfo{Labels: dinfLabels}, inferenceRoutingDistributedInference),
		Entry("untrusted distributed inference", false, false, &ResourceInfo{Labels: dinfLabels}, inferenceRoutingUnsupported),
		Entry("another converted workload kind", true, false,
			&ResourceInfo{Labels: map[string]string{ConvertedFromNativeLabel: TrueAsString, constants.LabelWorkloadKind: "Inference"}}, inferenceRoutingUnsupported),
		Entry("NIMService from the NIM API", false, true, &ResourceInfo{APIVersion: nimServiceAPIVersion, Kind: nimServiceKind}, inferenceRoutingNIMService),
		Entry("NIMService from the public API", false, false, &ResourceInfo{APIVersion: nimServiceAPIVersion, Kind: nimServiceKind}, inferenceRoutingUnsupported),
		Entry("another kind from the NIM API", false, true, &ResourceInfo{APIVersion: "v1", Kind: "Pod"}, inferenceRoutingUnsupported),
		Entry("distributed inference labels from the NIM API", false, true, &ResourceInfo{Labels: dinfLabels}, inferenceRoutingUnsupported),
	)

	Context("applying the intent", func() {
		It("leaves the configuration unchanged when the workload did not opt in", func() {
			manifest, configuration := composeDistributedInferenceForTest(nil)
			before, err := json.Marshal(configuration)
			Expect(err).NotTo(HaveOccurred())

			Expect(applyInferenceRouting(inferenceRoutingDistributedInference, nil, nil, manifest, configuration, "model")).To(BeNil())

			after, err := json.Marshal(configuration)
			Expect(err).NotTo(HaveOccurred())
			Expect(after).To(MatchJSON(before))
			Expect((*configuration.Networking)[0].ExternalAccess.Inference).To(BeNil())
		})

		DescribeTable("sets the inference block on the serving endpoint",
			func(routing *api.WorkloadV2ConfigurationInferenceRouting, expected string) {
				manifest, configuration := composeDistributedInferenceForTest(nil)
				Expect(applyInferenceRouting(inferenceRoutingDistributedInference, routing, defaults, manifest, configuration, "model")).To(BeNil())

				access := (*configuration.Networking)[0].ExternalAccess
				Expect(access.HttpRoute).NotTo(BeNil())
				raw, err := json.Marshal(access.Inference)
				Expect(err).NotTo(HaveOccurred())
				Expect(raw).To(MatchJSON(expected))
			},
			Entry("loadAwareRouting", &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)},
				`{"epp":{"profile":"gie-v1.5"},"pathPrefix":"/v1"}`),
		)

		It("rejects the composed protected serving port", func() {
			authenticated := workloadsapi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHENTICATED_USERS
			manifest, configuration := composeDistributedInferenceForTest(&authenticated)
			Expect(*(*(*configuration.Networking)[0].Services)[0].Spec.Ports[0].TargetPort).To(Equal(composer.Oauth2ContainerDefaultPort))

			routing := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}
			err := applyInferenceRouting(inferenceRoutingDistributedInference, routing, defaults, manifest, configuration, "model")
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("public distributed inference must not target the auth proxy port"))
		})

		It("rejects a workload without one serving endpoint", func() {
			manifest, _ := composeDistributedInferenceForTest(nil)
			routing := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}
			err := applyInferenceRouting(inferenceRoutingDistributedInference, routing, defaults, manifest, &models.WorkloadConfiguration{}, "model")
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(ContainSubstring("requires one serving port endpoint"))
		})
	})

	Context("applying the intent to a NIM service", func() {
		// The networking that the NIM API composes on a Gateway-only cluster.
		nimGatewayNetworking := func(servingPort *api.NimServiceServingPort, inferenceRouting bool) *[]api.NetworkingEndpoint {
			networking, err := nimServingPortToNetworkingConfig("my-nim", servingPort, nimServiceNetworkingInfo{
				IngressEnabled:   ptr.To(false),
				UseGatewayAPI:    true,
				GatewayName:      "inference",
				GatewayNamespace: "gateways",
			}, inferenceRouting)
			Expect(err).To(BeNil())
			return networking
		}
		publicServingPort := func(url string) *api.NimServiceServingPort {
			return &api.NimServiceServingPort{Port: ptr.To(int32(8000)), ExposeExternally: ptr.To(true), ExposedUrl: ptr.To(url)}
		}

		It("sets the inference block on the composed NIM serving endpoint", func() {
			configuration := &models.WorkloadConfiguration{Networking: cloneNetworkingConfig(nimGatewayNetworking(publicServingPort("https://my-nim.example.com"), true))}
			routing := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}
			Expect(applyInferenceRouting(inferenceRoutingNIMService, routing, defaults, nil, configuration, "my-nim")).To(BeNil())

			raw, err := json.Marshal((*configuration.Networking)[0].ExternalAccess.Inference)
			Expect(err).NotTo(HaveOccurred())
			Expect(raw).To(MatchJSON(`{"epp":{"profile":"gie-v1.5"},"pathPrefix":"/v1"}`))
		})

		It("rejects the composed path-based NIM endpoint", func() {
			configuration := &models.WorkloadConfiguration{Networking: cloneNetworkingConfig(nimGatewayNetworking(publicServingPort("https://example.com/project/my-nim"), true))}
			routing := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}
			err := applyInferenceRouting(inferenceRoutingNIMService, routing, defaults, nil, configuration, "my-nim")
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(err.Message).To(Equal("load-aware routing: NIM service requires a host-based URL; path-based URLs are not supported"))
		})

		It("rejects NIM networking composed without the Service reference", func() {
			configuration := &models.WorkloadConfiguration{Networking: cloneNetworkingConfig(nimGatewayNetworking(publicServingPort("https://my-nim.example.com"), false))}
			routing := &api.WorkloadV2ConfigurationInferenceRouting{LoadAwareRouting: ptr.To(true)}
			err := applyInferenceRouting(inferenceRoutingNIMService, routing, defaults, nil, configuration, "my-nim")
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("requires exactly one referenced serving Service"))
		})
	})
})
