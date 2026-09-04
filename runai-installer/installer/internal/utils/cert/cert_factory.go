/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cert

import (
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
)

func GetCertHandler(sslCertPathParam, sslKeyPathParam, hostParam configinputs.InstallInputType, runaiData *model.RunAIInstallData) CertHandler {
	sslCertPath := runaiData.GetInputValue(sslCertPathParam)
	sslKeyPath := runaiData.GetInputValue(sslKeyPathParam)
	if (sslCertPath == "" || sslKeyPath == "") && runaiData.Config.DevMode {
		return NewGenerateSelfSignedCert(runaiData.GetInputValue(hostParam))
	}
	return NewLoadCertFromPath(sslCertPath, sslKeyPath)
}
