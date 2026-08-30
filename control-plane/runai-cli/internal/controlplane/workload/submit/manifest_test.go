package submit

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManifestData", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "manifest-test-*")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Context("ReadAndParseManifest", func() {
		It("should successfully parse a valid manifest", func() {
			manifestContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: test-namespace
spec:
  containers:
  - name: test-container
    image: nginx:latest`

			manifestPath := filepath.Join(tmpDir, "valid-manifest.yaml")
			err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
			Expect(err).To(BeNil())

			data, err := ReadAndParseManifest(manifestPath)
			Expect(err).To(BeNil())
			Expect(data).NotTo(BeNil())
			Expect(data.NameFromYAML).To(Equal("test-pod"))
			Expect(data.NamespaceFromYAML).To(Equal("test-namespace"))
			Expect(data.KindFromYAML).To(Equal("Pod"))
			Expect(data.ParsedManifest).NotTo(BeNil())
			Expect(data.ParsedManifest["apiVersion"]).To(Equal("v1"))
		})

		It("should parse manifest without namespace", func() {
			manifestContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx:latest`

			manifestPath := filepath.Join(tmpDir, "no-namespace.yaml")
			err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
			Expect(err).To(BeNil())

			data, err := ReadAndParseManifest(manifestPath)
			Expect(err).To(BeNil())
			Expect(data.NameFromYAML).To(Equal("test-pod"))
			Expect(data.NamespaceFromYAML).To(Equal(""))
		})

		It("should parse manifest without name", func() {
			manifestContent := `apiVersion: v1
kind: Pod
metadata: {}
spec:
  containers:
  - name: test-container
    image: nginx:latest`

			manifestPath := filepath.Join(tmpDir, "no-name.yaml")
			err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
			Expect(err).To(BeNil())

			data, err := ReadAndParseManifest(manifestPath)
			Expect(err).To(BeNil())
			Expect(data.NameFromYAML).To(Equal(""))
		})

		It("should fail when file does not exist", func() {
			data, err := ReadAndParseManifest("/nonexistent/file.yaml")
			Expect(err).NotTo(BeNil())
			Expect(data).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to read manifest file"))
		})

		It("should fail with invalid YAML", func() {
			manifestContent := `this is not: valid: yaml::`

			manifestPath := filepath.Join(tmpDir, "invalid.yaml")
			err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
			Expect(err).To(BeNil())

			data, err := ReadAndParseManifest(manifestPath)
			Expect(err).NotTo(BeNil())
			Expect(data).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse YAML manifest"))
		})

		It("should parse PyTorchJob manifest", func() {
			manifestContent := `apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  name: pytorch-job
  namespace: ml-project
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      template:
        spec:
          containers:
          - name: pytorch
            image: pytorch/pytorch:latest`

			manifestPath := filepath.Join(tmpDir, "pytorch-job.yaml")
			err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
			Expect(err).To(BeNil())

			data, err := ReadAndParseManifest(manifestPath)
			Expect(err).To(BeNil())
			Expect(data.NameFromYAML).To(Equal("pytorch-job"))
			Expect(data.NamespaceFromYAML).To(Equal("ml-project"))
			Expect(data.KindFromYAML).To(Equal("PyTorchJob"))
			Expect(data.ParsedManifest["apiVersion"]).To(Equal("kubeflow.org/v1"))
		})
	})

})
