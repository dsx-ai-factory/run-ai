/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structs_test

import (
	"math"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"

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

var _ = Describe("Offset/NextToken adapter", func() {
	Describe("OffsetToNextToken", func() {
		It("should convert positive offset to NextToken", func() {
			offset := int32(100)

			token, err := structs.OffsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeNil())
			Expect(token.IsSet()).To(BeTrue())

			timestamp := token.GetTimestamp()
			Expect(timestamp.Unix()).To(Equal(int64(100)))
		})

		It("should handle offset of 0", func() {
			offset := int32(0)

			token, err := structs.OffsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeNil())
			Expect(token.IsSet()).To(BeTrue())

			timestamp := token.GetTimestamp()
			Expect(timestamp.Unix()).To(Equal(int64(0)))
		})

		It("should handle large offsets", func() {
			offset := int32(999999)

			token, err := structs.OffsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeNil())

			timestamp := token.GetTimestamp()
			Expect(timestamp.Unix()).To(Equal(int64(999999)))
		})

		It("should return error for negative offset", func() {
			offset := int32(-10)

			token, err := structs.OffsetToNextToken(offset)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("offset cannot be negative"))
			Expect(token).To(BeNil())
		})

		It("should create UTC timestamp", func() {
			offset := int32(50)

			token, err := structs.OffsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			timestamp := token.GetTimestamp()
			Expect(timestamp.Location()).To(Equal(time.UTC))
		})
	})

	Describe("NextTokenToOffset", func() {
		It("should convert NextToken back to offset", func() {
			originalOffset := int32(100)
			token, _ := structs.OffsetToNextToken(originalOffset)

			offset, err := structs.NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).ToNot(BeNil())
			Expect(*offset).To(Equal(originalOffset))
		})

		It("should handle nil token", func() {
			offset, err := structs.NextTokenToOffset(nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).To(BeNil())
		})

		It("should handle unset token", func() {
			token := &structs.NextToken{}

			offset, err := structs.NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).To(BeNil())
		})

		It("should handle offset of 0", func() {
			token, _ := structs.OffsetToNextToken(0)

			offset, err := structs.NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).ToNot(BeNil())
			Expect(*offset).To(Equal(int32(0)))
		})

		It("should handle large offsets", func() {
			originalOffset := int32(999999)
			token, _ := structs.OffsetToNextToken(originalOffset)

			offset, err := structs.NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).ToNot(BeNil())
			Expect(*offset).To(Equal(originalOffset))
		})

		It("should return error for negative timestamp (invalid token)", func() {
			invalidTime := time.Unix(-100, 0)
			token, _ := structs.NextTokenFromTimestamp(invalidTime)

			offset, err := structs.NextTokenToOffset(token)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("offset cannot be negative"))
			Expect(offset).To(BeNil())
		})

		It("should return error for a timestamp beyond math.MaxInt32 (overflow)", func() {
			// math.MaxInt32 + 1 seconds since epoch, i.e. just past 2038-01-19.
			overflowTime := time.Unix(int64(math.MaxInt32)+1, 0)
			token, err := structs.NextTokenFromTimestamp(overflowTime)
			Expect(err).ToNot(HaveOccurred())

			offset, err := structs.NextTokenToOffset(token)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("offset out of range"))
			Expect(offset).To(BeNil())
		})
	})

	Describe("Round-trip conversion", func() {
		It("should preserve offset through round-trip conversion", func() {
			testOffsets := []int32{0, 1, 50, 100, 1000, 50000, 999999}

			for _, originalOffset := range testOffsets {
				token, err := structs.OffsetToNextToken(originalOffset)
				Expect(err).ToNot(HaveOccurred())

				resultOffset, err := structs.NextTokenToOffset(token)
				Expect(err).ToNot(HaveOccurred())
				Expect(resultOffset).ToNot(BeNil())
				Expect(*resultOffset).To(Equal(originalOffset), "Failed for offset: %d", originalOffset)
			}
		})
	})
})
