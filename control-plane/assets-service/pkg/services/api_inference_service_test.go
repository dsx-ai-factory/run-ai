/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
)

func randomInferenceCreationRequestV2(randoms *testutils.RandomValues, envAssetId string) *api.InferenceCreationRequestV2 {
	return &api.InferenceCreationRequestV2{
		Meta: api.WorkloadCreationMeta{
			Name:      testutils.RandString(5),
			ProjectId: strconv.Itoa(int(randoms.ProjectId)),
			ClusterId: randoms.ClusterId,
		},
		Assets: api.AssetsIds{
			Environment: envAssetId,
		},
	}
}

var _ = Describe("Test Inference Service", func() {
	var randoms *testutils.RandomValues
	var dept1, dept2 *testutils.RandomValues
	var policyMockUpdateTime time.Time

	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		policyMockUpdateTime = time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		randoms = testutils.NewRandomValues()
		dept1 = randoms.CloneWithDifferentDepartment()
		dept2 = randoms.CloneWithDifferentDepartment().Forbidden()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, dept1, dept2)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, dept1, dept2))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		displayedCluster := testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, displayedCluster))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	Context("valid create request - basic", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Meta.Id).NotTo(Equal(""))
			Expect(inference.Spec.SpecificEnv).To(BeNil())
			Expect(inference.Job.JobId).To(Equal(""))
			Expect(inference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(inference.Workload, "metadata", "name", "value")
			Expect(field).To(Equal(reqPayload.Name))
			field, _ = getNestedField(inference.Workload, "metadata", "namespace", "value")
			Expect(field).To(Equal(reqPayload.Namespace))
			field, _ = getNestedField(inference.Workload, "spec", "image", "value")
			Expect(field).To(Equal(*envAsset.Spec.Image))
			field, _ = getNestedField(inference.Workload, "spec", "ports", "items", "serving-port", "value", "serviceType", "value")
			Expect(field).To(Equal("ServingPort"))
			field, _ = getNestedField(inference.Workload, "spec", "ports", "items", "serving-port", "value", "container", "value")
			Expect(field).To(Equal(float64(8080)))
		})
	})

	Context("valid create request - with autoscaling", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas:                 convutils.Ref(int32(0)),
					MaxReplicas:                 convutils.Ref(int32(5)),
					ThresholdMetric:             convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:              convutils.Ref(int32(80)),
					ScaleToZeroRetentionSeconds: convutils.Ref(int32(60)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Meta.Id).NotTo(Equal(""))
			Expect(inference.Spec.SpecificEnv).ToNot(BeNil())
			Expect(inference.Job.JobId).To(Equal(""))
			Expect(inference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(inference.Workload, "spec", "minScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.MinReplicas)))
			field, _ = getNestedField(inference.Workload, "spec", "maxScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.MaxReplicas)))
		})
	})

	Context("valid create request - with autoscaling custom metric", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas:     convutils.Ref(int32(1)),
					MaxReplicas:     convutils.Ref(int32(1)),
					ThresholdMetric: convutils.Ref("my_custom_metric"),
					ThresholdValue:  convutils.Ref(int32(80)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Meta.Id).NotTo(Equal(""))
			Expect(inference.Spec.SpecificEnv).ToNot(BeNil())
			Expect(inference.Job.JobId).To(Equal(""))
			Expect(inference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(inference.Workload, "spec", "metricName", "value")
			Expect(field).To(Equal(*reqPayload.SpecificEnv.AutoScaling.ThresholdMetric))
			field, _ = getNestedField(inference.Workload, "spec", "metric", "value")
			Expect(field).To(BeNil())
		})
	})

	Context("valid create request - with autoscaling advanced settings", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas:                  convutils.Ref(int32(0)),
					MaxReplicas:                  convutils.Ref(int32(5)),
					InitialReplicas:              convutils.Ref(int32(2)),
					ActivationReplicas:           convutils.Ref(int32(3)),
					ThresholdMetric:              convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:               convutils.Ref(int32(80)),
					MetricThresholdPercentage:    convutils.Ref(float32(55.5)),
					ConcurrencyHardLimit:         convutils.Ref(int32(1)),
					ScaleToZeroRetentionSeconds:  convutils.Ref(int32(60)),
					ScaleDownDelaySeconds:        convutils.Ref(int32(120)),
					InitializationTimeoutSeconds: convutils.Ref(int32(240)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Meta.Id).NotTo(Equal(""))
			Expect(inference.Spec.SpecificEnv).ToNot(BeNil())
			Expect(inference.Job.JobId).To(Equal(""))
			Expect(inference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(inference.Workload, "spec", "minScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.MinReplicas)))
			field, _ = getNestedField(inference.Workload, "spec", "maxScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.MaxReplicas)))
			field, _ = getNestedField(inference.Workload, "spec", "initialScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.InitialReplicas)))
			field, _ = getNestedField(inference.Workload, "spec", "activationScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.ActivationReplicas)))
			field, _ = getNestedField(inference.Workload, "spec", "targetUtilizationPercentage", "value")
			Expect(field).To(Equal("55.5"))
			field, _ = getNestedField(inference.Workload, "spec", "concurrencyHardLimit", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.ConcurrencyHardLimit)))
			field, _ = getNestedField(inference.Workload, "spec", "scaleDownDelaySeconds", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.ScaleDownDelaySeconds)))
			field, _ = getNestedField(inference.Workload, "spec", "initializationTimeoutSeconds", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.InitializationTimeoutSeconds)))
		})
	})

	Context("valid create request - with serving port authorization (type, users, local)", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			authType, err := api.NewServingPortAccessAuthorizationTypeEnumFromValue("authorizedUsersOrGroups")
			Expect(err).To(BeNil())

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				ServingPortAccess: &api.ServingPortAccess{
					AuthorizationType:      &authType,
					AuthorizedUsers:        &[]string{"user-a", "user-b"},
					ClusterLocalAccessOnly: convutils.Ref(true),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())

			var field interface{}
			field, _ = getNestedField(inference.Workload, "spec", "servingPortAuthorizationType", "value")
			Expect(field).To(Equal(string(*reqPayload.SpecificEnv.ServingPortAccess.AuthorizationType)))
			field, _ = getNestedField(inference.Workload, "spec", "servingPortAuthorizedUsers", "value")
			Expect(field).To(Equal(strings.Join(*reqPayload.SpecificEnv.ServingPortAccess.AuthorizedUsers, ",")))
			field, _ = getNestedField(inference.Workload, "spec", "isPrivateServiceUrl", "value")
			Expect(field).To(Equal(*reqPayload.SpecificEnv.ServingPortAccess.ClusterLocalAccessOnly))
		})
	})

	Context("valid create request - with serving port authorization (groups)", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			authType, err := api.NewServingPortAccessAuthorizationTypeEnumFromValue("authorizedUsersOrGroups")
			Expect(err).To(BeNil())

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				ServingPortAccess: &api.ServingPortAccess{
					AuthorizationType: &authType,
					AuthorizedGroups:  &[]string{"group-a", "group-b"},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())

			var field interface{}
			field, _ = getNestedField(inference.Workload, "spec", "servingPortAuthorizationType", "value")
			Expect(field).To(Equal(string(*reqPayload.SpecificEnv.ServingPortAccess.AuthorizationType)))
			field, _ = getNestedField(inference.Workload, "spec", "servingPortAuthorizedGroups", "value")
			Expect(field).To(Equal(strings.Join(*reqPayload.SpecificEnv.ServingPortAccess.AuthorizedGroups, ",")))
		})
	})

	Context("request with no servingPort", func() {
		It("should fail", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("specific env autoscaling", func() {
		It("must have max replicas, if min replicas is provided", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: maxReplicas is mandatory when minReplicas is provided"))
		})

		It("must have min replicas, if max replicas is provided", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: minReplicas is mandatory when maxReplicas is provided"))
		})

		It("must have min replicas which is non-negative", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(-1)),
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("MinReplicas: provided parameter is not respecting minimum value constraint"))
		})

		It("must have max replicas which is greater or equal to min replicas", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(3)),
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: maxReplicas (2) must be no less than minReplicas (3)"))
		})

		It("must have threshold value if threshold metric is provided", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric: convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdValue is mandatory when thresholdMetric is provided"))
		})

		It("must have threshold metric if threshold value is provided", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdValue: convutils.Ref(int32(80)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdMetric is mandatory when thresholdValue is provided"))
		})

		It("must have positive threshold value (given 0)", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric: convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:  convutils.Ref(int32(0)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdValue must be greater than 0"))
		})

		It("must have positive threshold value (given < 0)", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric: convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:  convutils.Ref(int32(-5)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdValue must be greater than 0"))
		})

		It("must have positive ScaleToZeroRetentionSeconds value (given < 0)", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric:             convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:              convutils.Ref(int32(80)),
					ScaleToZeroRetentionSeconds: convutils.Ref(int32(-5)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("ScaleToZeroRetentionSeconds: provided parameter is not respecting minimum value"))
		})
	})

	Context("invalid serving port authorization", func() {
		It("should fail", func() {
			// note that more detailed validation tests are implemented as part
			// of workload-composer
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			authType, err := api.NewServingPortAccessAuthorizationTypeEnumFromValue("authorizedUsersOrGroups")
			Expect(err).To(BeNil())

			// invalid - should be either users or groups, not both
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				ServingPortAccess: &api.ServingPortAccess{
					AuthorizationType: &authType,
					AuthorizedUsers:   &[]string{"user-a", "user-b"},
					AuthorizedGroups:  &[]string{"group-a", "group-b"},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("valid create request - v2 - basic", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomInferenceCreationRequestV2(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.InferenceV2
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Meta.Name).To(Equal(reqPayload.Meta.Name))
			Expect(inference.Assets).ToNot(BeNil())
			Expect(inference.Assets.Environment.Id).To(Equal(envAsset.Meta.Id))
			Expect(inference.Assets.Compute).To(BeNil())
			Expect(inference.Assets.Datasources).To(BeNil())
			Expect(inference.SpecificRunParams).To(BeNil())
		})
	})

	Context("valid create request - v2 - with autoscaling", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomInferenceCreationRequestV2(randoms, envAsset.Meta.Id)

			reqPayload.SpecificRunParams = &api.InferenceSpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(2)),
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.InferenceV2
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Meta.Name).To(Equal(reqPayload.Meta.Name))
			Expect(inference.Assets).ToNot(BeNil())
			Expect(inference.Assets.Environment.Id).To(Equal(envAsset.Meta.Id))
			Expect(inference.Assets.Compute).To(BeNil())
			Expect(inference.Assets.Datasources).To(BeNil())
			Expect(inference.SpecificRunParams).ToNot(BeNil())
			Expect(inference.SpecificRunParams.AutoScaling).ToNot(BeNil())
			Expect(*inference.SpecificRunParams.AutoScaling.MinReplicas).To(Equal(int32(2)))
		})
	})

	Context("invalid autoscaling - v2", func() {
		It("should fail", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomInferenceCreationRequestV2(randoms, envAsset.Meta.Id)

			reqPayload.SpecificRunParams = &api.InferenceSpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(5)), // bad param - min < max
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("valid create request - v2 - with serving port auth (type, users, local)", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomInferenceCreationRequestV2(randoms, envAsset.Meta.Id)

			authType, err := api.NewServingPortAccessAuthorizationTypeEnumFromValue("authorizedUsersOrGroups")
			Expect(err).To(BeNil())

			reqPayload.SpecificRunParams = &api.InferenceSpecificRunParams{
				ServingPortAccess: &api.ServingPortAccess{
					AuthorizationType:      &authType,
					AuthorizedUsers:        &[]string{"user-a", "user-b"},
					ClusterLocalAccessOnly: convutils.Ref(true),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.InferenceV2
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.SpecificRunParams).ToNot(BeNil())
			Expect(inference.SpecificRunParams.ServingPortAccess).ToNot(BeNil())
			Expect(inference.SpecificRunParams.ServingPortAccess.AuthorizationType).ToNot(BeNil())
			Expect(*inference.SpecificRunParams.ServingPortAccess.AuthorizationType).To(Equal(*reqPayload.SpecificRunParams.ServingPortAccess.AuthorizationType))
			Expect(inference.SpecificRunParams.ServingPortAccess.AuthorizedUsers).ToNot(BeNil())
			Expect(*inference.SpecificRunParams.ServingPortAccess.AuthorizedUsers).To(Equal(*reqPayload.SpecificRunParams.ServingPortAccess.AuthorizedUsers))
			Expect(inference.SpecificRunParams.ServingPortAccess.ClusterLocalAccessOnly).ToNot(BeNil())
			Expect(*inference.SpecificRunParams.ServingPortAccess.ClusterLocalAccessOnly).To(Equal(*reqPayload.SpecificRunParams.ServingPortAccess.ClusterLocalAccessOnly))
		})
	})

	Context("valid create request - v2 - with serving port auth (groups)", func() {
		It("should succeed", func() {
			randomEnvAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomEnvAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}

			envAsset, _, _ := createEnvironmentAsset(randomEnvAsset, router, headerForAdmin(randoms))

			reqPayload := randomInferenceCreationRequestV2(randoms, envAsset.Meta.Id)

			authType, err := api.NewServingPortAccessAuthorizationTypeEnumFromValue("authorizedUsersOrGroups")
			Expect(err).To(BeNil())

			reqPayload.SpecificRunParams = &api.InferenceSpecificRunParams{
				ServingPortAccess: &api.ServingPortAccess{
					AuthorizationType: &authType,
					AuthorizedGroups:  &[]string{"group-a", "group-b"},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.InferenceV2
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.SpecificRunParams).ToNot(BeNil())
			Expect(inference.SpecificRunParams.ServingPortAccess).ToNot(BeNil())
			Expect(inference.SpecificRunParams.ServingPortAccess.AuthorizationType).ToNot(BeNil())
			Expect(*inference.SpecificRunParams.ServingPortAccess.AuthorizationType).To(Equal(*reqPayload.SpecificRunParams.ServingPortAccess.AuthorizationType))
			Expect(inference.SpecificRunParams.ServingPortAccess.AuthorizedGroups).ToNot(BeNil())
			Expect(*inference.SpecificRunParams.ServingPortAccess.AuthorizedGroups).To(Equal(*reqPayload.SpecificRunParams.ServingPortAccess.AuthorizedGroups))
		})
	})
})
