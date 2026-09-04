/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integration_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/inference_workloads"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("native vs anyworkload composer parity for git-sync on Knative", func() {
	BeforeEach(func() {
		for _, name := range gitSyncEnvVars {
			GinkgoT().Setenv(name, "")
		}
		GinkgoT().Setenv("GIT_SYNC_IMAGE", gitSyncParityImage)
	})

	DescribeTable("init containers built for the Knative service must match what the cluster builds from StorageFields",
		func(fixtureName string) {
			ctx := context.Background()
			spec := loadInferenceSpec(fixtureName)
			options := buildSharedOptions(ctx)

			clusterInit := initKnativeFromNativePath(ctx, spec, options, fixtureName)
			anyworkloadInit := initContainersFromAnyworkloadKnativePath(ctx, spec, options, fixtureName)

			Expect(anyworkloadInit).To(Equal(clusterInit),
				"anyworkload InferenceToKnativeComposer's init containers diverged from what the cluster "+
					"would build via createGitSyncInitContainer for the same SupersetSpec",
			)
		},
		Entry("git-only, default branch", "git"),
		Entry("git with explicit branch", "git_with_branch"),
		Entry("git with password auth", "git_with_password"),
		Entry("git with SSH key auth", "git_with_ssh_key"),
	)
})

func initContainersFromAnyworkloadKnativePath(
	ctx context.Context,
	spec workloadsApi.InferenceSpecSpec,
	options composer.FieldsComposerOptions,
	workloadName string,
) []corev1.Container {
	GinkgoHelper()
	infComposer := inference_workloads.NewInferenceToKnativeComposer(
		workloadName,
		"team-a",
		"runai-team-a",
		"tester",
		nil,
		options,
	).WithClusterStorageConfig(gitSyncParityStorageConfig())

	target := inference_workloads.NewMinimalKnativeServiceTarget()
	Expect(infComposer.Compose(ctx, convutils.Ref(spec), target)()).To(BeNil(),
		"anyworkload Knative composer returned an error for fixture %s", workloadName)

	return target.Service.Spec.Template.Spec.PodSpec.InitContainers
}
