/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mpi

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateWorkload", func() {
	Context("launcherCreationPolicyCheck", func() {
		Context("valid", func() {
			It("LauncherCreationPolicy is AtStartup", func() {
				mpiField := &field_types.MpiField{
					Values: &field_types.MpiMembers{
						LauncherCreationPolicy: &field_types.StringField{
							Value: ptr.To(string(v2beta1.LauncherCreationPolicyAtStartup)),
						},
					},
				}
				Expect(mpiField).ToNot(BeNil())
				err := launcherCreationPolicyCheck(mpiField)
				Expect(err).ToNot(HaveOccurred())
			})

			It("LauncherCreationPolicy is WaitForWorkersReady", func() {
				mpiField := &field_types.MpiField{
					Values: &field_types.MpiMembers{
						LauncherCreationPolicy: &field_types.StringField{
							Value: ptr.To(string(v2beta1.LauncherCreationPolicyWaitForWorkersReady)),
						},
					},
				}
				Expect(mpiField).ToNot(BeNil())
				err := launcherCreationPolicyCheck(mpiField)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("errors", func() {
			It("MpiJob is nil", func() {
				var mpiField *field_types.MpiField
				Expect(mpiField).To(BeNil())
				err := launcherCreationPolicyCheck(mpiField)
				Expect(err).ToNot(HaveOccurred())
			})

			It("MpiJob is empty struct", func() {
				mpiField := &field_types.MpiField{}
				Expect(mpiField).ToNot(BeNil())
				err := launcherCreationPolicyCheck(mpiField)
				Expect(err).ToNot(HaveOccurred())
			})

			It("LauncherCreationPolicy is nil", func() {
				mpiField := &field_types.MpiField{
					Values: &field_types.MpiMembers{
						LauncherCreationPolicy: &field_types.StringField{
							Value: nil,
						},
					},
				}
				Expect(mpiField).ToNot(BeNil())
				err := launcherCreationPolicyCheck(mpiField)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("LauncherCreationPolicy invalid value \"\""))
			})

			It("LauncherCreationPolicy invalid value", func() {
				mpiField := &field_types.MpiField{
					Values: &field_types.MpiMembers{
						LauncherCreationPolicy: &field_types.StringField{
							Value: ptr.To("bla"),
						},
					},
				}
				Expect(mpiField).ToNot(BeNil())
				err := launcherCreationPolicyCheck(mpiField)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("LauncherCreationPolicy invalid value \"bla\""))
			})
		})
	})
})
