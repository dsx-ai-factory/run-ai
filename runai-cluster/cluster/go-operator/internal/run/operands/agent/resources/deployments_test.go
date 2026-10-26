/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/validator"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NMC webhook identities", func() {
	DescribeTable("authorizes configured accounts without trusting names in other namespaces",
		func(account string, extra, expected []string) {
			config := &runaiv1.RunaiConfig{}
			config.SetDefaultsWhereNeeded()
			config.Namespace = "custom-runai"
			config.Spec.Internal.Global.NMC.ServiceAccountName = account
			config.Spec.Internal.Global.NMC.AllowedServiceAccounts = extra
			var allowed []string
			for _, variable := range handleEnvVars(config) {
				if variable.Name == "ALLOWED_SERVICE_ACCOUNTS" {
					allowed = strings.Split(variable.Value, ",")
				}
			}
			Expect(allowed).To(ConsistOf(expected))
			for _, username := range expected {
				Expect(validator.AllowedServiceAccounts(username, allowed)).To(BeTrue(), username)
			}
			for _, username := range []string{
				"system:serviceaccount:tenant:main-account",
				"system:serviceaccount:tenant:extra-account",
				"system:serviceaccount:custom-runai:main-account-copy",
				"system:serviceaccount:custom-runai:unrelated",
				"",
			} {
				Expect(validator.AllowedServiceAccounts(username, allowed)).To(BeFalse(), username)
			}
		},
		Entry("default installation retains UID authorization", "", []string(nil), []string{}),
		Entry("NMC account without an extra allowlist", "main-account", []string(nil),
			[]string{"system:serviceaccount:custom-runai:main-account"}),
		Entry("chart-provided account and configured extras", "main-account",
			[]string{"main-account", "extra-account", "system:serviceaccount:management:ops"},
			[]string{"system:serviceaccount:custom-runai:main-account", "system:serviceaccount:custom-runai:extra-account", "system:serviceaccount:management:ops"}),
		Entry("extras without an NMC account", "", []string{"extra-account", ""},
			[]string{"system:serviceaccount:custom-runai:extra-account"}),
	)
})
