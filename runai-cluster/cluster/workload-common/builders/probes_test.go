/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

func createProbesTestAux(probeMembers *field_types.ProbeMembers, allowHttpGetHostField bool) (*corev1.Container, error) {
	probeSpec := &field_types.ProbeField{
		Values: probeMembers,
	}

	containerFields := &field_groups.ContainerFields{
		ReadinessProbe: probeSpec,
	}

	containerSpec := &corev1.Container{}

	err := CreateProbes(context.TODO(), containerFields, containerSpec, allowHttpGetHostField)

	return containerSpec, err
}

func createProbesTest(probeMembers *field_types.ProbeMembers) (*corev1.Container, error) {
	return createProbesTestAux(probeMembers, true)
}

func createProbesTestInference(probeMembers *field_types.ProbeMembers) (*corev1.Container, error) {
	return createProbesTestAux(probeMembers, false)
}

var _ = Describe("Probes tests", func() {
	Context("valid ReadinessProbe with all fields taken from default", func() {
		It("should return a valid probe", func() {
			probeMembers := &field_types.ProbeMembers{
				HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(8080)),
			}

			containerSpec, err := createProbesTest(probeMembers)
			Expect(err).To(BeNil())

			Expect(containerSpec.ReadinessProbe).ToNot(BeNil())

			Expect(containerSpec.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(0)))
			Expect(containerSpec.ReadinessProbe.TimeoutSeconds).To(Equal(int32(0)))
			Expect(containerSpec.ReadinessProbe.PeriodSeconds).To(Equal(int32(0)))
			Expect(containerSpec.ReadinessProbe.SuccessThreshold).To(Equal(int32(0)))
			Expect(containerSpec.ReadinessProbe.FailureThreshold).To(Equal(int32(0)))

			Expect(containerSpec.ReadinessProbe.HTTPGet).ToNot(BeNil())
			Expect(containerSpec.ReadinessProbe.HTTPGet.Port.IntVal).To(Equal(int32(8080)))
			Expect(containerSpec.ReadinessProbe.HTTPGet.Path).To(Equal(""))
			Expect(containerSpec.ReadinessProbe.HTTPGet.Host).To(Equal(""))
			Expect(string(containerSpec.ReadinessProbe.HTTPGet.Scheme)).To(Equal(""))

			Expect(containerSpec.ReadinessProbe.TCPSocket).To(BeNil())
			Expect(containerSpec.ReadinessProbe.GRPC).To(BeNil())
			Expect(containerSpec.ReadinessProbe.Exec).To(BeNil())

			Expect(containerSpec.StartupProbe).To(BeNil())
			Expect(containerSpec.LivenessProbe).To(BeNil())
		})
	})

	Context("valid ReadinessProbe with all fields set", func() {
		It("should return a valid probe", func() {
			probeMembers := &field_types.ProbeMembers{
				InitialDelaySeconds: field_types.NewOptionalIntField(pointer.Int(5)),
				TimeoutSeconds:      field_types.NewOptionalIntField(pointer.Int(6)),
				PeriodSeconds:       field_types.NewOptionalIntField(pointer.Int(7)),
				SuccessThreshold:    field_types.NewOptionalIntField(pointer.Int(8)),
				FailureThreshold:    field_types.NewOptionalIntField(pointer.Int(9)),

				HttpGetPort:   *field_types.NewOptionalIntField(pointer.Int(8080)),
				HttpGetPath:   field_types.NewStringField("/healthz"),
				HttpGetHost:   field_types.NewStringField("10.0.0.1"),
				HttpGetScheme: field_types.NewStringField("https"),
			}

			containerSpec, err := createProbesTest(probeMembers)
			Expect(err).To(BeNil())

			Expect(containerSpec.ReadinessProbe).ToNot(BeNil())

			Expect(containerSpec.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(5)))
			Expect(containerSpec.ReadinessProbe.TimeoutSeconds).To(Equal(int32(6)))
			Expect(containerSpec.ReadinessProbe.PeriodSeconds).To(Equal(int32(7)))
			Expect(containerSpec.ReadinessProbe.SuccessThreshold).To(Equal(int32(8)))
			Expect(containerSpec.ReadinessProbe.FailureThreshold).To(Equal(int32(9)))

			Expect(containerSpec.ReadinessProbe.HTTPGet).ToNot(BeNil())
			Expect(containerSpec.ReadinessProbe.HTTPGet.Port.IntVal).To(Equal(int32(8080)))
			Expect(containerSpec.ReadinessProbe.HTTPGet.Path).To(Equal("/healthz"))
			Expect(containerSpec.ReadinessProbe.HTTPGet.Host).To(Equal("10.0.0.1"))
			Expect(string(containerSpec.ReadinessProbe.HTTPGet.Scheme)).To(Equal("HTTPS"))
		})
	})

	Context("valid ReadinessProbe with some fields set - for inference workload", func() {
		It("should return a valid probe", func() {
			probeMembers := &field_types.ProbeMembers{
				InitialDelaySeconds: field_types.NewOptionalIntField(pointer.Int(5)),

				HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(8080)),
				HttpGetPath: field_types.NewStringField("/healthz"),
			}

			containerSpec, err := createProbesTestInference(probeMembers)
			Expect(err).To(BeNil())

			Expect(containerSpec.ReadinessProbe).ToNot(BeNil())

			Expect(containerSpec.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(5)))

			Expect(containerSpec.ReadinessProbe.HTTPGet).ToNot(BeNil())
			Expect(containerSpec.ReadinessProbe.HTTPGet.Port.IntVal).To(Equal(int32(8080)))
			Expect(containerSpec.ReadinessProbe.HTTPGet.Path).To(Equal("/healthz"))
		})
	})

	Context("invalid probe", func() {
		Context("no port", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetPort is mandatory"))
			})
		})

		Context("port number out of range", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(90000)),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetPort must be in the range"))
			})
		})

		Context("bad path format (doesn't start with /)", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(8080)),
					HttpGetPath: field_types.NewStringField("http://healthz"),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetPath must start with /"))
			})
		})

		Context("bad path format (includes spaces)", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(8080)),
					HttpGetPath: field_types.NewStringField("/healthz /"),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetPath must not include spaces"))
			})
		})

		Context("bad host format (includes spaces)", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(8080)),
					HttpGetHost: field_types.NewStringField("my.domain .com"),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetHost must not include spaces"))
			})
		})

		Context("host provided for inference workload", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort: *field_types.NewOptionalIntField(pointer.Int(8080)),
					HttpGetHost: field_types.NewStringField("localhost"),
				}

				_, err := createProbesTestInference(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetHost is not allowed for this workload type"))
			})
		})

		Context("bad scheme", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort:   *field_types.NewOptionalIntField(pointer.Int(8080)),
					HttpGetScheme: field_types.NewStringField("ftp"),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetScheme must be either HTTP or HTTPS"))
			})
		})

		Context("invalid InitialDelaySeconds", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort:         *field_types.NewOptionalIntField(pointer.Int(8080)),
					InitialDelaySeconds: field_types.NewOptionalIntField(pointer.Int(-1)),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("initialDelaySeconds must be greater than or equal to 0"))
			})
		})

		Context("invalid TimeoutSeconds", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort:    *field_types.NewOptionalIntField(pointer.Int(8080)),
					TimeoutSeconds: field_types.NewOptionalIntField(pointer.Int(0)),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("timeoutSeconds must be greater than or equal to 1"))
			})
		})

		Context("invalid PeriodSeconds", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort:   *field_types.NewOptionalIntField(pointer.Int(8080)),
					PeriodSeconds: field_types.NewOptionalIntField(pointer.Int(0)),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("periodSeconds must be greater than or equal to 1"))
			})
		})

		Context("invalid SuccessThreshold", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort:      *field_types.NewOptionalIntField(pointer.Int(8080)),
					SuccessThreshold: field_types.NewOptionalIntField(pointer.Int(0)),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("successThreshold must be greater than or equal to 1"))
			})
		})

		Context("invalid FailureThreshold", func() {
			It("should fail with the appropriate error message", func() {
				probeMembers := &field_types.ProbeMembers{
					HttpGetPort:      *field_types.NewOptionalIntField(pointer.Int(8080)),
					FailureThreshold: field_types.NewOptionalIntField(pointer.Int(-1)),
				}

				_, err := createProbesTest(probeMembers)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("failureThreshold must be greater than or equal to 0"))
			})
		})
	})
})

var _ = Describe("CreateTcpReadinessProbe", func() {
	It("should set a TCP readiness probe when exactly one port is provided", func() {
		ports := []corev1.ContainerPort{{ContainerPort: 1234}}
		container := &corev1.Container{}
		err := CreateTcpReadinessProbe(ports, container)
		Expect(err).To(BeNil())
		Expect(container.ReadinessProbe).ToNot(BeNil())
		Expect(container.ReadinessProbe.TCPSocket).ToNot(BeNil())
		Expect(container.ReadinessProbe.TCPSocket.Port.IntVal).To(Equal(int32(1234)))
	})

	It("should return an error if zero ports are provided", func() {
		ports := []corev1.ContainerPort{}
		container := &corev1.Container{}
		err := CreateTcpReadinessProbe(ports, container)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("tcp readiness probe expects exactly one port"))
		Expect(container.ReadinessProbe).To(BeNil())
	})

	It("should return an error if more than one port is provided", func() {
		ports := []corev1.ContainerPort{{ContainerPort: 1234}, {ContainerPort: 5678}}
		container := &corev1.Container{}
		err := CreateTcpReadinessProbe(ports, container)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("tcp readiness probe expects exactly one port"))
		Expect(container.ReadinessProbe).To(BeNil())
	})
})
