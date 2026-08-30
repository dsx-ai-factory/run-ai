package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/secret-generator/pkg/generator"
)

var _ = Describe("GeneratePassword", func() {
	DescribeTable("should produce a password with the expected length",
		func(requestedLength, expectedLength int) {
			password, err := generator.GeneratePassword(requestedLength)
			Expect(err).NotTo(HaveOccurred())
			Expect(password).To(HaveLen(expectedLength))
		},
		Entry("standard 32-character password", 32, 32),
		Entry("short 8-character password at the minimum boundary", 8, 8),
		Entry("long 64-character password", 64, 64),
		Entry("requested length below minimum (4) is bumped to 8", 4, 8),
	)

	It("should include at least one character from each required class (lowercase, uppercase, digit, special)", func() {
		for i := 0; i < 50; i++ {
			password, err := generator.GeneratePassword(32)
			Expect(err).NotTo(HaveOccurred())

			Expect(password).To(MatchRegexp(`[a-z]`), "password should contain a lowercase letter")
			Expect(password).To(MatchRegexp(`[A-Z]`), "password should contain an uppercase letter")
			Expect(password).To(MatchRegexp(`[0-9]`), "password should contain a digit")
			Expect(password).To(MatchRegexp(`[!@#%^*()\-_=+]`), "password should contain a special character")
		}
	})

	It("should always start with a letter so it is safe to embed unquoted in NATS-style configs", func() {
		// Without the leading-letter constraint, position 0 would be a digit
		// or special char a sizable fraction of the time. 200 iterations make
		// a regression virtually certain to be caught here.
		for i := 0; i < 200; i++ {
			password, err := generator.GeneratePassword(32)
			Expect(err).NotTo(HaveOccurred())
			Expect(password).To(MatchRegexp(`^[a-zA-Z]`))
		}
	})

	It("should generate unique passwords across 100 invocations", func() {
		passwords := make(map[string]bool)
		for i := 0; i < 100; i++ {
			p, err := generator.GeneratePassword(32)
			Expect(err).NotTo(HaveOccurred())
			Expect(passwords).NotTo(HaveKey(p), "duplicate password generated")
			passwords[p] = true
		}
	})
})

var _ = Describe("GenerateAlphanumeric", func() {
	DescribeTable("should produce an alphanumeric string with the requested length",
		func(length int) {
			result, err := generator.GenerateAlphanumeric(length)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(length))
		},
		Entry("32-character string", 32),
		Entry("short 8-character string", 8),
		Entry("long 64-character string", 64),
	)

	It("should fall back to default length of 32 when given zero", func() {
		value, err := generator.GenerateAlphanumeric(0)
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(HaveLen(32))
	})

	It("should contain only alphanumeric characters (a-z, A-Z, 0-9)", func() {
		result, err := generator.GenerateAlphanumeric(100)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(MatchRegexp(`^[a-zA-Z0-9]+$`))
	})

	It("should always start with a letter so it is safe to embed unquoted in NATS config", func() {
		// The bug is probabilistic: without the constraint, ~16% (10/62) of draws
		// start with a digit. Running 200 iterations makes a regression virtually
		// certain to be caught here.
		for i := 0; i < 200; i++ {
			result, err := generator.GenerateAlphanumeric(32)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(MatchRegexp(`^[a-zA-Z][a-zA-Z0-9]*$`))
		}
	})
})

var _ = Describe("GenerateHex", func() {
	DescribeTable("should produce a hex string with the exact requested length",
		func(length int) {
			result, err := generator.GenerateHex(length)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(length))
		},
		Entry("even length of 32", 32),
		Entry("odd length of 33 (truncated from extra byte)", 33),
	)

	It("should fall back to default length of 32 when given zero", func() {
		value, err := generator.GenerateHex(0)
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(HaveLen(32))
	})

	It("should contain only lowercase hex characters (a-f, 0-9)", func() {
		result, err := generator.GenerateHex(100)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(MatchRegexp(`^[a-f0-9]+$`))
	})
})

var _ = Describe("GenerateUUID", func() {
	It("should produce a valid UUID v4 string matching xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx", func() {
		for i := 0; i < 20; i++ {
			uuid, err := generator.GenerateUUID()
			Expect(err).NotTo(HaveOccurred())
			Expect(uuid).To(MatchRegexp(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`))
		}
	})

	It("should generate unique UUIDs across 100 invocations", func() {
		uuids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			uuid, err := generator.GenerateUUID()
			Expect(err).NotTo(HaveOccurred())
			Expect(uuids).NotTo(HaveKey(uuid), "duplicate UUID generated")
			uuids[uuid] = true
		}
	})
})

var _ = Describe("GenerateValue", func() {
	It("should return the chart value when chartValue is specified (priority over generate)", func() {
		spec := generator.SecretKeySpec{
			Name:       "test",
			ChartValue: stringPtr("my-chart-password"),
			Type:       "password",
			Length:     24,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("my-chart-password"))
	})

	It("should not use empty chartValue for non-static types (fall through to generate)", func() {
		spec := generator.SecretKeySpec{
			Name:       "test",
			ChartValue: stringPtr(""),
			Type:       "password",
			Length:     16,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(16))
	})

	It("should fall back to password generation for static type (static is handled by KeyResolver)", func() {
		spec := generator.SecretKeySpec{
			Name: "test",
			Type: "static",
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
	})

	It("should generate a password of the specified length for type 'password'", func() {
		spec := generator.SecretKeySpec{
			Name:   "test",
			Type:   "password",
			Length: 16,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(16))
	})

	It("should generate an alphanumeric string of the specified length for type 'alphanumeric'", func() {
		spec := generator.SecretKeySpec{
			Name:   "test",
			Type:   "alphanumeric",
			Length: 24,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(24))
		Expect(result).To(MatchRegexp(`^[a-zA-Z0-9]+$`))
	})

	It("should generate a hex string of the specified length for type 'hex'", func() {
		spec := generator.SecretKeySpec{
			Name:   "test",
			Type:   "hex",
			Length: 32,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(32))
		Expect(result).To(MatchRegexp(`^[a-f0-9]+$`))
	})

	It("should generate a valid UUID for type 'uuid' regardless of length setting", func() {
		spec := generator.SecretKeySpec{
			Name: "test",
			Type: "uuid",
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(MatchRegexp(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`))
	})

	It("should default to password generation when type is empty", func() {
		spec := generator.SecretKeySpec{
			Name:   "test",
			Length: 20,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(20))
	})

	It("should fall back to password generation for an unknown type", func() {
		spec := generator.SecretKeySpec{
			Name:   "test",
			Type:   "completely-unknown-type",
			Length: 20,
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(20))
	})

	It("should use default length of 32 when length is not specified", func() {
		spec := generator.SecretKeySpec{
			Name: "test",
			Type: "password",
		}
		result, err := generator.GenerateValue(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(32))
	})
})
