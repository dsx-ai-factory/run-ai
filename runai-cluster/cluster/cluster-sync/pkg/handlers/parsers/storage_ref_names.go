/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
)

const userCredentialPrefix = "USERCREDENTIAL:" // #nosec G101 -- Parser marker, not a credential.

var spaceSplitter = regexp.MustCompile(`\s+`)

func pvcClaimName(pvc *flds.PvcsItemValue, workloadBaseName string, nameSuffix string, idx int) string {
	claimName := ptr.Deref(pvc.ClaimName, "")
	if claimName != "" {
		// claim name is set, use it
		return claimName
	}

	suffix := fmt.Sprintf("%d", idx) // default suffix is idx
	if nameSuffix != "" {
		// nameSuffix is given, the result suffix should be nameSuffix-idx
		suffix = fmt.Sprintf("%s-%d", nameSuffix, idx)
	}

	claimName = fmt.Sprintf("pvc-%s-%s", workloadBaseName, suffix)

	return claimName
}

func parseImagePullSecrets(imagePullSecretsAsString string, workloadName string) (datastructs.Set[string], error) {
	allNames := datastructs.Set[string]{}

	trimmed := strings.TrimSpace(imagePullSecretsAsString)
	if trimmed == "" {
		return allNames, nil
	}

	secretsAsArray := spaceSplitter.Split(trimmed, -1)

	for _, name := range secretsAsArray {
		// Remove USERCREDENTIAL prefix if present
		if strings.HasPrefix(name, userCredentialPrefix) {
			// Native storage observation historically has no workload-kind prefix.
			name = fmt.Sprintf("-%s-%s", workloadName, strings.TrimPrefix(name, userCredentialPrefix))
		}
		//  validate the secret name using the official k8s function
		errs := validation.IsDNS1123Subdomain(name)
		if len(errs) != 0 {
			return nil, fmt.Errorf("invalid secret name %s: %v", name, errs[0])
		}

		allNames.Add(name)
	}

	return allNames, nil
}
