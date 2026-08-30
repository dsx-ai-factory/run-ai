package source

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
)

func TestKubeflowSuspendKartas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeflow suspend Kartas")
}

var _ = Describe("Kubeflow training Karta suspend definitions", func() {
	DescribeTable("exposes suspend and resume actions for each registered version",
		func(build func() models.Karta, expectedVersion, expectedKind string) {
			karta := build()
			root := karta.Spec.StructureDefinition.RootComponent

			Expect(root.Kind).NotTo(BeNil())
			Expect(root.Kind.Group).To(Equal("kubeflow.org"))
			Expect(root.Kind.Version).To(Equal(expectedVersion))
			Expect(root.Kind.Kind).To(Equal(expectedKind))
			Expect(root.SuspendDefinition).NotTo(BeNil())
			Expect(root.SuspendDefinition.SuspendActions).To(HaveLen(1))
			Expect(root.SuspendDefinition.ResumeActions).To(HaveLen(1))
			Expect(root.SuspendDefinition.SuspendActions[0].Path).To(Equal(".spec.runPolicy.suspend"))
			Expect(root.SuspendDefinition.SuspendActions[0].Value).To(Equal("true"))
			Expect(root.SuspendDefinition.ResumeActions[0].Path).To(Equal(".spec.runPolicy.suspend"))
			Expect(root.SuspendDefinition.ResumeActions[0].Value).To(Equal("false"))
		},
		Entry("PyTorch v1", BuildKubeflowPyTorchJobV1, "v1", "PyTorchJob"),
		Entry("TF v1", BuildKubeflowTFJobV1, "v1", "TFJob"),
		Entry("XGBoost v1", BuildKubeflowXGBoostJobV1, "v1", "XGBoostJob"),
		Entry("JAX v1", BuildKubeflowJAXJobV1, "v1", "JAXJob"),
		Entry("MPI v1", BuildKubeflowMPIJobV1, "v1", "MPIJob"),
		Entry("MPI v2beta1", BuildKubeflowMPIJobV2beta1, "v2beta1", "MPIJob"),
	)
})
