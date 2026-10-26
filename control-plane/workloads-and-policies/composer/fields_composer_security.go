/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"net/http"
	"regexp"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/google/uuid"
)

const (
	// as this is only prefix, we ask the linter to ignore it
	// nolint:gosec
	UserCredentialsPrefix = "USERCREDENTIAL"

	NgcApiKeyGenericSecretPrefix  = "ngcgs"
	NgcApiKeyDockerRegistryPrefix = "ngcdr"
)

type UidGid struct {
	Uid *int64
	Gid *int64
}

func (fc *FieldsComposer) ComposeImagePullSecrets(assetsMapping []workloadsApi.AssetsMapping, src []workloadsApi.ImagePullSecret,
	target any, setter FnArraySetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		var assetsMap map[string]workloadsApi.AssetsMapping
		if len(assetsMapping) != 0 {
			assetsMap = CreateAssetMappingMap(assetsMapping)
		}

		secretsArr := make([]string, 0, len(src))

		for _, secret := range src {
			if convutils.BoolIsTrue(secret.Exclude.Get()) {
				continue
			}

			if convutils.IsEmptyOrNull(secret.Name) {
				return workloadsApi.NewError(http.StatusBadRequest, "name must be provided for all instances of imagePullSecret")
			}

			if convutils.BoolIsTrue(secret.UserCredential.Get()) {
				if !fc.Options.UserCredentialsSupported {
					return workloadsApi.NewError(http.StatusBadRequest, "imagePullSecrets: userCredential is not supported")
				}

				// need to handle the translation of the credential name to secret name
				// in case the mapping is known (this happens in the agent). when agent is unknown
				// we (control plane) we just keep it with this USERCREDENTIAL:xxx form.
				if len(assetsMap) != 0 {
					assetMapping, exists := (assetsMap)[*secret.Name]
					if !exists {
						return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("asset mapping not found for credential name: %s (env var: %s)", *secret.Name, *secret.Name))
					}
					if convutils.IsEmptyOrNull(assetMapping.Id) {
						return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("id of asset named '%s' not found", convutils.Deref(assetMapping.Name, "")))
					}

					credentialName, buildErr := BuildUserCredentialName(*secret.Name, assetMapping.Id, nil, *assetMapping.Kind, NgcApiKeyDockerRegistryPrefix)
					if buildErr != nil {
						return buildErr
					}

					secretsArr = append(secretsArr, credentialName)
				} else {
					secretsArr = append(secretsArr, fmt.Sprintf("%s:%s", UserCredentialsPrefix, *secret.Name))
				}
			} else {
				secretsArr = append(secretsArr, *secret.Name)
			}
		}

		return setter(secretsArr, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeCapabilities(src []workloadsApi.Capability,
	target any, setter FnMapToMapSetter[workloadsApi.Capability], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		capaMap := make(map[string]*workloadsApi.Capability)

		for idx, capability := range src {
			key := fmt.Sprintf("capability-%d", idx)
			capaMap[key] = &capability
		}

		return setter(capaMap, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeDnsPolicy(hostNetwork *bool, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if convutils.BoolIsTrue(hostNetwork) {
			dnsPolicy := "ClusterFirstWithHostNet"
			return setter(&dnsPolicy, target, cookie)
		}
		return nil
	}
}

func (fc *FieldsComposer) ComposeRunAsNonRoot(src *bool, privileged *bool,
	target any, setter FnBooleanSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if convutils.BoolIsTrue(src) && convutils.BoolIsTrue(privileged) {
			return workloadsApi.NewError(http.StatusBadRequest, "runAsNonRoot can't be true when privileged is true")
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeAllowPrivilegeEscalation(src *bool, privileged *bool,
	target any, setter FnBooleanSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if convutils.BoolIsFalse(src) && convutils.BoolIsTrue(privileged) {
			return workloadsApi.NewError(http.StatusBadRequest, "allowPrivilegeEscalation can't be false when privileged is true")
		}

		return setter(src, target, cookie)
	}
}

var supplementalPattern = "^\\d+(,\\d+)*$"
var supplementalRegex = regexp.MustCompile(supplementalPattern)

func (fc *FieldsComposer) ComposeSupplementalGroups(src *string,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !supplementalRegex.MatchString(*src) {
			return workloadsApi.NewError(http.StatusBadRequest, "malformed supplementalGroups value: expecting comma-separated set of integers")
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeRunAsUser(src *string,
	target any, setter FnBooleanSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		uidGidFromTheImage := convutils.Deref(src, string(workloadsApi.UIDGIDSOURCE_FROM_THE_IMAGE)) == string(workloadsApi.UIDGIDSOURCE_FROM_THE_IMAGE)

		if !uidGidFromTheImage {
			return setter(convutils.Ref(true), target, cookie)
		}

		return nil
	}
}

// ComposeSeccompProfile projects seccompProfileType via setter.
// If src is set it wins; otherwise defaultValue is used when non-nil
// (e.g. RuntimeDefault for distributed). If both are nil the field is left unset.
func (fc *FieldsComposer) ComposeSeccompProfile(src *workloadsApi.SeccompProfileType,
	defaultValue *string, target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		var value *string
		switch {
		case src != nil:
			value = convutils.Ref(string(*src))
		case defaultValue != nil:
			value = defaultValue
		default:
			return nil
		}
		return setter(value, target, cookie)
	}
}

// BuildUserCredentialName constructs the user credential name string in the format:
// USERCREDENTIAL:<name>-<id[:10]>[,<key>]
// key is used only in environment variables, not in imagePullSecrets
func BuildUserCredentialName(name string, id *string, key *string, kind workloadsApi.AssetKind, ngcPrefix string) (string, *workloadsApi.Error) {
	result := UserCredentialsPrefix + ":"
	// User credentials of NGC API Key kind have a special prefix
	if kind == workloadsApi.ASSETKIND_NGC_API_KEY {
		result += ngcPrefix + "-"
	}
	result += name
	if id != nil {
		_, err := uuid.Parse(*id)
		if err != nil {
			return "", workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("invalid asset id %s: %v", *id, err))
		}
		result += "-" + (*id)[:10]
	}
	if key != nil {
		result += "," + *key
	}
	return result, nil
}
