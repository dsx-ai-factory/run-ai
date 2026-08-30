package policy

import (
	"context"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ControlPlanePoliciesUsage", func() {
	var (
		// fakeClient client.Client
		mockCtrl   *gomock.Controller
		scheme     *runtime.Scheme
		basePolicy ipv1.InteractivePolicy
		k8sClient  client.Client
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
		mockCtrl = gomock.NewController(GinkgoT())

		basePolicy = ipv1.InteractivePolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "policy",
				Labels: map[string]string{},
			},
			Spec: ipv1.InteractiveSpec{
				ResourceAllocationFields: field_groups.ResourceAllocationFields{
					Gpu: &flds.NumberField{
						Value: ptr.To("0"),
						Rules: &flds.NumberRules{
							Max: ptr.To("6"),
						},
					},
				},
			},
		}

	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("when getting existing policies indicator in config map", func() {
		Context("GetExistingPoliciesIndicatorConfigMap", func() {
			It("should return error if config map does not exist", func() {
				_, err := GetExistingPoliciesIndicatorConfigMap(context.Background(), k8sClient)
				Expect(err).To(HaveOccurred())
			})

			It("should return config map if it exists", func() {

				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ExistingPoliciesConfigMapName,
						Namespace: ExistingPoliciesConfigMapNamespace,
					},
				}
				Expect(k8sClient.Create(context.Background(), configMap)).To(Succeed())
				retrievedConfigMap, err := GetExistingPoliciesIndicatorConfigMap(context.Background(), k8sClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrievedConfigMap.Name).To(Equal(configMap.Name))
				Expect(retrievedConfigMap.Namespace).To(Equal(configMap.Namespace))
			})
		})
		Context("CheckOldPoliciesExistance", func() {
			It("CheckOldPoliciesExistance should return configmap value if exists", func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ExistingPoliciesConfigMapName,
						Namespace: ExistingPoliciesConfigMapNamespace,
					},
					Data: map[string]string{OldPolicyExistsConfigMapKey: "true"},
				}
				Expect(k8sClient.Create(context.Background(), configMap)).To(Succeed())
				isExist, err := CheckOldPoliciesExistance(context.Background(), k8sClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(isExist).To(BeTrue())
			})

			It("CheckOldPoliciesExistance should create a configmap if it does not exists", func() {
				_, preCreateErr := GetExistingPoliciesIndicatorConfigMap(context.Background(), k8sClient)
				Expect(preCreateErr).To(HaveOccurred())
				Expect(preCreateErr.Error()).To(ContainSubstring("not found"))

				_, err := CheckOldPoliciesExistance(context.Background(), k8sClient)
				Expect(err).NotTo(HaveOccurred())

				_, postCreateErr := GetExistingPoliciesIndicatorConfigMap(context.Background(), k8sClient)
				Expect(postCreateErr).NotTo(HaveOccurred())
			})

			Context("with policies", func() {
				It("CheckOldPoliciesExistance should return TRUE if policy without label exists", func() {
					listInteractive := &ipv1.InteractivePolicyList{}

					labelSelector, labelErr := createNotExistingLabelSelector(clusterCommonConstants.PolicyLabel)
					Expect(labelErr).To(BeNil())

					interactivePolicyList := k8sClient.List(context.Background(), listInteractive, labelSelector, client.Limit(1))
					Expect(interactivePolicyList).ToNot(HaveOccurred())
					Expect(listInteractive.Items).To(HaveLen(0))

					policyWithLabel := basePolicy

					policyWithLabel.Namespace = clusterCommonConstants.NamespaceRunAi
					policyWithLabel.Labels = map[string]string{
						clusterCommonConstants.DepartmentScopeLabel: "1",
					}
					policyWithLabel.ResourceVersion = ""

					err := k8sClient.Create(context.Background(), &policyWithLabel)
					Expect(err).ToNot(HaveOccurred())

					interactivePolicyList2 := k8sClient.List(context.Background(), listInteractive, labelSelector, client.Limit(1))
					Expect(interactivePolicyList2).ToNot(HaveOccurred())
					Expect(listInteractive.Items).To(HaveLen(1))

					_, preCreateErr := GetExistingPoliciesIndicatorConfigMap(context.Background(), k8sClient)
					Expect(preCreateErr).To(HaveOccurred())
					Expect(preCreateErr.Error()).To(ContainSubstring("not found"))

					isExist, err := CheckOldPoliciesExistance(context.Background(), k8sClient)
					Expect(err).NotTo(HaveOccurred())
					Expect(isExist).To(BeTrue())
				})

				It("CheckOldPoliciesExistance should return FALSE if policy exists but with label", func() {
					policyWithLabel := basePolicy

					policyWithLabel.Namespace = clusterCommonConstants.NamespaceRunAi
					policyWithLabel.Labels = map[string]string{
						clusterCommonConstants.PolicyLabel:          "true",
						clusterCommonConstants.DepartmentScopeLabel: "1",
					}
					policyWithLabel.ResourceVersion = ""

					err := k8sClient.Create(context.Background(), &policyWithLabel)
					Expect(err).ToNot(HaveOccurred())

					isExist, err := CheckOldPoliciesExistance(context.Background(), k8sClient)
					Expect(err).NotTo(HaveOccurred())
					Expect(isExist).To(BeFalse())
				})
			})
		})

		Context("createNotExistingLabelSelector", func() {
			It("creating 2 policies 1 with label 1 without -> should find 1 item", func() {
				listInteractiveType := &ipv1.InteractivePolicyList{}

				labelSelector, labelErr := createNotExistingLabelSelector(clusterCommonConstants.PolicyLabel)
				Expect(labelErr).To(BeNil())

				listErr := k8sClient.List(context.Background(), listInteractiveType, labelSelector)
				Expect(listErr).ToNot(HaveOccurred())
				Expect(listInteractiveType.Items).To(HaveLen(0))

				policyWithLabel := ipv1.InteractivePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "policy-with-label",
						Namespace: clusterCommonConstants.NamespaceRunAi,
						Labels: map[string]string{
							clusterCommonConstants.PolicyLabel: "true",
						},
					},
					Spec: ipv1.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: &flds.NumberField{
								Value: ptr.To("0"),
								Rules: &flds.NumberRules{
									Max: ptr.To("6"),
								},
							},
						},
					},
				}

				withLabelErr := k8sClient.Create(context.Background(), &policyWithLabel)
				Expect(withLabelErr).ToNot(HaveOccurred())

				policyWithoutLabel := ipv1.InteractivePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "policy-without-label",
						Labels:    map[string]string{},
						Namespace: clusterCommonConstants.NamespaceRunAi,
					},
					Spec: ipv1.InteractiveSpec{
						ResourceAllocationFields: field_groups.ResourceAllocationFields{
							Gpu: &flds.NumberField{
								Value: ptr.To("0"),
								Rules: &flds.NumberRules{
									Max: ptr.To("6"),
								},
							},
						},
					},
				}

				withoutLabelError := k8sClient.Create(context.Background(), &policyWithoutLabel)
				Expect(withoutLabelError).ToNot(HaveOccurred())

				listInteractive2 := &ipv1.InteractivePolicyList{}
				list2Err := k8sClient.List(context.Background(), listInteractive2, labelSelector)
				Expect(list2Err).ToNot(HaveOccurred())
				Expect(len(listInteractive2.Items)).To(Equal(1))
			})
		})
		Context("SetupControlPlanePoliciesUsage", func() {
			It("should return TRUE if cluster has no existing policies", func() {
				useControlPlanePolicies := SetupControlPlanePoliciesUsage(context.Background(), k8sClient)
				Expect(useControlPlanePolicies).To(BeTrue())

			})

			It("should return TRUE if all existing policies have the control plane label", func() {
				policyWithLabel := basePolicy

				policyWithLabel.Namespace = clusterCommonConstants.NamespaceRunAi
				policyWithLabel.Labels = map[string]string{
					clusterCommonConstants.PolicyLabel:          "true",
					clusterCommonConstants.DepartmentScopeLabel: "1",
				}
				policyWithLabel.ResourceVersion = ""

				err := k8sClient.Create(context.Background(), &policyWithLabel)
				Expect(err).ToNot(HaveOccurred())

				useControlPlanePolicies := SetupControlPlanePoliciesUsage(context.Background(), k8sClient)
				Expect(useControlPlanePolicies).To(BeTrue())
			})

			It("should return TRUE if policies without label exist but APPLY_CONTROL_PLANE_POLICIES runaiconfig variable is ON", func() {
				os.Setenv(utils.EnvVarApplyControlPlanePolicies, "true")

				policyWithLabel := basePolicy

				policyWithLabel.Namespace = clusterCommonConstants.NamespaceRunAi
				policyWithLabel.Labels = map[string]string{}
				policyWithLabel.ResourceVersion = ""

				err := k8sClient.Create(context.Background(), &policyWithLabel)
				Expect(err).ToNot(HaveOccurred())

				useControlPlanePolicies := SetupControlPlanePoliciesUsage(context.Background(), k8sClient)
				Expect(useControlPlanePolicies).To(BeTrue())
				// cleanup
				os.Unsetenv(utils.EnvVarApplyControlPlanePolicies)
			})

			It("should return FALSE if policies without label exist but  APPLY_CONTROL_PLANE_POLICIES runaiconfig variable is ON", func() {
				policyWithLabel := basePolicy

				policyWithLabel.Namespace = clusterCommonConstants.NamespaceRunAi
				policyWithLabel.Labels = map[string]string{}
				policyWithLabel.ResourceVersion = ""

				err := k8sClient.Create(context.Background(), &policyWithLabel)
				Expect(err).ToNot(HaveOccurred())

				useControlPlanePolicies := SetupControlPlanePoliciesUsage(context.Background(), k8sClient)
				Expect(useControlPlanePolicies).To(BeFalse())
			})
		})
	})
})
