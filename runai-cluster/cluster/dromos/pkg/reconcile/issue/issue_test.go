package issue_test

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIssue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issue Suite")
}

var _ = Describe("issue.Group", func() {
	It("stringifies", func() {
		Expect(issue.Internal.String()).To(Equal("Internal"))
		Expect(issue.Network.String()).To(Equal("Network"))
		Expect(issue.Compute.String()).To(Equal("Compute"))
		Expect(issue.Storage.String()).To(Equal("Storage"))
	})
})

var _ = Describe("issue.New", func() {
	var (
		secret *corev1.Secret
		err    error
	)

	BeforeEach(func() {
		secret = &corev1.Secret{
			TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "my-secret", Namespace: "ns"},
		}
		err = errors.New("boom")
	})

	It("populates fields from the object and args", func() {
		i := issue.New(secret, err, true, "could not create", issue.Storage)

		Expect(i.Name).To(Equal("my-secret"))
		Expect(i.GVK.Kind).To(Equal("Secret"))
		Expect(i.IsCritical).To(BeTrue())
		Expect(i.Group).To(Equal(issue.Storage))
		Expect(errors.Unwrap(i)).To(MatchError("boom"))
	})

	It("formats Message as [Kind:Name] description: error", func() {
		i := issue.New(secret, err, true, "could not create", issue.Storage)
		Expect(i.Error()).To(Equal("[Secret:my-secret] could not create: boom"))
	})

	It("formats Message with only description when error is nil", func() {
		i := issue.New(secret, nil, false, "warning only", issue.Network)
		Expect(i.Error()).To(Equal("[Secret:my-secret] warning only"))
	})

	It("formats Message with only error when description is empty", func() {
		i := issue.New(secret, err, true, "", issue.Internal)
		Expect(i.Error()).To(Equal("[Secret:my-secret] boom"))
	})

	It("non-critical when isCritical is false", func() {
		i := issue.New(secret, err, false, "warn", issue.Compute)
		Expect(i.IsCritical).To(BeFalse())
	})
})
