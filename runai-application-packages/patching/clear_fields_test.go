package patching

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"io"
	"os"
	"sigs.k8s.io/yaml"
)

var _ = Describe("removing fields test", func() {
	Context("remove fields from json payload", func() {
		It("removal mix", func() {
			payload := readYaml("input/payload_1.yaml")

			err := ClearFields(payload, []string{
				"spec.image",
				"spec.environmentVariables[1]",
				"spec.environmentVariables[3]",
				"spec.storage.pvc[0].claimName",
				"spec.storage.pvc[1].claimInfo.volumeMode",
				"array[0]",
				"array[1]",
			}, ClearFieldsOptions{})
			Expect(err).To(BeNil())

			expected := readYaml("expected/payload_1_test_1.yaml")
			testutils.Expect(payload).ToEqual(expected)
		})

		It("removal all", func() {
			payload := readYaml("input/payload_1.yaml")

			err := ClearFields(payload, []string{
				"spec",
				"array",
			}, ClearFieldsOptions{})
			Expect(err).To(BeNil())

			expected := readYaml("expected/payload_1_test_2.yaml")
			testutils.Expect(payload).ToEqual(expected)
		})
	})

	Context("non existent path", func() {
		ignore := ClearFieldsOptions{
			IgnoreNonExisting: true,
		}

		dontIgnore := ClearFieldsOptions{}

		origPayload := readYaml("input/payload_1.yaml")

		payload := readYaml("input/payload_1.yaml")

		It("non existing field - ignore dependent", func() {
			err := ClearFields(payload, []string{"oops"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("oops: field oops does not exist"))

			err = ClearFields(payload, []string{"oops"}, ignore)
			Expect(err).To(BeNil())
			Expect(payload).To(Equal(origPayload))
		})

		It("negative index - cannot be ignored", func() {
			err := ClearFields(payload, []string{"array[-1]"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid path 'array[-1]': expecting last part to be field[index]"))

			err = ClearFields(payload, []string{"array[-1]"}, ignore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid path 'array[-1]': expecting last part to be field[index]"))
		})

		It("inner field does not exist", func() {
			err := ClearFields(payload, []string{"array[1].none"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("array[1].none: field none does not exist"))

			err = ClearFields(payload, []string{"array[1].none"}, ignore)
			Expect(err).To(BeNil())
			Expect(payload).To(Equal(origPayload))
		})

		It("not an array, cannot be ignored", func() {
			err := ClearFields(payload, []string{"spec.workingDir[1]"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("field workingDir is not itemized"))

			err = ClearFields(payload, []string{"spec.workingDir[1]"}, ignore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("field workingDir is not itemized"))
		})

		It("array field not found, can be ignored", func() {
			err := ClearFields(payload, []string{"spec.workingDir[1].none"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("spec.workingDir[1].none not found: unknown parameter spec.workingDir.1"))

			err = ClearFields(payload, []string{"spec.workingDir[1].none"}, ignore)
			Expect(err).To(BeNil())
			Expect(payload).To(Equal(origPayload))
		})

		It("index exceeded, can be ignored", func() {
			err := ClearFields(payload, []string{"spec.environmentVariables[5]"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables[5] exceeds the number of instances 5"))

			err = ClearFields(payload, []string{"spec.environmentVariables[5]"}, ignore)
			Expect(err).To(BeNil())
			Expect(payload).To(Equal(origPayload))
		})

		It("index non numeric, cannot be ignored", func() {
			err := ClearFields(payload, []string{"spec.environmentVariables[a].name"}, dontIgnore)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("path spec.environmentVariables[a].name does not point to a field in the spec"))

			err = ClearFields(payload, []string{"spec.environmentVariables[5]"}, ignore)
			Expect(err).To(BeNil())
			Expect(payload).To(Equal(origPayload))
		})
	})
})

func readYaml(fname string) map[string]interface{} {
	file, err := os.Open(fname)
	Expect(err).To(BeNil())
	defer file.Close()

	data, err := io.ReadAll(file)
	Expect(err).To(BeNil())

	return toMap(string(data))
}

func toMap(yamlText string) map[string]interface{} {
	result := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(yamlText), &result)
	Expect(err).To(BeNil())

	return result
}
