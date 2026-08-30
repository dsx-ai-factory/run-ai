package admission_controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Admission Controller", func() {
	var (
		admissionController *AdmissionController
	)

	BeforeEach(func() {
		admissionController = &AdmissionController{
			Enabled: ptr.To[bool](true),
		}
	})

	It("nothing is set Replicas should be 1", func() {
		admissionController.SetDefaultsWhereNeeded("registry.example.com/runai", ptr.To[int32](1))
		Expect(*admissionController.Replicas).To(Equal(int32(1)))
		Expect(admissionController.Image).NotTo(BeNil())
		Expect(admissionController.Image.Registry).To(Equal(ptr.To("registry.example.com/runai")))
		Expect(admissionController.Image.Name).To(Equal(ptr.To("admission-controller")))
	})

	It("global.replicaCount is nil Replicas should be 1", func() {
		admissionController.SetDefaultsWhereNeeded("", nil)
		Expect(*admissionController.Replicas).To(Equal(int32(1)))
	})

	It("global.replicaCount is set, cluster.replicas is not set be equal to global.replicaCount", func() {
		admissionController.SetDefaultsWhereNeeded("", ptr.To[int32](2))
		Expect(*admissionController.Replicas).To(Equal(int32(2)))
	})

	It("global.replicaCount is set, cluster.replicas is set should bot change", func() {
		admissionController.Replicas = ptr.To[int32](2)
		admissionController.SetDefaultsWhereNeeded("", ptr.To[int32](3))
		Expect(*admissionController.Replicas).To(Equal(int32(2)))
	})

})
