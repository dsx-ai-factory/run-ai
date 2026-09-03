/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"
)

var _ = Describe("ReadApplicationSecret", func() {
	var (
		tempFile *os.File
		err      error
	)

	BeforeEach(func() {
		tempFile, err = os.CreateTemp("/tmp", "test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			return
		}
	})

	It("should read the secret from the file", func() {
		secret := "my-secret"
		err := os.WriteFile(tempFile.Name(), []byte(secret), 0644)
		if err != nil {
			return
		}

		result, err := auth.ReadApplicationSecret(tempFile.Name())
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(secret))
	})

	It("should return an error if the file does not exist", func() {
		_, err := auth.ReadApplicationSecret("/non/existent/file")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("FileBased64TokenReadWriter", func() {
	var (
		tempFile *os.File
		err      error
		fbrw     *auth.FileBased64TokenReadWriter
	)

	BeforeEach(func() {
		tempFile, err = os.CreateTemp("/tmp", "test")
		Expect(err).NotTo(HaveOccurred())
		fbrw = auth.NewFileBased64TokenReadWriter()
	})

	AfterEach(func() {
		os.Remove(tempFile.Name())
	})

	It("should write and read the token to/from the file", func() {
		claims := &auth.CustomClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		validToken, err := token.SignedString([]byte("secret"))
		Expect(err).NotTo(HaveOccurred())

		jsonToken := &oauth2.Token{
			AccessToken:  validToken,
			RefreshToken: validToken,
		}

		err = fbrw.WriteToken(tempFile.Name(), jsonToken)
		Expect(err).NotTo(HaveOccurred())

		readToken, err := fbrw.ReadToken(tempFile.Name())
		Expect(err).NotTo(HaveOccurred())
		Expect(readToken.AccessToken).To(Equal(jsonToken.AccessToken))
		Expect(readToken.RefreshToken).To(Equal(jsonToken.RefreshToken))
	})

	It("should return an error if the file does not exist", func() {
		_, err := fbrw.ReadToken("/non/existent/file")
		Expect(err).To(HaveOccurred())
	})
})
