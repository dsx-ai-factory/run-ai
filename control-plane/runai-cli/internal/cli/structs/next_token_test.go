package structs_test

import (
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NextToken", func() {
	Context("NextTokenFromValue function", func() {
		It("empty", func() {
			_, err := structs.NextTokenFromValue("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("parsing time \"\" as"))
		})

		It("no encoded", func() {
			_, err := structs.NextTokenFromValue("bla")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("illegal base64 data"))
		})

		It("encoded token to unix time 1", func() {
			nt, err := structs.NextTokenFromValue("MTk3MC0wMS0wMVQwMjowMDowMSswMjowMA==")
			Expect(err).ToNot(HaveOccurred())
			Expect(nt.String()).To(HavePrefix("1970-01-01T02:00:01+02:00"))
		})
	})

	Context("NextTokenFromTimestamp function", func() {
		It("zero", func() {
			_, err := structs.NextTokenFromTimestamp(time.Time{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("next token cannot be with 0 timestamp"))
		})

		It("one sec", func() {
			nt, err := structs.NextTokenFromTimestamp(time.Unix(1, 0))
			Expect(err).ToNot(HaveOccurred())
			// in local it uses UTC +2 where in make test it uses UTC, so the prefix is the same
			Expect(nt.String()).To(HavePrefix("1970-01-01T"))
			Expect(nt.String()).To(ContainSubstring("00:01"))
			// in local it uses UTC +2 where in make test it uses UTC, so the prefix is the same
			Expect(nt.Encode()).To(HavePrefix("MTk3MC0wMS0wMVQwM"))
		})
	})
})
