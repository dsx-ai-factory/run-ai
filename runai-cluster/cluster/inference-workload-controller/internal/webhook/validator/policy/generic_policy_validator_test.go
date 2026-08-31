/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"encoding/json"
	"errors"
	inferencepolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	workloadspolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads/policy"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AssetsTest struct {
	AdmissionFakeRequest admission.Request
	PolicyType           string
}

var _ = Describe("Policy Validator CreateMergeStackForFrontend Suite", func() {
	var (
		testEntries            []TableEntry
		admissionFakeRequest   admission.Request
		ctx                    context.Context
		scheme                 *runtime.Scheme
		admissionResponse      admission.Response
		genericPolicyValidator *workloadspolicy.ValidatorHandler
	)

	testEntries = []TableEntry{
		Entry("Inference",
			AssetsTest{
				AdmissionFakeRequest: createAdmissionFakeRequest(constants.KindInferencePolicy, "inference-policy"),
				PolicyType:           constants.KindInferenceWorkload}),
	}

	BeforeEach(func() {
		ctx = context.Background()

		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
	})

	Context("Apply Defaults", func() {
		It("Policy defaults should be accepted within workload object", func() {
			emptyPolicy := inferencepolicy.NewEmptyInferencePolicy()
			emptyPolicyQ, _ := utils.NewJsonQ(emptyPolicy.Spec)

			policySpec := inferencepolicy.InferenceSpec{
				WorkloadFields: field_groups.WorkloadFields{
					ContainerFields: field_groups.ContainerFields{
						Annotations: &field_types.DictionaryField{
							Value: &map[string]field_types.DictionaryItemAndRules{
								"DEFAULTS": {
									Value: ptr.To("default_value"),
								},
							},
						},
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"DEFAULTS": {
									Value: &field_types.PvcsItemValue{
										StorageClass: ptr.To("default_sc"),
										Path:         ptr.To("default_path"),
									},
								},
							},
						},
					},
				},
			}
			policySpecQ, _ := utils.NewJsonQ(policySpec)

			workloadSpec := inferencepolicy.InferenceSpec{
				WorkloadFields: field_groups.WorkloadFields{
					ContainerFields: field_groups.ContainerFields{
						Annotations: &field_types.DictionaryField{
							Value: &map[string]field_types.DictionaryItemAndRules{
								"ann_1": {
									Value: ptr.To("has_value"),
								},
								"ann_2": {
									Value: nil,
								},
							},
						},
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"none": {
									Value: nil,
								},
								"only_sc": {
									Value: &field_types.PvcsItemValue{
										StorageClass: ptr.To("has_sc"),
										ClaimName:    ptr.To("claim1"),
										Size:         ptr.To("10M"),
									},
								},
								"only_path": {
									Value: &field_types.PvcsItemValue{
										Path:      ptr.To("has_path"),
										ClaimName: ptr.To("claim1"),
										Size:      ptr.To("10M"),
									},
								},
								"has_both": {
									Value: &field_types.PvcsItemValue{
										Path:         ptr.To("has_path"),
										StorageClass: ptr.To("has_sc"),
										ClaimName:    ptr.To("claim1"),
										Size:         ptr.To("10M"),
									},
								},
							},
						},
					},
				},
			}
			workloadSpecQ, _ := utils.NewJsonQ(workloadSpec)

			mergeStack := enforcer.NewMergeStack(workloadSpecQ, "test", emptyPolicyQ.JumpToTop(), true)
			mergeStack.AddLevel(policySpecQ, enforcer.LevelPolicy, "test")

			err := mergeStack.MergeAndValidate(ctx, enforcer.MergeAndValidateOptions{})
			Expect(err).To(BeNil(), "MergeAndValidate Func should succeed")

			var workload jobs.JobWorkload
			actualWorkload, err := jobs.UnmarshalWorkload(mergeStack.MergedValues(), &workload.TypeMeta, &workload.ObjectMeta)
			Expect(err).To(BeNil(), "Expecting to get the actual spec")

			expectedWorkloadSpec := jobs.JobSpec{
				WorkloadFields: field_groups.WorkloadFields{
					ContainerFields: field_groups.ContainerFields{
						Annotations: &field_types.DictionaryField{
							Value: &map[string]field_types.DictionaryItemAndRules{
								"ann_1": {
									Value: ptr.To("has_value"),
								},
								"ann_2": {
									Value: ptr.To("default_value"),
								},
							},
						},
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"none": {
									Value: &field_types.PvcsItemValue{
										StorageClass: ptr.To("default_sc"),
										Path:         ptr.To("default_path"),
									},
								},
								"only_sc": {
									Value: &field_types.PvcsItemValue{
										StorageClass: ptr.To("has_sc"),
										Path:         ptr.To("default_path"),
										ClaimName:    ptr.To("claim1"),
										Size:         ptr.To("10M"),
									},
								},
								"only_path": {
									Value: &field_types.PvcsItemValue{
										StorageClass: ptr.To("default_sc"),
										Path:         ptr.To("has_path"),
										ClaimName:    ptr.To("claim1"),
										Size:         ptr.To("10M"),
									},
								},
								"has_both": {
									Value: &field_types.PvcsItemValue{
										Path:         ptr.To("has_path"),
										StorageClass: ptr.To("has_sc"),
										ClaimName:    ptr.To("claim1"),
										Size:         ptr.To("10M"),
									},
								},
							},
						},
					},
				},
			}
			Expect(actualWorkload.Spec).To(Equal(expectedWorkloadSpec))
		})
	})

	Context("Verify one policy per namespace ", func() {
		var (
			k8sClient client.Client
		)
		BeforeEach(func() {
			k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)

			genericPolicyValidator = workloadspolicy.NewPolicyValidator(k8sClient, InitInferencePolicyFromReq, []string{})
			genericPolicyValidator.PolicySyncEnabledFlag = true

			admissionFakeRequest = admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					RequestKind: &metav1.GroupVersionKind{},
				}}
			admissionFakeRequest.RequestKind.Kind = constants.KindInferencePolicy
			admissionFakeRequest.UserInfo.Username = "system:serviceaccount:runai:workload-controller"
			admissionFakeRequest.Namespace = constants.NamespaceRunAi
			admissionFakeRequest.Name = "inference-policy"
		})

		It("cant create more than one policy for the same project/department/tenant ", func() {

			basePolicy := inferencepolicy.InferencePolicy{
				ObjectMeta: resource.NewRunaiObject("gpuBelowZero"),
				Spec: inferencepolicy.InferenceSpec{
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						Gpu: &field_types.NumberField{
							Value: ptr.To("0"),
						},
					},
				},
			}
			basePolicy.Labels = map[string]string{
				constants.PolicyLabel:       "true",
				constants.ProjectScopeLabel: "1",
			}
			dumpJson, err := json.Marshal(basePolicy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal gpuBelowZero policy template")

			admissionFakeRequest.Object.Raw = dumpJson

			err = k8sClient.Create(ctx, &basePolicy)
			Expect(err).ToNot(HaveOccurred())

			policy, err := genericPolicyValidator.CreatePolicyAdapterFN(admissionFakeRequest)
			Expect(err).NotTo(HaveOccurred())
			admissionResponse := genericPolicyValidator.HandleCreate(ctx, policy, admissionFakeRequest)
			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("Cant create new policies while there are 1 policies with the same labels and namespace"))))
		})

	})

	Context("Rules Verification & Syntax", func() {

		BeforeEach(func() {
			k8sClient := client2.InitFakeClientWithObjectNoStatus(scheme)
			genericPolicyValidator = workloadspolicy.NewPolicyValidator(k8sClient, InitInferencePolicyFromReq, []string{})
		})

		DescribeTable("gpu.value: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("gpuBelowZero")
			policy.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
				Gpu: &field_types.NumberField{
					Value: ptr.To("-5.4"),
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal gpuBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("gpu.value: must be no less than 0"))))
		}, testEntries)

		DescribeTable("gpuLimit.value: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("gpuLimitBelowZero")
			policy.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
				GpuLimit: &field_types.NumberField{
					Value: ptr.To("-5.5"),
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal gpuLimitBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("gpuLimit.value: must be no less than 0"))))
		}, testEntries)

		DescribeTable("cpu.value: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("cpuBelowZero")
			policy.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
				Cpu: &field_types.NumberField{
					Value: ptr.To("-54"),
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal cpuBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("cpu.value: must be no less than 0"))))
		}, testEntries)

		DescribeTable("runAsUid.value: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("procsBelowZero")
			policy.Spec.RunAsUid = &field_types.IntField{
				Value: ptr.To(-3),
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal runAsUidBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("runAsUid.value: must be no less than 0"))))
		}, testEntries)

		DescribeTable("cpuLimit.value: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("cpuLimitBelowZero")
			policy.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
				CpuLimit: &field_types.NumberField{
					Value: ptr.To("-3.3"),
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal cpuLimitBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("cpuLimit.value: must be no less than 0"))))
		}, testEntries)

		DescribeTable("ports.toolow.container: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("containerPortBelowOne")
			policy.Spec.NetworkingFields.Ports = &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"toolow": {
						Value: &field_types.PortItemValue{
							ContainerPort: ptr.To(-1),
						},
					},
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal containerPortBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.toolow.container: must be no less than 0"))))
		}, testEntries)

		DescribeTable("ports.toolow.external: must be no less than 0", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("hostPortBelowOne")
			policy.Spec.NetworkingFields.Ports = &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"toolow": {
						Value: &field_types.PortItemValue{
							HostPort: ptr.To(-1),
						},
					},
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal hostPortBelowZero policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.toolow.external: must be no less than 0"))))
		}, testEntries)

		DescribeTable("runAsUid.min must be no greater than runAsUid.max", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("minMaxErr")
			policy.Spec.RunAsUid = workloadspolicy.IntFieldMinMax(ptr.To(5), ptr.To(3))

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal minMaxErr policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("runAsUid.min must be no greater than runAsUid.max"))))
		}, testEntries)

		DescribeTable("runAsUid.min: must be no less than 0", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("minLowerThanRunaiMin")
			policy.Spec.RunAsUid = workloadspolicy.IntFieldMinMax(ptr.To(-1), nil)

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal minLowerThanRunaiMin policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("runAsUid.min: must be no less than 0"))))
		}, testEntries)

		DescribeTable("runAsUid.max: must be no less than 0", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("maxLowerThanRunaiMin")
			policy.Spec.RunAsUid = workloadspolicy.IntFieldMinMax(nil, ptr.To(-6))

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal maxLowerThanRunaiMin policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("runAsUid.max: must be no less than 0"))))
		}, testEntries)

		DescribeTable("containerPorts.min must be no greater than containerPorts.max", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("minMaxErr")
			policy.Spec.NetworkingFields.Ports = containerPortMinMax(ptr.To(5), ptr.To(3))

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal minMaxErr policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.itemRules.container.min must be no greater than ports.itemRules.container.max"))))
		}, testEntries)

		DescribeTable("ports.itemRules.container.min: must be no less than 0", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("minLowerThanRunaiMin")
			policy.Spec.NetworkingFields.Ports = containerPortMinMax(ptr.To(-1), nil)

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal minLowerThanRunaiMin policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.itemRules.container.min: must be no less than 0"))))
		}, testEntries)

		DescribeTable("ports.itemRules.container.max: must be no less than 0", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("maxLoweRThanRunaiMin")
			policy.Spec.NetworkingFields.Ports = containerPortMinMax(nil, ptr.To(-1))

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal maxLowerThanRunaiMin policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.itemRules.container.max: must be no less than 0"))))
		}, testEntries)

		DescribeTable("ports.deletedNotOk: 'deleted' is not allowed in a policy", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("deletedItemIsForbidden")
			policy.Spec.NetworkingFields.Ports = &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"valueIsOk": {
						Value: &field_types.PortItemValue{},
					},
					"deletedNotOk": {
						Deleted: ptr.To(true),
					},
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal deletedItemIsForbidden policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.deletedNotOk: 'deleted' is not allowed in a policy"))))
		}, testEntries)

		DescribeTable("ports.ruleOnly: 'value' is missing", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("valueRequiresRules")
			policy.Spec.Ports = &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"ruleAndValue": {
						Rules: &field_types.MapItemRules{
							CanRemove: ptr.To(false),
						},
						Value: &field_types.PortItemValue{
							ContainerPort: ptr.To(1234),
						},
					},
					"ruleOnly": {
						Rules: &field_types.MapItemRules{
							CanRemove: ptr.To(false),
						},
					},
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal valueRequiresRules policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Errored(http.StatusInternalServerError, errors.New("ports.ruleOnly: 'value' is missing"))))
		}, testEntries)

		DescribeTable("runAsUid.min & runAsUid.max are aligned", func(assetsTest AssetsTest) {

			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("minMaxAlignedWithRunai")
			policy.Spec.RunAsUid = workloadspolicy.IntFieldMinMax(ptr.To(4), ptr.To(6))

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal minMaxOk policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Allowed("Ok")))
		}, testEntries)

		DescribeTable("containerPorts.min & containerPorts.max are aligned", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("minMaxAlignedWithRunai")
			policy.Spec.NetworkingFields.Ports = containerPortMinMax(ptr.To(0), ptr.To(32767))

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal minMaxAlignedWithRunai policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Allowed("Ok")))
		}, testEntries)

		DescribeTable("Cant edit valid policy template", func(assetsTest AssetsTest) {
			policy := jobs.JobWorkload{}
			policy.ObjectMeta = resource.NewRunaiObject("valueRequiresRules")
			policy.Spec.Ports = &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"myvolume": {
						Rules: &field_types.MapItemRules{
							CanEdit: ptr.To(false),
						},
						Value: &field_types.PortItemValue{
							ContainerPort: ptr.To(1234),
						},
					},
				},
			}

			dumpJson, err := json.Marshal(policy)
			Expect(err).ToNot(HaveOccurred(), "failed to unmarshal validCantEdit policy template")

			assetsTest.AdmissionFakeRequest.Object.Raw = dumpJson
			admissionResponse = handleCreate(ctx, assetsTest, genericPolicyValidator)

			Expect(admissionResponse).To(Equal(admission.Allowed("Ok")))
		}, testEntries)
	})

})

