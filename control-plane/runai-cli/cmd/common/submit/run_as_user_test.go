/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security Settings Functions", func() {
	BeforeEach(func() {
		// Set up logger
		instance.SetLogger(zap.NewNop())
	})

	Describe("calculateSecurityRunAsUserSettings", func() {
		var (
			tokenString string
			err         error
		)

		BeforeEach(func() {
			claims := &auth.CustomClaims{
				UserID:              "123",
				UserEmail:           "test@run.ai",
				Uid:                 "1000",
				Gid:                 "1000",
				SupplementaryGroups: []int{2000, 3000},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err = token.SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when claims are successfully retrieved from token", func() {
			It("should return claims from token", func() {
				uid, gid, groups := calculateSecurityRunAsUserSettings(tokenString)
				Expect(uid).To(Equal(int64(1000)))
				Expect(gid).To(Equal(int64(1000)))
				Expect(groups).To(Equal([]int{2000, 3000}))
			})
		})

		Context("when getting claims from token fails", func() {
			It("should fall back to OS settings", func() {
				uid, gid, sup := calculateSecurityRunAsUserSettings("invalid token")
				Expect(uid).To(Equal(int64(os.Getuid())))
				Expect(gid).To(Equal(int64(os.Getgid())))

				supGroup, err := os.Getgroups()
				Expect(err).NotTo(HaveOccurred())
				Expect(sup).To(Equal(supGroup))
			})
		})
	})

	Describe("getSecurityRunAsUserSettingsFromJWT", func() {
		Context("when claims are not empty", func() {
			var claims *auth.CustomClaims
			BeforeEach(func() {
				claims = &auth.CustomClaims{
					UserID:              "123",
					UserEmail:           "test@run.ai",
					Uid:                 "1000",
					Gid:                 "1000",
					SupplementaryGroups: []int{2000, 3000},
				}
			})

			It("should return claims from JWT", func() {
				uid, gid, groups := getSecurityRunAsUserSettingsFromJWT(claims)
				Expect(uid).To(Equal(int64(1000)))
				Expect(gid).To(Equal(int64(1000)))
				Expect(groups).To(Equal([]int{2000, 3000}))
			})
		})

		Context("when claims are empty", func() {
			var claims *auth.CustomClaims
			BeforeEach(func() {
				claims = &auth.CustomClaims{
					UserID:    "123",
					UserEmail: "test@run.ai",
				}
			})

			It("should fall back to OS settings", func() {
				uid, gid, _ := getSecurityRunAsUserSettingsFromJWT(claims)
				Expect(uid).To(Equal(int64(os.Getuid())))
				Expect(gid).To(Equal(int64(os.Getgid())))
			})
		})

		Context("when claims are partially empty", func() {
			var claims *auth.CustomClaims
			It("should use only UID", func() {
				claims = &auth.CustomClaims{
					UserID:    "123",
					UserEmail: "test@run.ai",
					Uid:       "1000",
				}
				uid, gid, sup := getSecurityRunAsUserSettingsFromJWT(claims)
				Expect(uid).To(Equal(int64(1000)))
				Expect(gid).To(Equal(int64(0)))
				Expect(sup).To(BeNil())
			})

			It("should use only GID", func() {
				claims = &auth.CustomClaims{
					UserID:    "123",
					UserEmail: "test@run.ai",
					Gid:       "1000",
				}
				uid, gid, sup := getSecurityRunAsUserSettingsFromJWT(claims)
				Expect(uid).To(Equal(int64(0)))
				Expect(gid).To(Equal(int64(1000)))
				Expect(sup).To(BeNil())
			})

			It("should use only SupplementaryGroups", func() {
				claims = &auth.CustomClaims{
					UserID:              "123",
					UserEmail:           "test@run.ai",
					SupplementaryGroups: []int{2000, 3000},
				}
				uid, gid, sup := getSecurityRunAsUserSettingsFromJWT(claims)
				Expect(uid).To(Equal(int64(0)))
				Expect(gid).To(Equal(int64(0)))
				Expect(sup).To(Equal([]int{2000, 3000}))
			})
		})
	})

	Describe("getSecurityRunAsUserSettingsFromOs", func() {
		It("should return current user's UID, GID, and supplementary groups", func() {
			uid, gid, groups := getSecurityRunAsUserSettingsFromOs()
			Expect(uid).To(Equal(int64(os.Getuid())))
			Expect(gid).To(Equal(int64(os.Getgid())))
			osGroups, _ := os.Getgroups()
			Expect(groups).To(Equal(osGroups))
		})
	})
})
