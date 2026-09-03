/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

type CredentialType string

const (
	GenericSecret  CredentialType = "genericSecret"
	DockerRegistry CredentialType = "dockerRegistry"
	NgcApiKey      CredentialType = "ngcApiKey"
)

var validCredentialTypes = []CredentialType{GenericSecret, DockerRegistry, NgcApiKey}

func CredentialTypeFromFlag(flag string) (CredentialType, error) {
	if flag == "" {
		return "", nil
	}
	ct := CredentialType(strings.TrimSpace(flag))
	for _, valid := range validCredentialTypes {
		if ct == valid {
			return ct, nil
		}
	}
	return "", fmt.Errorf("unsupported credential type: %q (allowed: %s, %s, %s)", flag, GenericSecret, DockerRegistry, NgcApiKey)
}

func List(txn *transaction.Txn, cpUrl, token, typeFlag string) (*Output, error) {
	if _, err := CredentialTypeFromFlag(typeFlag); err != nil {
		return nil, err
	}
	apiClient, err := client.NewClient(cpUrl, token)
	if err != nil {
		return nil, err
	}
	return NewService(apiClient).List(txn.GetCtx(), typeFlag)
}
