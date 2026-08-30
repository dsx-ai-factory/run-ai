package gitsync

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NewFullExample    = "name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION"
	NewMinimalExample = "name=NAME,repository=REPO,path=PATH,secret=SECRET"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Parser Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("Set everything", func() {
			ap := NewParser(NewFullExample)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Name)).To(Equal("NAME"))
			Expect(*ap.Get(Repository)).To(Equal("REPO"))
			Expect(*ap.Get(Path)).To(Equal("PATH"))
			Expect(*ap.Get(Secret)).To(Equal("SECRET"))
			Expect(*ap.Get(Revision)).To(Equal("REVISION"))
		})
	})

	Context("New GitSync validation", func() {
		It("Set everything", func() {
			git, err := NewGit(NewFullExample)
			Expect(err).To(BeNil())
			Expect(git.GetName()).To(Equal("NAME"))
			Expect(git.GetRepository()).To(Equal("REPO"))
			Expect(git.GetPath()).To(Equal("PATH"))
			Expect(git.GetPasswordSecret()).To(Equal("SECRET"))
			Expect(git.GetRevision()).To(Equal("REVISION"))
		})

		It("Set minimal", func() {
			git, err := NewGit(NewMinimalExample)
			Expect(err).To(BeNil())
			Expect(git.GetName()).To(Equal("NAME"))
			Expect(git.GetRepository()).To(Equal("REPO"))
			Expect(git.GetPath()).To(Equal("PATH"))
			Expect(git.GetPasswordSecret()).To(Equal("SECRET"))
			Expect(git.HasRevision()).To(Equal(false))
			Expect(git.HasBranch()).To(Equal(false))
		})
	})
})
