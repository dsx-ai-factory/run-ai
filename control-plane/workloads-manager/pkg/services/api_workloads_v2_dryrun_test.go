package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/clients/workloads"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/settings"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
)

const (
	urlValidate = "/api/v2/workloads?dryRun=true"
)

// buildManifest creates a K8s manifest map with the given name and namespace
func buildManifest(name, namespace string) map[string]interface{} {
	manifest := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": name,
		},
	}

	if namespace != "" {
		manifest["metadata"].(map[string]interface{})["namespace"] = namespace
	}

	return manifest
}

var _ = Describe("WorkloadV2 dryRun", func() {
	var (
		router        chi.Router
		randoms       *testutils.RandomValues
		csclient      *clusterservice.SimpleMock
		workloadsMock *workloads.SimpleMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		k8sUtil := NewK8sResourceUtil()
		impl = &WorkloadsV2Impl{
			k8sUtil: k8sUtil,
		}

		authMock.SetNoPermissions().AddPermittedTenants(randoms)

		org_unit.SetMock(org_unit.NewSimpleMock().
			WithRandoms(randoms))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, clusterInfoOf(randoms, testutils.DefaultClusterVersionForTesting)))

		checker.InitClusterCapabilitiesChecker("", defaultClusterCacheConfig(), "")

		setting_pkg.Init(nil, nil)
		setting_pkg.Register[*settings.YamlSubmission]()

		// Setup workload types mock
		readyPhase := "Ready"
		deploymentWorkloadType := workloadsApi.WorkloadTypeConfig{
			Name:         "Deployment",
			Group:        "apps",
			CategoryName: convutils.Ref("default-category"),
			PriorityName: convutils.Ref("default-priority"),
			ClusterStatuses: []workloadsApi.WorkloadTypeStatus{
				{
					ClusterId: convutils.Ref(randoms.ClusterId),
					Version:   convutils.Ref("v1"),
					Phase:     &readyPhase,
				},
			},
		}

		workloadsMock = workloads.NewSimpleMock().
			WithPrioritiesAndCategories(randoms.ClusterId, defaultPrioritiesAndCategories).
			WithWorkloadType(deploymentWorkloadType)
		workloads.SetMock(workloadsMock)

		csclient = clusterservice.NewSimpleMock()
		csclient.WithDefaultNodePool(randoms.ClusterId)
		clusters.SetMock(csclient)

		dao.RestartUpdateCountForTest()

		router = createRouterForTest()
	})

	AfterEach(func() {
		cleanTables()
	})

	Context("effective name calculation", func() {
		It("should return API metadata name when provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "api-name",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Name).To(Equal("api-name"))
			Expect(response.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
		})

		It("should return manifest name when API metadata name is empty", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Name).To(Equal("manifest-name"))
		})

		It("should generate name with suffix when UseGivenNameAsPrefix is true", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:                 "prefix",
					UseGivenNameAsPrefix: true,
					ProjectId:            randoms.ProjectIdStr,
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Name).To(HavePrefix("prefix-"))
			Expect(len(response.Metadata.Name)).To(BeNumerically(">", len("prefix-")))
		})
	})

	Context("effective priority calculation", func() {
		It("should return API metadata priority when provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Priority).NotTo(BeNil())
			Expect(*response.Metadata.Priority).To(Equal("high"))
		})

		It("should fail validation for non-available priority with proper error field", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("invalid-priority"),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("priority 'invalid-priority' is not available"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldPriority))
		})

		It("should apply default priority from workload type when not provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					// Priority not set - should come from workload type default
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Priority).NotTo(BeNil())
			Expect(*response.Metadata.Priority).To(Equal("default-priority"))
		})
	})

	Context("effective category calculation", func() {
		It("should return API metadata category when provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Category:  convutils.Ref("train"),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Category).NotTo(BeNil())
			Expect(*response.Metadata.Category).To(Equal("train"))
		})

		It("should apply default category from workload type when not provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					// Category not set - should come from workload type default
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Category).NotTo(BeNil())
			Expect(*response.Metadata.Category).To(Equal("default-category"))
		})

		It("should fail validation for non-available category with proper error field", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Category:  convutils.Ref("invalid-category"),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("category 'invalid-category' is not available"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldCategory))
		})
	})

	Context("effective preemptibility calculation", func() {
		It("should return API metadata preemptibility when provided", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:           "test-workload",
					ProjectId:      randoms.ProjectIdStr,
					Priority:       convutils.Ref("low"),
					Preemptibility: convutils.Ref(api.PREEMPTIBILITY_NON_PREEMPTIBLE),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Preemptibility).NotTo(BeNil())
			Expect(*response.Metadata.Preemptibility).To(Equal(api.PREEMPTIBILITY_NON_PREEMPTIBLE))
		})

		It("should calculate preemptibility from high priority as non-preemptible", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("high"),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Preemptibility).NotTo(BeNil())
			Expect(*response.Metadata.Preemptibility).To(Equal(api.PREEMPTIBILITY_NON_PREEMPTIBLE))
		})

		It("should calculate preemptibility from low priority as preemptible", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("low"),
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.Preemptibility).NotTo(BeNil())
			Expect(*response.Metadata.Preemptibility).To(Equal(api.PREEMPTIBILITY_PREEMPTIBLE))
		})
	})

	Context("project resolution", func() {
		It("should fail when both projectId and clusterId are empty with proper error field", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name: "test-workload",
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("either projectId or clusterId must be provided"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldProject))
		})

		It("should fail when both projectId and clusterId are provided with proper error field", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("projectId and clusterId are mutually exclusive"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldProject))
		})

		It("should resolve project from clusterId and namespace", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ClusterId: randoms.ClusterId,
				},
				Manifest: buildManifest("manifest-name", fmt.Sprintf("runai-%s", randoms.ProjectIdStr)),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
		})

		It("should keep the metadata project and ignore namespace", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: buildManifest("manifest-name", "runai-ns"),
			}

			response := validateWorkloadForTest(router, randoms, &request)
			Expect(response.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
		})
	})

	Context("name validation", func() {
		It("should fail when name is empty with proper error field", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: buildManifest("", "runai-team-a"),
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("Workload name must not be empty"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldName))
		})

		It("should fail when name has invalid characters with proper error field", func() {
			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "Invalid_Name",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: buildManifest("manifest-name", ""),
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("must consist of lower case alphanumeric"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldName))
		})
	})

	Context("GVK validation", func() {
		It("should fail for unknown workload type with proper error field", func() {
			unknownManifest := map[string]interface{}{
				"apiVersion": "unknown.group/v1",
				"kind":       "UnknownKind",
				"metadata": map[string]interface{}{
					"name": "test-workload",
				},
			}

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "test-workload",
					ProjectId: randoms.ProjectIdStr,
				},
				Manifest: unknownManifest,
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			Expect(submissionErrs).To(HaveLen(1))
			Expect(submissionErrs[0].Message).To(ContainSubstring("workload type not found for GVK"))
			Expect(submissionErrs[0].ComplianceIssues).To(HaveLen(1))
			Expect(submissionErrs[0].ComplianceIssues[0].Field).To(Equal(ErrorFieldGVK))
		})

		Context("combined metadata and manifest", func() {
			It("should correctly use all metadata fields", func() {
				request := api.WorkloadV2CreateRequest{
					Metadata: api.WorkloadV2MetadataCreateParams{
						Name:           "api-workload",
						ProjectId:      randoms.ProjectIdStr,
						Priority:       convutils.Ref("high"),
						Category:       convutils.Ref("train"),
						Preemptibility: convutils.Ref(api.PREEMPTIBILITY_PREEMPTIBLE),
					},
					Manifest: buildManifest("manifest-name", ""),
				}

				response := validateWorkloadForTest(router, randoms, &request)
				Expect(response.Metadata.Name).To(Equal("api-workload"))
				Expect(response.Metadata.ProjectId).To(Equal(randoms.ProjectIdStr))
				Expect(*response.Metadata.Priority).To(Equal("high"))
				Expect(*response.Metadata.Category).To(Equal("train"))
				Expect(*response.Metadata.Preemptibility).To(Equal(api.PREEMPTIBILITY_PREEMPTIBLE))
			})
		})
	})

	Context("multiple validation errors", func() {
		It("should return multiple validation errors at once", func() {
			// Create a request with multiple validation issues from different validation steps:
			// 1. Unknown GVK (from getWorkloadTypeFromGVK)
			// 2. Invalid category (from validateCategoryAndPriority)
			// 3. Invalid name (from ValidateWorkloadName)
			unknownManifest := map[string]interface{}{
				"apiVersion": "unknown.group/v1",
				"kind":       "UnknownKind",
				"metadata": map[string]interface{}{
					"name": "test-workload",
				},
			}

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "INVALID-NAME", // Invalid: contains uppercase
					ProjectId: randoms.ProjectIdStr,
					Category:  convutils.Ref("invalid-category"),
				},
				Manifest: unknownManifest,
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			// Should have multiple errors - GVK, category, and name from different validation steps
			Expect(len(submissionErrs)).To(BeNumerically(">=", 3))

			// Collect all error fields
			errorFields := make(map[string]bool)
			for _, err := range submissionErrs {
				for _, issue := range err.ComplianceIssues {
					errorFields[issue.Field] = true
				}
			}

			// Verify we got errors from different validation steps
			Expect(errorFields).To(HaveKey(ErrorFieldGVK))
			Expect(errorFields).To(HaveKey(ErrorFieldCategory))
			Expect(errorFields).To(HaveKey(ErrorFieldName))
		})

		It("should return GVK and priority errors together", func() {
			// Different multi-error scenario: GVK error + priority error
			unknownManifest := map[string]interface{}{
				"apiVersion": "unknown.group/v1",
				"kind":       "UnknownKind",
				"metadata": map[string]interface{}{
					"name": "valid-name",
				},
			}

			request := api.WorkloadV2CreateRequest{
				Metadata: api.WorkloadV2MetadataCreateParams{
					Name:      "valid-name",
					ProjectId: randoms.ProjectIdStr,
					Priority:  convutils.Ref("invalid-priority"),
				},
				Manifest: unknownManifest,
			}

			resp := tryToValidateWorkloadForTest(router, randoms, &request)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))

			submissionErrs := parseSubmissionErrors(resp)
			// Should have at least 2 errors (GVK and priority)
			Expect(len(submissionErrs)).To(BeNumerically(">=", 2))

			// Collect all error fields
			errorFields := make(map[string]bool)
			for _, err := range submissionErrs {
				for _, issue := range err.ComplianceIssues {
					errorFields[issue.Field] = true
				}
			}

			// Verify we got both GVK and priority errors
			Expect(errorFields).To(HaveKey(ErrorFieldGVK))
			Expect(errorFields).To(HaveKey(ErrorFieldPriority))
		})
	})
})

func tryToValidateWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, request *api.WorkloadV2CreateRequest) *httptest.ResponseRecorder {
	validateReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testHost, urlValidate), makePayload(request))
	Expect(err).To(BeNil())
	validateReq.Header = *headerForAdmin(randoms)

	validateResp := httptest.NewRecorder()
	router.ServeHTTP(validateResp, validateReq)

	return validateResp
}

func validateWorkloadForTest(router chi.Router, randoms *testutils.RandomValues, request *api.WorkloadV2CreateRequest) *api.WorkloadV2 {
	validateResp := tryToValidateWorkloadForTest(router, randoms, request)
	Expect(validateResp.Code).To(Equal(http.StatusAccepted))

	var result api.WorkloadV2
	jsonErr := json.Unmarshal(validateResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}

func parseSubmissionErrors(resp *httptest.ResponseRecorder) []api.SubmissionErrorV2 {
	var result []api.SubmissionErrorV2
	jsonErr := json.Unmarshal(resp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())
	return result
}
