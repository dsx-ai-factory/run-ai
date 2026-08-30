package services

import (
	"encoding/json"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"

	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("Test Model Inference Service", func() {
	var randoms *testutils.RandomValues
	var dept1, dept2 *testutils.RandomValues
	var policyMockUpdateTime time.Time

	var envAsset api.EnvironmentAsset
	var compAsset api.ComputeAsset
	var modelAsset api.ModelAsset

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

		randomTenantEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
		randomTenantEnvironmentAsset.Spec.Connections = []api.Connection{
			{
				Name: "serving-port",
				InternalToolInfo: &api.InternalToolInfo{
					ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
					ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
					ContainerPort:  8080,
				},
			},
		}
		randomTenantEnvironmentAsset.Spec.EnvironmentVariables = &[]api.EnvironmentVariableOfAsset{
			{Name: convutils.Ref(MODEL_ASSET_ENV_VAR_RUNAI_MODEL), Value: convutils.Ref("model")},
			{Name: convutils.Ref(MODEL_ASSET_ENV_VAR_RUNAI_MODEL_NAME), Value: convutils.Ref("model_name")},
		}

		envAsset, _, _ = createEnvironmentAsset(randomTenantEnvironmentAsset, router, headerForAdmin(randoms))
		compAsset, _, _ = createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
		modelAsset, _, _ = createModelAsset(buildRandomTenantModelAsset(envAsset, compAsset), router, headerForAdmin(randoms))
	})

	Context("trying to create model inference with no permission", func() {
		It("should return forbidden error", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)
			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerWithMissingInfo(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("Model contains asset from different project", func() {
		It("should fail with asset not found in project error", func() {
			modelAsset, _, _ := createModelAsset(buildRandomProjectModelAsset(dept1, envAsset, compAsset), router, headerForAdmin(randoms))

			reqPayload := randomModelInferenceCreationRequest(dept2, modelAsset.Meta.Id, compAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForNonAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
			Expect(createHttpResp.Body).To(ContainSubstring("not permitted to perform this operation"))
		})
	})

	Context("incorrect asset type", func() {
		It("should fail validation", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, envAsset.Meta.Id, compAsset.Meta.Id) // envAsset instead of modelAsset

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("requested resource does not exist"))
		})
	})

	Context("specific env autoscaling", func() {
		It("must have max replicas, if min replicas is provided", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: maxReplicas is mandatory when minReplicas is provided"))
		})

		It("must have min replicas, if max replicas is provided", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: minReplicas is mandatory when maxReplicas is provided"))
		})

		It("must have min replicas which is non-negative", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(-1)),
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("MinReplicas: provided parameter is not respecting minimum value constraint"))
		})

		It("must have max replicas which is greater or equal to min replicas", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(3)),
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: maxReplicas (2) must be no less than minReplicas (3)"))
		})

		It("must have threshold value if threshold metric is provided", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric: convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdValue is mandatory when thresholdMetric is provided"))
		})

		It("must have threshold metric if threshold value is provided", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdValue: convutils.Ref(int32(80)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdMetric is mandatory when thresholdValue is provided"))
		})

		It("must have positive threshold value (given 0)", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric: convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:  convutils.Ref(int32(0)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdValue must be greater than 0"))
		})

		It("must have positive threshold value (given < 0)", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					ThresholdMetric: convutils.Ref(composer.AutoScalingMetric_CONCURRENCY),
					ThresholdValue:  convutils.Ref(int32(-5)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: thresholdValue must be greater than 0"))
		})

		It("must have MinReplicas 0 when using ScaleToZeroRetentionSeconds", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas:                 convutils.Ref(int32(1)),
					MaxReplicas:                 convutils.Ref(int32(5)),
					ThresholdMetric:             convutils.Ref(composer.AutoScalingMetric_THROUGHPUT),
					ThresholdValue:              convutils.Ref(int32(80)),
					ScaleToZeroRetentionSeconds: convutils.Ref(int32(300)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoScaling: scaleToZeroRetentionSeconds is only supported when minReplicas is 0"))
		})
	})

	Context("valid create request", func() {
		It("should succeed", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var modelInference api.ModelInference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &modelInference)
			Expect(jsonErr).To(BeNil())
			Expect(modelInference.Meta.Id).NotTo(Equal(""))
			Expect(modelInference.Spec.SpecificEnv).NotTo(BeNil())
			Expect(modelInference.Job.JobId).To(Equal(""))
			Expect(modelInference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(modelInference.Workload, "metadata", "name", "value")
			Expect(field).To(Equal(reqPayload.Name))
			field, _ = getNestedField(modelInference.Workload, "metadata", "namespace", "value")
			Expect(field).To(Equal(reqPayload.Namespace))
			field, _ = getNestedField(modelInference.Workload, "spec", "image", "value")
			Expect(field).To(Equal(*envAsset.Spec.Image))
			field, _ = getNestedField(modelInference.Workload, "spec", "memoryLimit", "value")
			Expect(field).To(Equal(*compAsset.Spec.CpuMemoryLimit))
		})

		It("should succeed with old env assets", func() {
			randomTenantEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomTenantEnvironmentAsset.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
					},
				},
			}
			randomTenantEnvironmentAsset.Spec.EnvironmentVariables = nil

			envAsset, _, _ = createEnvironmentAsset(randomTenantEnvironmentAsset, router, headerForAdmin(randoms))
			compAsset, _, _ = createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			modelAsset, _, _ = createModelAsset(buildRandomTenantModelAsset(envAsset, compAsset), router, headerForAdmin(randoms))

			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var modelInference api.ModelInference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &modelInference)
			Expect(jsonErr).To(BeNil())
			Expect(modelInference.Meta.Id).NotTo(Equal(""))
			Expect(modelInference.Spec.SpecificEnv).To(BeNil())
			Expect(modelInference.Job.JobId).To(Equal(""))
			Expect(modelInference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(modelInference.Workload, "metadata", "name", "value")
			Expect(field).To(Equal(reqPayload.Name))
			field, _ = getNestedField(modelInference.Workload, "metadata", "namespace", "value")
			Expect(field).To(Equal(reqPayload.Namespace))
			field, _ = getNestedField(modelInference.Workload, "spec", "image", "value")
			Expect(field).To(Equal(*envAsset.Spec.Image))
			field, _ = getNestedField(modelInference.Workload, "spec", "memoryLimit", "value")
			Expect(field).To(Equal(*compAsset.Spec.CpuMemoryLimit))
		})

		It("should succeed with compute assets as request param", func() {
			customUserCompAsset, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			// create another compute asset
			// and set it as the compute asset for the model inference
			reqPayload.Assets.Compute = customUserCompAsset.Meta.Id

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var modelInference api.ModelInference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &modelInference)
			Expect(jsonErr).To(BeNil())

			var field interface{}
			field, _ = getNestedField(modelInference.Workload, "spec", "memoryLimit", "value")
			Expect(field).To(Equal(*customUserCompAsset.Spec.CpuMemoryLimit))
		})

		It("should succeed when using autoscaling", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas:                 convutils.Ref(int32(0)),
					MaxReplicas:                 convutils.Ref(int32(5)),
					ThresholdMetric:             convutils.Ref(composer.AutoScalingMetric_THROUGHPUT),
					ThresholdValue:              convutils.Ref(int32(80)),
					ScaleToZeroRetentionSeconds: convutils.Ref(int32(300)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var modelInference api.ModelInference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &modelInference)
			Expect(jsonErr).To(BeNil())
			Expect(modelInference.Meta.Id).NotTo(Equal(""))
			Expect(modelInference.Spec.SpecificEnv).NotTo(BeNil())
			Expect(modelInference.Job.JobId).To(Equal(""))
			Expect(modelInference.Workload).NotTo(BeEmpty())

			var field interface{}
			field, _ = getNestedField(modelInference.Workload, "spec", "minScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.MinReplicas)))
			field, _ = getNestedField(modelInference.Workload, "spec", "scaleToZeroRetentionSeconds", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.ScaleToZeroRetentionSeconds)))
			field, _ = getNestedField(modelInference.Workload, "spec", "maxScale", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.MaxReplicas)))
			field, _ = getNestedField(modelInference.Workload, "spec", "metric", "value")
			Expect(field).To(Equal(string(*reqPayload.SpecificEnv.AutoScaling.ThresholdMetric)))
			field, _ = getNestedField(modelInference.Workload, "spec", "target", "value")
			Expect(field).To(Equal(float64(*reqPayload.SpecificEnv.AutoScaling.ThresholdValue)))
		})

		It("verify model name is set in environment variable", func() {
			reqPayload := randomModelInferenceCreationRequest(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlModelInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var modelInference api.ModelInference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &modelInference)
			Expect(jsonErr).To(BeNil())
			Expect(modelInference.Meta.Id).NotTo(Equal(""))
			Expect(modelInference.Workload).NotTo(BeNil())

			modelField, _ := getNestedField(modelInference.Workload, "spec", "environment", "items", MODEL_ASSET_ENV_VAR_RUNAI_MODEL, "value")
			Expect(modelField).To(Equal(*modelAsset.Spec.HuggingfaceId))

			ModelNameField, _ := getNestedField(modelInference.Workload, "spec", "environment", "items", MODEL_ASSET_ENV_VAR_RUNAI_MODEL_NAME, "value")
			Expect(ModelNameField).To(BeNil())
		})
	})

	Context("valid create request - v2", func() {
		It("should succeed", func() {
			reqPayload := randomModelInferenceCreationRequestV2(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificRunParams = &api.InferenceSpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(2)),
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var modelInference api.ModelInferenceV2
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &modelInference)
			Expect(jsonErr).To(BeNil())
			Expect(modelInference.Meta.Name).To(Equal(reqPayload.Meta.Name))
			Expect(modelInference.Assets).ToNot(BeNil())
			Expect(modelInference.Assets.Environment.Id).To(Equal(envAsset.Meta.Id))
			Expect(modelInference.Assets.Compute).ToNot(BeNil())
			Expect(modelInference.Assets.Compute.Id).To(Equal(compAsset.Meta.Id))
			Expect(modelInference.Assets.Datasources).To(BeNil())
			Expect(modelInference.SpecificRunParams).ToNot(BeNil())
			Expect(modelInference.SpecificRunParams.EnvironmentVariables).ToNot(BeNil())
			Expect(len(*modelInference.SpecificRunParams.EnvironmentVariables)).To(Equal(1))
			Expect(*(*modelInference.SpecificRunParams.EnvironmentVariables)[0].Name).To(Equal(MODEL_ASSET_ENV_VAR_RUNAI_MODEL))
			Expect(*(*modelInference.SpecificRunParams.EnvironmentVariables)[0].Value).To(Equal(*modelAsset.Spec.HuggingfaceId))
			Expect(modelInference.SpecificRunParams.AutoScaling).ToNot(BeNil())
			Expect(*modelInference.SpecificRunParams.AutoScaling.MinReplicas).To(Equal(int32(2)))
		})
	})

	Context("invalid autoscaling - v2", func() {
		It("should succeed", func() {
			reqPayload := randomModelInferenceCreationRequestV2(randoms, modelAsset.Meta.Id, compAsset.Meta.Id)

			reqPayload.SpecificRunParams = &api.InferenceSpecificRunParams{
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(5)), // bad param - min>max
					MaxReplicas: convutils.Ref(int32(2)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlModelInferenceV2, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})
})

func randomModelInferenceCreationRequest(randoms *testutils.RandomValues, modelAssetId string, ComputeAssetId string) *api.ModelInferenceCreationRequest {
	return &api.ModelInferenceCreationRequest{
		Name:      testutils.RandString(5),
		ProjectId: int32(randoms.ProjectId),
		ClusterId: randoms.ClusterId,
		Namespace: testutils.RandString(8),
		ModelId:   modelAssetId,
		Assets: api.ModelInferenceCreationRequestAssets{
			Compute: ComputeAssetId,
		},
	}
}

func randomModelInferenceCreationRequestV2(randoms *testutils.RandomValues, modelAssetId string, computeAssetId string) *api.ModelInferenceCreationRequestV2 {
	return &api.ModelInferenceCreationRequestV2{
		Meta: api.WorkloadCreationMeta{
			Name:      testutils.RandString(5),
			ProjectId: strconv.Itoa(int(randoms.ProjectId)),
			ClusterId: randoms.ClusterId,
		},
		Assets: api.ModelInferenceCreationRequestV2Assets{
			Compute: computeAssetId,
			Model:   modelAssetId,
		},
	}
}

func getNestedField(data map[string]interface{}, keys ...string) (interface{}, bool) {
	current := data
	for _, key := range keys {
		val, ok := current[key]
		if !ok {
			return nil, false
		}

		if nested, isMap := val.(map[string]interface{}); isMap {
			current = nested
		} else {
			return val, true
		}
	}

	return current, true
}
