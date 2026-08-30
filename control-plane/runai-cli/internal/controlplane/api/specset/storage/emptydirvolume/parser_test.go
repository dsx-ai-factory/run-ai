package emptydirvolume

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample = "name=NAME,path=PATH,medium=MEDIUM,size-limit=SIZE_LIMIT"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EmptyDirVolume Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Name)).To(Equal("NAME"))
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(Medium)).To(Equal("MEDIUM"))
			Expect(*ap.Get(SizeLimit)).To(Equal("SIZE_LIMIT"))
		})
	})

	Context("New EmptyDirVolume validation", func() {
		It("Set everything", func() {
			emptyDirVolume, err := NewEmptyDirVolume(NewFullExample)
			Expect(err).To(BeNil())
			Expect(emptyDirVolume.GetName()).To(Equal("NAME"))
			Expect(emptyDirVolume.GetPath()).To(Equal("PATH"))
			Expect(emptyDirVolume.GetMedium()).To(Equal("MEDIUM"))
			Expect(emptyDirVolume.GetSizeLimit()).To(Equal("SIZE_LIMIT"))
		})
	})
})