func handleCreate(ctx context.Context, assetsTest AssetsTest, genericPolicyValidator *workloadspolicy.ValidatorHandler) admission.Response {

	policy, err := genericPolicyValidator.CreatePolicyAdapterFN(assetsTest.AdmissionFakeRequest)
	Expect(err).NotTo(HaveOccurred())

	if assetsTest.PolicyType == constants.KindInferenceWorkload {
		return genericPolicyValidator.HandleCreate(ctx, policy, assetsTest.AdmissionFakeRequest)
	}

	return genericPolicyValidator.HandleCreate(ctx, policy, assetsTest.AdmissionFakeRequest)
}

func createAdmissionFakeRequest(kind, name string) admission.Request {
	InferenceAdmissionFakeRequest := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			RequestKind: &metav1.GroupVersionKind{},
		}}
	InferenceAdmissionFakeRequest.RequestKind.Kind = kind
	InferenceAdmissionFakeRequest.UserInfo.Username = "system:serviceaccount:runai:workload-controller"
	InferenceAdmissionFakeRequest.Namespace = constants.NamespaceRunAi
	InferenceAdmissionFakeRequest.Name = name
	return InferenceAdmissionFakeRequest
}

func containerPortMinMax(min *int, max *int) *field_types.PortsField {
	return &field_types.PortsField{
		ItemRules: &field_types.PortItemRules{
			ContainerPort: &field_types.IntRules{
				Min: min,
				Max: max,
			},
		},
	}
}
