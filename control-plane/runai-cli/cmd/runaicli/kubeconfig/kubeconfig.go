/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubeconfig

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/k8s"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	runaiCliBypassValidationName  = "bypass-validation"
	runaiCliBypassValidationUsage = "bypass cluster domain validation"

	clusterDomainMismatch = `Cluster domain mismatch
Configured cluster domain: %s
Kubernetes cluster domain: %s`

	clusterDomainMismatchDetails = `The cluster URL in your kubeconfig does not match the cluster URL configured in the Run:ai CLI.
This mismatch may cause unexpected behavior when executing commands.

To resolve:
- Verify your current Kubernetes context:
	> kubectl config current-context
- Update your Run:ai CLI configuration with the correct URL:
	> runai config set --cp-url <CORRECT_URL>
- (Optional) To bypass this validation, use the --bypass-validation flag.`
)

var bypassValidation bool

// Cmd represents the kubeconfig command
var Cmd = &cobra.Command{
	Use:     consts.KubeConfigStr,
	Short:   consts.KubeConfigShortMsg,
	Aliases: []string{"kc"},
	Args:    cobra.NoArgs,
	GroupID: consts.GroupConfigID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		instance := config.GetInstance()
		if bypassValidation {
			instance.GetLogger().Debug("bypassing cluster domain validation")
			return nil
		}

		txn := instance.GetTxn()

		matches, err := k8s.ValidateClusterIdentity(txn.GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToValidateClusterKubeConfigMatch, err)
		}

		if !matches {
			k8sDomain, _ := k8s.GetServerUrl(txn.GetCtx())
			clusterDomain := config.GetInstance().GetConfigValue(config.ClusterDomain)
			mismatchString := fmt.Sprintf(clusterDomainMismatch, clusterDomain, k8sDomain)
			return cmdtxn.CmdErrorPrintRaw(cmd, clusterDomainMismatchDetails, fmt.Errorf("%s", mismatchString))
		}
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func init() {
	Cmd.PersistentFlags().BoolVar(&bypassValidation, runaiCliBypassValidationName, false, runaiCliBypassValidationUsage)
}
