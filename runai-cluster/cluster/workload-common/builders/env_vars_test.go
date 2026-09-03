/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var value = "ls -la /${RUNAI_PROJECT}/${RUNAI_JOB_NAME} | grep ${RUNAI_UID} | grep ${RUNAI_GID} | grep ${RUNAI_GID}"

func TestNoValuesForVar(t *testing.T) {
	result := ConstructEnvVar(context.Background(), "COMMAND", &value, nil, nil, "", "")
	assert.Equal(t, result.Value, value)
}

func TestPatchPartialValues(t *testing.T) {
	ctx := workloadproject.SetProjectInContext(context.Background(), "team-a")
	result := ConstructEnvVar(ctx, "COMMAND", &value, &field_groups.ContainerFields{
		RunAsGid: field_types.NewOptionalIntField(ptr.To(20)),
	}, nil, "", "")
	assert.Equal(t, result.Value, "ls -la /team-a/${RUNAI_JOB_NAME} | grep ${RUNAI_UID} | grep 20 | grep 20")
}

func TestPatchAllValues(t *testing.T) {
	ctx := workloadproject.SetProjectInContext(context.Background(), "team-a")
	result := ConstructEnvVar(ctx, "COMMAND", &value, &field_groups.ContainerFields{
		RunAsGid: field_types.NewOptionalIntField(ptr.To(20)),
		RunAsUid: field_types.NewOptionalIntField(ptr.To(501)),
	},
		&v1.ObjectMeta{
			Name: "job1",
		}, "", "")
	assert.Equal(t, result.Value, "ls -la /team-a/job1 | grep 501 | grep 20 | grep 20")
}

var _ = Describe("EnvVars tests", func() {
	Context("ConfigMap", func() {
		It("CONFIGMAP:xxx,yyy", func() {
			value, err := getEnvVarSource("CONFIGMAP:xxx,yyy", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.ConfigMapKeyRef).ToNot(BeNil())
			Expect(value.ConfigMapKeyRef.Key).To(Equal("yyy"))
			Expect(value.ConfigMapKeyRef.Name).To(Equal("xxx"))
		})
		It("CONFIGMAP", func() {
			value, err := getEnvVarSource("CONFIGMAP", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(BeNil())
		})
		It("CONFIGMAP:", func() {
			value, err := getEnvVarSource("CONFIGMAP:", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
		It("CONFIGMAP:aa", func() {
			value, err := getEnvVarSource("CONFIGMAP:aa", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("key value is empty"))
			Expect(value).To(BeNil())
		})
		It("CONFIGMAP:,bb", func() {
			value, err := getEnvVarSource("CONFIGMAP:,bb", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
		It("CONFIGMAP:,", func() {
			value, err := getEnvVarSource("CONFIGMAP:,", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
	})

	Context("Secret", func() {
		It("SECRET:xxx,yyy", func() {
			value, err := getEnvVarSource("SECRET:xxx,yyy", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.SecretKeyRef).ToNot(BeNil())
			Expect(value.SecretKeyRef.Key).To(Equal("yyy"))
			Expect(value.SecretKeyRef.Name).To(Equal("xxx"))
		})
		It("SECRET", func() {
			value, err := getEnvVarSource("SECRET", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(BeNil())
		})
		It("SECRET:", func() {
			value, err := getEnvVarSource("SECRET:", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
		It("SECRET:aa", func() {
			value, err := getEnvVarSource("SECRET:aa", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.SecretKeyRef).ToNot(BeNil())
			Expect(value.SecretKeyRef.Key).To(Equal(""))
			Expect(value.SecretKeyRef.Name).To(Equal("aa"))
		})
		It("SECRET:,bb", func() {
			value, err := getEnvVarSource("SECRET:,bb", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
		It("SECRET:,", func() {
			value, err := getEnvVarSource("SECRET:,", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
	})

	Context("PodFieldRef", func() {
		It("PODFIELDREF:xxx,yyy", func() {
			value, err := getEnvVarSource("PODFIELDREF:xxx,yyy", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.FieldRef).ToNot(BeNil())
			Expect(value.FieldRef.FieldPath).To(Equal("xxx"))
		})
		It("PODFIELDREF", func() {
			value, err := getEnvVarSource("PODFIELDREF", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(BeNil())
		})
		It("PODFIELDREF:", func() {
			value, err := getEnvVarSource("PODFIELDREF:", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path value is empty"))
			Expect(value).To(BeNil())
		})
		It("PODFIELDREF:aa", func() {
			value, err := getEnvVarSource("PODFIELDREF:aa", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.FieldRef).ToNot(BeNil())
			Expect(value.FieldRef.FieldPath).To(Equal("aa"))
		})
		It("PODFIELDREF:,bb", func() {
			value, err := getEnvVarSource("PODFIELDREF:,bb", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path value is empty"))
			Expect(value).To(BeNil())
		})
		It("PODFIELDREF:,", func() {
			value, err := getEnvVarSource("PODFIELDREF:,", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path value is empty"))
			Expect(value).To(BeNil())
		})
	})

	Context("User Credential", func() {
		It("USERCREDENTIAL:xxx,yyy", func() {
			value, err := getEnvVarSource("USERCREDENTIAL:xxx,yyy", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.SecretKeyRef).ToNot(BeNil())
			Expect(value.SecretKeyRef.Key).To(Equal("yyy"))
			Expect(value.SecretKeyRef.Name).To(Equal("iw-test-1-xxx"))
		})
		It("USERCREDENTIAL", func() {
			value, err := getEnvVarSource("USERCREDENTIAL", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(BeNil())
		})
		It("USERCREDENTIAL:", func() {
			value, err := getEnvVarSource("USERCREDENTIAL:", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
		It("USERCREDENTIAL:aa", func() {
			value, err := getEnvVarSource("USERCREDENTIAL:aa", "iw", "test-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(value).ToNot(BeNil())
			Expect(value.SecretKeyRef).ToNot(BeNil())
			Expect(value.SecretKeyRef.Key).To(Equal(""))
			Expect(value.SecretKeyRef.Name).To(Equal("iw-test-1-aa"))
		})
		It("USERCREDENTIAL:,bb", func() {
			value, err := getEnvVarSource("USERCREDENTIAL:,bb", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
		It("USERCREDENTIAL:,", func() {
			value, err := getEnvVarSource("USERCREDENTIAL:,", "iw", "test-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("name value is empty"))
			Expect(value).To(BeNil())
		})
	})

})
