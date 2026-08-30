package datavolume

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatavolumeParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datavolume Parser Suite")
}

var _ = Describe("Datavolume Parser", func() {
	Describe("ParseDataVolume", func() {
		It("should parse valid data volume string", func() {
			dv := "name=my-datavolume,mountpath=/data"
			parsed, err := ParseDataVolume(dv)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed.Name).To(Equal("my-datavolume"))
			Expect(parsed.MountPath).To(Equal("/data"))
		})

		It("should trim whitespace from values", func() {
			dv := "name= my-datavolume ,mountpath= /data "
			parsed, err := ParseDataVolume(dv)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed.Name).To(Equal("my-datavolume"))
			Expect(parsed.MountPath).To(Equal("/data"))
		})

		It("should fail when name is missing", func() {
			dv := "mountpath=/data"
			_, err := ParseDataVolume(dv)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datavolume format"))
			Expect(err.Error()).To(ContainSubstring("required format is name=NAME,mountpath=PATH"))
		})

		It("should fail when mountpath is missing", func() {
			dv := "name=my-datavolume"
			_, err := ParseDataVolume(dv)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datavolume format"))
			Expect(err.Error()).To(ContainSubstring("required format is name=NAME,mountpath=PATH"))
		})

		It("should fail on empty string", func() {
			_, err := ParseDataVolume("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datavolume format"))
		})

		It("should fail on invalid format", func() {
			dv := "invalid-format"
			_, err := ParseDataVolume(dv)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datavolume format"))
		})

		It("should fail on unknown key with helpful message", func() {
			dv := "naome=dv1,mountpath=/data"
			_, err := ParseDataVolume(dv)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datavolume format"))
			Expect(err.Error()).To(ContainSubstring("required format is name=NAME,mountpath=PATH"))
		})
	})

	Describe("ParseDataVolumeList", func() {
		It("should parse multiple data volumes", func() {
			dvList := []string{
				"name=dv1,mountpath=/data1",
				"name=dv2,mountpath=/data2",
			}
			parsed, err := ParseDataVolumeList(dvList)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(parsed)).To(Equal(2))
			Expect(parsed[0].Name).To(Equal("dv1"))
			Expect(parsed[0].MountPath).To(Equal("/data1"))
			Expect(parsed[1].Name).To(Equal("dv2"))
			Expect(parsed[1].MountPath).To(Equal("/data2"))
		})

		It("should return empty list for empty input", func() {
			parsed, err := ParseDataVolumeList([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(parsed)).To(Equal(0))
		})

		It("should fail on first invalid entry", func() {
			dvList := []string{
				"name=dv1,mountpath=/data1",
				"invalid",
			}
			_, err := ParseDataVolumeList(dvList)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewDataVolumeInstanceList", func() {
		It("should create instances with resolved IDs", func() {
			parsedList := []ParsedDataVolume{
				{Name: "dv1", MountPath: "/data1"},
				{Name: "dv2", MountPath: "/data2"},
			}
			nameToIdMap := map[string]string{
				"dv1": "uuid-1",
				"dv2": "uuid-2",
			}

			instances, err := NewDataVolumeInstanceList(parsedList, nameToIdMap)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(instances)).To(Equal(2))
			Expect(instances[0].GetId()).To(Equal("uuid-1"))
			Expect(instances[0].GetMountPath()).To(Equal("/data1"))
			Expect(instances[1].GetId()).To(Equal("uuid-2"))
			Expect(instances[1].GetMountPath()).To(Equal("/data2"))
		})

		It("should fail when data volume name is not found", func() {
			parsedList := []ParsedDataVolume{
				{Name: "dv1", MountPath: "/data1"},
				{Name: "non-existent", MountPath: "/data2"},
			}
			nameToIdMap := map[string]string{
				"dv1": "uuid-1",
			}

			_, err := NewDataVolumeInstanceList(parsedList, nameToIdMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("data volume 'non-existent' not found"))
		})
	})
})
