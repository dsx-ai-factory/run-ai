package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("imagePullSecrets test", func() {
	var (
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		// Setup common test variables
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

	})

	Context("both workload and service account", func() {
		It("should return one from each", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			ctx := context.Background()

			srvAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "my-ns",
				},
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "service-account"},
					{Name: "both"},
				},
			}

			err := fakeClient.Create(ctx, srvAccount)
			Expect(err).To(BeNil())

			result, err := CreateImagePullSecrets(ctx, fakeClient, "my-ns", "  workload  both ", "workload-1", "tw")
			Expect(err).To(BeNil())

			expected := datastructs.NewSet([]string{
				"workload",
				"both",
				"service-account",
			}...)

			actual := datastructs.NewSet[string]()

			for _, entry := range result {
				actual.Add(entry.Name)
			}

			Expect(actual).To(Equal(expected))
		})
	})

	Context("only workload", func() {
		It("should return workload list", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			ctx := context.Background()

			result, err := CreateImagePullSecrets(ctx, fakeClient, "my-ns", "workload-1    workload-2", "workload-1", "tw")
			Expect(err).To(BeNil())

			expected := datastructs.NewSet([]string{
				"workload-1",
				"workload-2",
			}...)

			actual := datastructs.NewSet[string]()

			for _, entry := range result {
				actual.Add(entry.Name)
			}

			Expect(actual).To(Equal(expected))
		})
	})

	Context("user credential imagePullSecret", func() {
		It("should return without prefix", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			ctx := context.Background()

			result, err := CreateImagePullSecrets(ctx, fakeClient, "my-ns", "USERCREDENTIAL:workload-1    workload-2", "workload-1", "tw")
			Expect(err).To(BeNil())

			expected := datastructs.NewSet([]string{
				"tw-workload-1-workload-1",
				"workload-2",
			}...)

			actual := datastructs.NewSet[string]()

			for _, entry := range result {
				actual.Add(entry.Name)
			}

			Expect(actual).To(Equal(expected))
		})
	})

	Context("incorrect format", func() {
		It("should be rejected", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			ctx := context.Background()

			_, err := CreateImagePullSecrets(ctx, fakeClient, "my-ns", "workload-1 workload43505^$$EDD-2", "workload-1", "tw")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("consist of lower case"))
		})
	})

	Context("no secrets in workload", func() {
		It("should return nil to allow k8s to do the regular job", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			ctx := context.Background()

			srvAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "my-ns",
				},
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "service-account-1"},
					{Name: "service-account-2"},
				},
			}

			err := fakeClient.Create(ctx, srvAccount)
			Expect(err).To(BeNil())

			result, err := CreateImagePullSecrets(ctx, fakeClient, "my-ns", "", "workload-1", "tw")
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})

})
