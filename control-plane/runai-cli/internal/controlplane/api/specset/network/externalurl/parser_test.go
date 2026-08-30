package externalurl

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "container=9443,url=https://external.runai.com,authusers=user1:user2:user3,authgroups=group1:group2"
	NewMinimalExample = "container=9443"
	InvalidExample    = "bla"

	ErrorMissingMandatory = "cannot create object, missing mandatory fields [container], in value: %s"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "New External URL Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Container)).To(Equal("9443"))
			Expect(*ap.Get(Url)).To(Equal("https://external.runai.com"))
			Expect(*ap.Get(AuthUsers)).To(Equal("user1:user2:user3"))
			Expect(*ap.Get(AuthGroups)).To(Equal("group1:group2"))
		})
	})

	Context("New External URL validation", func() {
		It("Set everything", func() {
			eurl, err := NewExternalUrl(NewFullExample)
			Expect(err).To(BeNil())
			Expect(eurl.GetContainer()).To(Equal(int32(9443)))
			Expect(eurl.GetUrl()).To(Equal("https://external.runai.com"))
			Expect(eurl.GetAuthorizedUsers()).To(Equal([]string{"user1", "user2", "user3"}))
			Expect(eurl.GetAuthorizedGroups()).To(Equal([]string{"group1", "group2"}))
		})

		It("Set minimal", func() {
			eurl, err := NewExternalUrl(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(eurl.GetContainer()).To(Equal(int32(9443)))
			Expect(eurl.HasUrl()).To(Equal(false))
			Expect(eurl.HasAuthorizedUsers()).To(Equal(false))
			Expect(eurl.HasAuthorizedGroups()).To(Equal(false))
		})
	})

	Context("Invalid tests", func() {
		It("Invalid string", func() {
			_, err := NewExternalUrl(InvalidExample)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(fmt.Errorf(ErrorMissingMandatory, InvalidExample)))
		})
	})
})
