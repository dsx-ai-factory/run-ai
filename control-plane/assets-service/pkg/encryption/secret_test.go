/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package encryption

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("validate storing keys in secrets", func() {
	Context("malformed key", func() {
		It("should fail", func() {
			secret := corev1.Secret{
				Data: map[string][]byte{
					"xxx-yyy": nil,
				},
			}
			_, err := parseSecretData(&secret)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("malformed secret key: xxx-yyy"))
		})
	})

	Context("non numeric id in key", func() {
		It("should fail", func() {
			secret := corev1.Secret{
				Data: map[string][]byte{
					"key-yyy-zzz": nil,
				},
			}
			_, err := parseSecretData(&secret)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("non numeric id in secret data key: key-yyy-zzz"))
		})
	})

	Context("unsupported secret key", func() {
		It("should fail", func() {
			secret := corev1.Secret{
				Data: map[string][]byte{
					"key-3-zzz": nil,
				},
			}
			_, err := parseSecretData(&secret)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("unsupported secret key: key-3-zzz"))
		})
	})

	Context("invalid date timestamp", func() {
		It("should fail", func() {
			secret := corev1.Secret{
				Data: map[string][]byte{
					"key-3-created": []byte("xxx"),
				},
			}
			_, err := parseSecretData(&secret)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("invalid secret creation key: xxx"))
		})
	})
})
