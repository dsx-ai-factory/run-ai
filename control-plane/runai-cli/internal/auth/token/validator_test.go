/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
)

var _ = Describe("Validator", func() {
	var (
		validator *Validator
		claims    *auth.CustomClaims
	)

	BeforeEach(func() {
		validator = NewTokenValidator(0)
	})

	Context("Validate", func() {
		It("returns true for a valid, non-expired token", func() {
			expiresAt := time.Now().Add(1 * time.Hour)
			claims = &auth.CustomClaims{}
			claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())
			valid, err := validator.Validate(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeTrue())
		})

		It("returns false for an expired token", func() {
			expiresAt := time.Now().Add(-1 * time.Hour)
			claims = &auth.CustomClaims{}
			claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())

			valid, err := validator.Validate(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeFalse())
		})

		It("returns false for an empty token", func() {
			valid, err := validator.Validate("")
			Expect(err).To(BeNil())
			Expect(valid).To(BeFalse())
		})

		It("returns error if GetClaims fails", func() {
			valid, err := validator.Validate("some_token")
			Expect(err).To(HaveOccurred())
			Expect(valid).To(BeFalse())
		})

		It("respects expirationDelta when validating token", func() {
			expirationDelta := 30 * time.Second
			validatorWithDelta := NewTokenValidator(expirationDelta)
			// Token expires in 20 seconds from now
			expiresAt := time.Now().Add(20 * time.Second)
			claims = &auth.CustomClaims{}
			claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())

			// With expirationDelta=0, token is valid
			valid, err := validator.Validate(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeTrue())

			// With expirationDelta=30s, token is considered expired (since expiresAt < now+delta)
			valid, err = validatorWithDelta.Validate(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeFalse())
		})
	})
})
