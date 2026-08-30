package parser

import (
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	Empty                 = ""
	BadExample            = "source=REPO,branch=BRANCH_NAME,rev=REVISION,username=USER,password==PASSWORD,target=TARGET_DIRECTORY_TO_CLONE"
	GoodExample           = "source=REPO,branch=BRANCH_NAME,rev=REVISION,username=USER,password=PASSWORD,target=TARGET_DIRECTORY_TO_CLONE"
	GoodExampleWithSpaces = "source=REPO,   branch=BRANCH_NAME,   rev=REVISION,   username=USER   " +
		",password=PASSWORD, target=TARGET_DIRECTORY_TO_CLONE"
	GoodPartExampleWithSpaces = "source=REPO,   branch=BRANCH_NAME   ,password=PASSWORD"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Validation Tests")
}

var _ = Describe("Parser Tests", Ordered, func() {
	Context("Sanity validation", Ordered, func() {
		It("empty value", func() {
			ap := NewAttributeMapper(Empty, NewStrictMapParser("parser_name", mapset.NewSet[string]()))
			err := ap.Parse()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("value is an empty string"))
		})

		It("value with 2 equals", func() {
			ap := NewAttributeMapper(BadExample, NewLooseMapParser("parser_name",
				mapset.NewSet[string]("source", "branch", "rev", "username", "password", "target")))
			err := ap.Parse()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("value (password==PASSWORD) has multiple `=` signs"))
			Expect(ap.GetMap()).To(BeNil())
		})

		It("Good loose example - success", func() {
			ap := NewAttributeMapper(GoodExample,
				NewLooseMapParser("parser_name", mapset.NewSet[string]("source", "branch", "rev", "username", "password", "target")))
			Expect(ap.Parse()).To(BeNil())
			apMap := ap.GetMap()
			Expect(len(apMap)).To(Equal(6))
			Expect(apMap["source"]).To(Equal("REPO"))
			Expect(apMap["branch"]).To(Equal("BRANCH_NAME"))
			Expect(apMap["rev"]).To(Equal("REVISION"))
			Expect(apMap["username"]).To(Equal("USER"))
			Expect(apMap["password"]).To(Equal("PASSWORD"))
			Expect(apMap["target"]).To(Equal("TARGET_DIRECTORY_TO_CLONE"))
		})

		It("Good loose example with spaces - success", func() {
			ap := NewAttributeMapper(GoodExampleWithSpaces,
				NewLooseMapParser("parser_name", mapset.NewSet[string]("source", "branch", "rev", "username", "password", "target")))
			Expect(ap.Parse()).To(BeNil())
			apMap := ap.GetMap()
			Expect(len(apMap)).To(Equal(6))
			Expect(apMap["source"]).To(Equal("REPO"))
			Expect(apMap["branch"]).To(Equal("BRANCH_NAME"))
			Expect(apMap["rev"]).To(Equal("REVISION"))
			Expect(apMap["username"]).To(Equal("USER"))
			Expect(apMap["password"]).To(Equal("PASSWORD"))
			Expect(apMap["target"]).To(Equal("TARGET_DIRECTORY_TO_CLONE"))
		})

		It("Good loose example with spaces - success", func() {
			ap := NewAttributeMapper(GoodPartExampleWithSpaces,
				NewLooseMapParser("ok", mapset.NewSet[string]("source", "branch", "rev", "username", "password", "target")))
			Expect(ap.Parse()).To(BeNil())
			apMap := ap.GetMap()
			Expect(len(apMap)).To(Equal(3))
			Expect(apMap["source"]).To(Equal("REPO"))
			Expect(apMap["branch"]).To(Equal("BRANCH_NAME"))
			Expect(apMap["rev"]).To(Equal(""))
			Expect(apMap["username"]).To(Equal(""))
			Expect(apMap["password"]).To(Equal("PASSWORD"))
			Expect(apMap["target"]).To(Equal(""))
		})

		It("Strict parser - adding none dictionary attribute", func() {
			ap := NewAttributeMapper(GoodExample,
				NewStrictMapParser("parser_name", mapset.NewSet[string]("source", "branch", "password")))

			err := ap.Parse()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(HaveSuffix("dictionary does not contain key: rev, value: %s", GoodExample))
		})

		It("Strict parser - not all attribute has values", func() {
			ap := NewAttributeMapper(GoodExample,
				NewStrictMapParser("parser_name", mapset.NewSet[string]("source", "branch", "rev", "username", "password", "target", "missing_value")))

			err := ap.Parse()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("dictionary must contain all keys: [missing_value]"))
		})
	})
})
