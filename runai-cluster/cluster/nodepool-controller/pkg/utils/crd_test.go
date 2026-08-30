package utils

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUtilsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

const testCRDName = "widgets.example.com"

// establishedCRD builds a CRD object with the given name and Established condition status.
func establishedCRD(name string, established bool) *apiextensionsv1.CustomResourceDefinition {
	status := apiextensionsv1.ConditionFalse
	if established {
		status = apiextensionsv1.ConditionTrue
	}
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
				{Type: apiextensionsv1.Established, Status: status},
			},
		},
	}
}

var _ = Describe("IsCRDInstalled", func() {
	var ctx context.Context

	readerWith := func(objs ...client.Object) client.Reader {
		scheme := runtime.NewScheme()
		Expect(apiextensionsv1.AddToScheme(scheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	}

	BeforeEach(func() { ctx = context.Background() })

	It("returns true when the CRD exists and is established", func() {
		Expect(IsCRDInstalled(ctx, readerWith(establishedCRD(testCRDName, true)), testCRDName)).To(BeTrue())
	})

	It("returns false when the CRD does not exist", func() {
		Expect(IsCRDInstalled(ctx, readerWith(), testCRDName)).To(BeFalse())
	})

	It("returns false when the CRD exists but is not established", func() {
		Expect(IsCRDInstalled(ctx, readerWith(establishedCRD(testCRDName, false)), testCRDName)).To(BeFalse())
	})
})
