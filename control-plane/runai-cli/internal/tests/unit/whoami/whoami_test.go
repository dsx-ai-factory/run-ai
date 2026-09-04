/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package whoami

import (
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Whoami Test")
}

const (
	TestFilePath = "whoami"
	claimsOutput = "User test@run.ai logged in with ID 58bbf46a-1245-4862-a2fd-27a5e52028b7\n"
)

var _ = Describe("Whoami Test", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		clients.MockAuthenticationClaims(mockCtrl)
		clients.MockClusters(mockCtrl)
	})

	BeforeEach(func() {

	})

	Context("Pre whoami command login", Ordered, func() {
		It("whoami should failed", func() {
			output, err := cli.RunCommand("whoami", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(consts.ReLoginMsg))
			Expect(output).To(ContainSubstring(consts.ReLoginMsg))
		})

	})

	Context("with valid token", Ordered, func() {
		It("whoami should run", func() {
			auth.GetManager().Init(fake.NewValidTokenBackend, nil)
			_, err = cli.RunCommand("config set --cp-url="+cli.ClusterOneDomain, envMap)
			if err != nil {
				Fail("failed to set cp-url, error: " + err.Error())
			}
			output, err := cli.RunCommand("whoami", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(claimsOutput))
		})
	})
})
