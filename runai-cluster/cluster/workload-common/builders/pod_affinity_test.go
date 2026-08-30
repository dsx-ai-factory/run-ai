package builders

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodAffinity tests", func() {
	var (
		podAffinity                  bool
		podAffinityTopology          string
		podAffinitySchedulingRule    string
		affinity                     *corev1.Affinity
		affinityPtr                  **corev1.Affinity
		jobName                      string
		isDistributed                bool
		expectedRequiredAffinityLen  int
		expectedPreferredAffinityLen int
	)

	BeforeEach(func() {
		// Setup common test variables
		podAffinity = true
		podAffinityTopology = "kubernetes.io/hostname"
		podAffinitySchedulingRule = constants.RequiredDuringSchedulingIgnoredDuringExecutionValue
		affinity = nil
		affinityPtr = &affinity
		jobName = "test-job"
		isDistributed = false
		expectedRequiredAffinityLen = 0
		expectedPreferredAffinityLen = 0
	})

	Context("Basic functionality", func() {
		It("should do nothing when PodAffinity is false", func() {
			podAffinity = false

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity).To(BeNil())
		})

		It("should do nothing when affinity pointer is nil", func() {
			affinityPtr = nil

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should create affinity when none exists yet", func() {
			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity).NotTo(BeNil())
			Expect(affinity.PodAffinity).NotTo(BeNil())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
			Expect(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())

			podAffinityTerm := affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0]
			Expect(podAffinityTerm.TopologyKey).To(Equal(podAffinityTopology))
			Expect(podAffinityTerm.LabelSelector.MatchExpressions).To(HaveLen(1))
			Expect(podAffinityTerm.LabelSelector.MatchExpressions[0].Key).To(Equal(constants.LabelRelease))
			Expect(podAffinityTerm.LabelSelector.MatchExpressions[0].Values).To(ContainElement(jobName))
		})
	})

	Context("Scheduling rules", func() {
		It("should use required affinity when specified", func() {
			podAffinitySchedulingRule = constants.RequiredDuringSchedulingIgnoredDuringExecutionValue

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
			Expect(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
		})

		It("should use preferred affinity when specified", func() {
			podAffinitySchedulingRule = constants.PreferredDuringSchedulingIgnoredDuringExecutionValue

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
			Expect(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))

			preferredTerm := affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
			Expect(preferredTerm.Weight).To(Equal(int32(UsersAffinityWeight)))
			Expect(preferredTerm.PodAffinityTerm.TopologyKey).To(Equal(podAffinityTopology))
		})

		It("should return nil when invalid scheduling rule provided", func() {
			podAffinitySchedulingRule = "InvalidRule"

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity).To(BeNil())
		})

		It("should use default scheduling rule when empty", func() {
			podAffinitySchedulingRule = ""

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
			Expect(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
		})
	})

	Context("Existing affinity", func() {
		BeforeEach(func() {
			// Setup existing affinity
			existingNodeAffinity := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "gpu-type",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"nvidia"},
								},
							},
						},
					},
				},
			}

			affinity = &corev1.Affinity{
				NodeAffinity: existingNodeAffinity,
			}
		})

		It("should add pod affinity to existing affinity", func() {
			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.NodeAffinity).NotTo(BeNil())
			Expect(affinity.PodAffinity).NotTo(BeNil())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
		})

		It("should add pod affinity to existing pod affinity", func() {
			// Create existing pod affinity
			existingPodAffinity := &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						TopologyKey: "zone",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"database"},
								},
							},
						},
					},
				},
			}

			affinity.PodAffinity = existingPodAffinity
			expectedRequiredAffinityLen = 1

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(expectedRequiredAffinityLen + 1))

			// Verify both original and new pod affinity terms exist
			foundOriginal := false
			foundNew := false

			for _, term := range affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				if term.TopologyKey == "zone" {
					foundOriginal = true
				} else if term.TopologyKey == podAffinityTopology {
					foundNew = true
				}
			}

			Expect(foundOriginal).To(BeTrue())
			Expect(foundNew).To(BeTrue())
		})

		It("should add preferred pod affinity to existing preferred pod affinity", func() {
			podAffinitySchedulingRule = constants.PreferredDuringSchedulingIgnoredDuringExecutionValue

			// Create existing preferred pod affinity
			existingPodAffinity := &corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 50,
						PodAffinityTerm: corev1.PodAffinityTerm{
							TopologyKey: "region",
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "app",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"cache"},
									},
								},
							},
						},
					},
				},
			}

			affinity.PodAffinity = existingPodAffinity
			expectedPreferredAffinityLen = 1

			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(expectedPreferredAffinityLen + 1))

			// Verify both original and new pod affinity terms exist
			foundOriginal := false
			foundNew := false

			for _, term := range affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				if term.PodAffinityTerm.TopologyKey == "region" {
					foundOriginal = true
				} else if term.PodAffinityTerm.TopologyKey == podAffinityTopology {
					foundNew = true
				}
			}

			Expect(foundOriginal).To(BeTrue())
			Expect(foundNew).To(BeTrue())
		})
	})

	Context("Distributed jobs", func() {
		BeforeEach(func() {
			isDistributed = true
		})

		It("should use different label selector for distributed jobs", func() {
			err := CreatePodAffinity(podAffinity, podAffinityTopology, podAffinitySchedulingRule, affinityPtr, jobName, isDistributed)

			Expect(err).NotTo(HaveOccurred())
			Expect(affinity).NotTo(BeNil())
			Expect(affinity.PodAffinity).NotTo(BeNil())
			Expect(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))

			podAffinityTerm := affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0]
			Expect(podAffinityTerm.LabelSelector.MatchExpressions).To(HaveLen(1))
			Expect(podAffinityTerm.LabelSelector.MatchExpressions[0].Key).To(Equal(constants.LabelTrainingKubeFlow))
			Expect(podAffinityTerm.LabelSelector.MatchExpressions[0].Values).To(ContainElement(jobName))
		})
	})
})
