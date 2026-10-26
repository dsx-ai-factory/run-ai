/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fields composer test", func() {
	Context("copy int", func() {
		It("should copy correctly", func() {
			var src = 5
			var wrongSrc = 6.5
			var target int64

			fc := FieldsComposer{}

			err := fc.CopyInt(&src, &target, intToIntTest, nil)()
			Expect(err).To(BeNil())
			Expect(target).To(Equal(int64(5)))

			err = fc.CopyInt(&wrongSrc, &target, intToIntTest, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("unsupported type *float64 for int field"))
		})
	})

	Context("copy string to map", func() {
		It("should copy correctly", func() {
			var target map[string]string

			expectedMap := make(map[string]string)

			fc := FieldsComposer{}

			err := fc.CopyStringToMap(nil, &target, mapTtoMapTest, "test1")()
			Expect(err).To(BeNil())
			Expect(target).To(BeEmpty())

			err = fc.CopyStringToMap(convutils.Ref("test2val"), &target, mapTtoMapTest, "test2")()
			Expect(err).To(BeNil())
			expectedMap["test2"] = "test2val"
			testutils.Expect(target).ToEqual(expectedMap)

			err = fc.CopyStringToMap(convutils.Ref("test2replaced"), &target, mapTtoMapTest, "test2")()
			Expect(err).To(BeNil())
			expectedMap["test2"] = "test2replaced"
			testutils.Expect(target).ToEqual(expectedMap)

			err = fc.CopyStringToMap(convutils.Ref("test3val"), &target, mapTtoMapTest, "test3")()
			Expect(err).To(BeNil())
			expectedMap["test3"] = "test3val"
			testutils.Expect(target).ToEqual(expectedMap)
		})
	})

	Context("env variables errors", func() {
		It("should report errors correctly", func() {
			fc := FieldsComposer{
				Options: OptionsForPolicy(),
			}

			err := fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:      *workloadsApi.NewNullableString(convutils.Ref("test1")),
					ConfigMap: *workloadsApi.NewNullableEnvironmentVariableConfigMap(&workloadsApi.EnvironmentVariableConfigMap{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required configMap name is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					ConfigMap: *workloadsApi.NewNullableEnvironmentVariableConfigMap(&workloadsApi.EnvironmentVariableConfigMap{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required configMap key is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:   *workloadsApi.NewNullableString(convutils.Ref("test1")),
					Secret: *workloadsApi.NewNullableEnvironmentVariableSecret(&workloadsApi.EnvironmentVariableSecret{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required secret name is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					Secret: *workloadsApi.NewNullableEnvironmentVariableSecret(&workloadsApi.EnvironmentVariableSecret{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required secret key is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:        *workloadsApi.NewNullableString(convutils.Ref("test1")),
					PodFieldRef: *workloadsApi.NewNullableEnvironmentVariablePodFieldReference(&workloadsApi.EnvironmentVariablePodFieldReference{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required filedRef path is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:           *workloadsApi.NewNullableString(convutils.Ref("test1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(&workloadsApi.EnvironmentVariableUserCredential{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required user credential name is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(&workloadsApi.EnvironmentVariableUserCredential{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required user credential key is not provided"))
		})
	})

	Context("env variables cluster support", func() {
		It("should report errors correctly", func() {
			vars := []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					ConfigMap: *workloadsApi.NewNullableEnvironmentVariableConfigMap(&workloadsApi.EnvironmentVariableConfigMap{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
						Key:  *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					Secret: *workloadsApi.NewNullableEnvironmentVariableSecret(&workloadsApi.EnvironmentVariableSecret{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
						Key:  *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					PodFieldRef: *workloadsApi.NewNullableEnvironmentVariablePodFieldReference(&workloadsApi.EnvironmentVariablePodFieldReference{
						Path: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(&workloadsApi.EnvironmentVariableUserCredential{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
						Key:  *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}

			fc := FieldsComposer{
				Options: FieldsComposerOptions{
					UserCredentialsSupported: false,
				},
			}
			err := fc.ComposeEnvironmentVariables(nil, vars, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': userCredential is not supported"))
		})
	})

	Context("annotation container-name bound", func() {
		// The Kubernetes annotation key name segment (after the "/") is
		// capped at 63 chars. Our per-container annotation formats reserve
		// some of those chars for fixed text, leaving a smaller budget for
		// the container name. A K8s container name can legitimately be
		// longer than that budget (up to 63 chars), so the composer must
		// reject overlong names instead of producing an annotation key that
		// k8s will refuse at admission time.

		gpuMemory := "8Gi"
		gpuMemoryRequest := "4Gi"
		gpuPortion := 0.8
		gpuPortionRequest := 0.5

		annotationsSupportedFc := FieldsComposer{
			Options: FieldsComposerOptions{GpuMemoryLimitAnnotationsSupported: true},
		}

		It("derives the correct max container-name length from each annotation format", func() {
			Expect(maxContainerNameLenForAnnotationKey(CudaGpuMemoryLimitAnnotationKeyFormat)).To(Equal(36))
			Expect(maxContainerNameLenForAnnotationKey(KaiGpuMemoryPortionLimitAnnotationKeyFormat)).To(Equal(28))
		})

		It("ComposeMemoryLimitAnnotation accepts a container name at the max length", func() {
			maxLen := maxContainerNameLenForAnnotationKey(CudaGpuMemoryLimitAnnotationKeyFormat)
			containerName := strings.Repeat("a", maxLen)
			annotations := map[string]string{}
			sizing := GpuSizingInfo{MemoryLimit: &gpuMemory, MemoryRequest: &gpuMemoryRequest}

			err := annotationsSupportedFc.ComposeMemoryLimitAnnotation(&sizing, containerName, &annotations, StringToMapString)()
			Expect(err).To(BeNil())
			expectedKey := fmt.Sprintf(CudaGpuMemoryLimitAnnotationKeyFormat, containerName)
			Expect(annotations).To(HaveKeyWithValue(expectedKey, gpuMemory))
		})

		It("ComposeMemoryLimitAnnotation rejects a container name one char over the max", func() {
			maxLen := maxContainerNameLenForAnnotationKey(CudaGpuMemoryLimitAnnotationKeyFormat)
			containerName := strings.Repeat("a", maxLen+1)
			annotations := map[string]string{}
			sizing := GpuSizingInfo{MemoryLimit: &gpuMemory, MemoryRequest: &gpuMemoryRequest}

			err := annotationsSupportedFc.ComposeMemoryLimitAnnotation(&sizing, containerName, &annotations, StringToMapString)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("is too long"))
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("at most %d chars", maxLen)))
			Expect(annotations).To(BeEmpty())
		})

		It("ComposePortionLimitAnnotation accepts a container name at the max length", func() {
			maxLen := maxContainerNameLenForAnnotationKey(KaiGpuMemoryPortionLimitAnnotationKeyFormat)
			containerName := strings.Repeat("a", maxLen)
			annotations := map[string]string{}
			sizing := GpuSizingInfo{PortionLimit: &gpuPortion, PortionRequest: &gpuPortionRequest}

			err := annotationsSupportedFc.ComposePortionLimitAnnotation(&sizing, containerName, &annotations, StringToMapString)()
			Expect(err).To(BeNil())
			expectedKey := fmt.Sprintf(KaiGpuMemoryPortionLimitAnnotationKeyFormat, containerName)
			Expect(annotations).To(HaveKeyWithValue(expectedKey, "0.8"))
		})

		It("ComposePortionLimitAnnotation rejects a container name one char over the max", func() {
			maxLen := maxContainerNameLenForAnnotationKey(KaiGpuMemoryPortionLimitAnnotationKeyFormat)
			containerName := strings.Repeat("a", maxLen+1)
			annotations := map[string]string{}
			sizing := GpuSizingInfo{PortionLimit: &gpuPortion, PortionRequest: &gpuPortionRequest}

			err := annotationsSupportedFc.ComposePortionLimitAnnotation(&sizing, containerName, &annotations, StringToMapString)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("is too long"))
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("at most %d chars", maxLen)))
			Expect(annotations).To(BeEmpty())
		})

	})
})

func intToIntTest(src *int64, target any, cookie any) *workloadsApi.Error {
	targetPtr, ok := target.(*int64)
	Expect(ok).To(BeTrue())
	*targetPtr = *src
	return nil
}

func mapTtoMapTest(src map[string]*string, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	Expect(ok).To(BeTrue())

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	for key, value := range src {
		if value != nil {
			(*targetPtr)[key] = *value
		}
	}

	return nil
}
