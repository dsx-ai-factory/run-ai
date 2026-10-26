/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package specset

import (
	"errors"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

var ErrEmptyValue = errors.New("value is empty")

const credentialsParserName = "credentials"
const credentialNameCredFmt = parser.PersonalCredentialAttributeName + "=%s," + parser.PersonalCredentialAttributeCredName + "=%s"
const (
	EmptyString  = ""
	EqualSignStr = "="

	DockerRegistryDefaultKey = ".dockerconfigjson"
	NgcApiKeyDefaultKey      = "NGC_API_KEY"
)

var (
	unifiedCredentialParserSet = mapset.NewSet[string](
		parser.PersonalCredentialAttributeType,
		parser.PersonalCredentialAttributeName,
		parser.PersonalCredentialAttributeCredName,
		parser.PersonalCredentialAttributeKey,
	)
	credentialGenericParserSet = mapset.NewSet[string](
		parser.PersonalCredentialAttributeName,
		parser.PersonalCredentialAttributeCredName,
		parser.PersonalCredentialAttributeKey,
	)
	credentialDockerNgcParserSet = mapset.NewSet[string](
		parser.PersonalCredentialAttributeName,
		parser.PersonalCredentialAttributeCredName,
	)
)

const imagePullSecretsParserName = "image-pull-secrets"

var imagePullSecretsParserSet = mapset.NewSet[string]("type", "name")

type imagePullMyCredentialsType string

const (
	imagePullMyCredDockerRegistry imagePullMyCredentialsType = "dockerRegistry"
	imagePullMyCredNgcApiKey      imagePullMyCredentialsType = "ngcApiKey"
)

func imagePullMyCredentialsTypeFromFlag(flag string) (imagePullMyCredentialsType, error) {
	switch t := imagePullMyCredentialsType(strings.TrimSpace(flag)); t {
	case imagePullMyCredDockerRegistry, imagePullMyCredNgcApiKey:
		return t, nil
	default:
		return "", fmt.Errorf(
			"unsupported credential type: %q (allowed: %s, %s)",
			flag, imagePullMyCredDockerRegistry, imagePullMyCredNgcApiKey,
		)
	}
}

func SetNullableString(value string, nullable *openapi.NullableString) {
	if value != EmptyString {
		nullable.Set(ptr.To(value))
	}
}

func SetNullableInt32(value int32, nullableInt32 *openapi.NullableInt32) {
	if value > 0 {
		nullableInt32.Set(ptr.To[int32](value))
	}
}

func SetNullableFloat64(value float64, nullable *openapi.NullableFloat64) {
	if value > 0 {
		nullable.Set(ptr.To(value))
	}
}

func SetNullableBool(value bool, nullable *openapi.NullableBool) {
	if value {
		nullable.Set(ptr.To(value))
	}
}

func SetGenericType[T, K comparable](value string, attributeConverter func(value string) (*T, error),
	nullableSpecAttribute *K, nullableConstractor func(*T) *K) error {
	if value == EmptyString {
		return ErrEmptyValue
	}
	gen, err := attributeConverter(value)
	if err != nil {
		return err
	}
	*nullableSpecAttribute = *nullableConstractor(gen)
	return nil
}

func SplitEqualSepValidateTwoSection(value string) ([]string, error) {
	if !strings.Contains(value, EqualSignStr) {
		return nil, fmt.Errorf("value (%s) does not contain `=` sign", value)
	}
	valueSplitter := strings.SplitN(value, EqualSignStr, 2)
	return valueSplitter, nil
}
func routeUnifiedCredential(credType, envName, credName, key string, genericVars, dockerVars, ngcApiKeyVars *[]string) error {
	base := fmt.Sprintf(credentialNameCredFmt, envName, credName)
	switch credType {
	case string(openapi.ASSETKIND_GENERIC_SECRET):
		if key == "" {
			return fmt.Errorf(parser.ErrorPersonalCredentialKeyRequired, openapi.ASSETKIND_GENERIC_SECRET)
		}
		*genericVars = append(*genericVars, fmt.Sprintf("%s,"+parser.PersonalCredentialAttributeKey+"=%s", base, key))
	case string(openapi.ASSETKIND_DOCKER_REGISTRY):
		if key != "" {
			return fmt.Errorf(parser.ErrorPersonalCredentialKeyDisallowed, openapi.ASSETKIND_DOCKER_REGISTRY)
		}
		*dockerVars = append(*dockerVars, base)
	case string(openapi.ASSETKIND_NGC_API_KEY):
		if key != "" {
			return fmt.Errorf(parser.ErrorPersonalCredentialKeyDisallowed, openapi.ASSETKIND_NGC_API_KEY)
		}
		*ngcApiKeyVars = append(*ngcApiKeyVars, base)
	default:
		return fmt.Errorf(parser.ErrorPersonalCredentialInvalidType, credType, openapi.ASSETKIND_DOCKER_REGISTRY, openapi.ASSETKIND_GENERIC_SECRET, openapi.ASSETKIND_NGC_API_KEY)
	}
	return nil
}

func parseUnifiedCredentialVars(variables []string) (genericVars, dockerVars, ngcApiKeyVars []string, err error) {
	for _, v := range variables {
		am := parser.NewAttributeMapper(v, parser.NewLooseMapParserAllowValueContainEqual(credentialsParserName, unifiedCredentialParserSet))
		if err := am.Parse(); err != nil {
			return nil, nil, nil, err
		}
		if err := parser.CheckRequiredNonEmpty(am, parser.PersonalCredentialAttributeType, parser.PersonalCredentialAttributeName, parser.PersonalCredentialAttributeCredName); err != nil {
			return nil, nil, nil, err
		}
		parsed := am.GetMap()
		if err := routeUnifiedCredential(
			parsed[parser.PersonalCredentialAttributeType],
			parsed[parser.PersonalCredentialAttributeName],
			parsed[parser.PersonalCredentialAttributeCredName],
			parsed[parser.PersonalCredentialAttributeKey],
			&genericVars, &dockerVars, &ngcApiKeyVars,
		); err != nil {
			return nil, nil, nil, err
		}
	}
	return genericVars, dockerVars, ngcApiKeyVars, nil
}

func SetEnvironmentVariables(
	envVariables, fieldRefVariables, secretVariables, credVars []string,
	specFunc func(v []openapi.EnvironmentVariable),
) error {
	envVar, err := GenerateEnvironmentList(envVariables)
	if err != nil {
		return err
	}
	fieldRefVariablesVar, err := GenerateFieldRefList(fieldRefVariables)
	if err != nil {
		return err
	}
	envVar = append(envVar, fieldRefVariablesVar...)
	secretVariablesVar, err := GenerateSecretEnvList(secretVariables)
	if err != nil {
		return err
	}
	envVar = append(envVar, secretVariablesVar...)
	credGenericVars, credDockerVars, credNgcApiKeyVars, err := parseUnifiedCredentialVars(credVars)
	if err != nil {
		return err
	}
	credEnv, err := generateUserCredentialEnvList(credGenericVars, credDockerVars, credNgcApiKeyVars)
	if err != nil {
		return err
	}
	envVar = append(envVar, credEnv...)
	specFunc(envVar)
	return nil
}

func generateUserCredentialEnvList(credGenericVars, credDockerVars, credNgcApiKeyVars []string) ([]openapi.EnvironmentVariable, error) {
	var envVars []openapi.EnvironmentVariable
	credGenericEnv, err := GenerateUserCredentialGenericEnvList(credGenericVars)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, credGenericEnv...)
	credDockerEnv, err := GenerateUserCredentialEnvList(credDockerVars, DockerRegistryDefaultKey)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, credDockerEnv...)
	credNgcEnv, err := GenerateUserCredentialEnvList(credNgcApiKeyVars, NgcApiKeyDefaultKey)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, credNgcEnv...)
	return envVars, nil
}

func GenerateEnvironmentList(variables []string) ([]openapi.EnvironmentVariable, error) {
	var envConverter []openapi.EnvironmentVariable
	for _, env := range variables {
		envSplitter, err := SplitEqualSepValidateTwoSection(env)
		if err != nil {
			return nil, err
		}
		envConverter = append(envConverter, openapi.EnvironmentVariable{
			Name:  *openapi.NewNullableString(&envSplitter[0]),
			Value: *openapi.NewNullableString(&envSplitter[1]),
		})
	}
	return envConverter, nil
}

func GenerateFieldRefList(variables []string) ([]openapi.EnvironmentVariable, error) {
	var envConverter []openapi.EnvironmentVariable
	for _, env := range variables {
		envSplitter, err := SplitEqualSepValidateTwoSection(env)
		if err != nil {
			return nil, err
		}
		envConverter = append(envConverter, openapi.EnvironmentVariable{
			Name: *openapi.NewNullableString(&envSplitter[0]),
			PodFieldRef: *openapi.NewNullableEnvironmentVariablePodFieldReference(
				&openapi.EnvironmentVariablePodFieldReference{Path: *openapi.NewNullableString(&envSplitter[1])}),
		})
	}
	return envConverter, nil
}

// GenerateSecretEnvList parses ENV_VAR=secret-name,key=secret-key format.
func GenerateSecretEnvList(variables []string) ([]openapi.EnvironmentVariable, error) {
	var envConverter []openapi.EnvironmentVariable
	for _, env := range variables {
		envSplitter, err := SplitEqualSepValidateTwoSection(env)
		if err != nil {
			return nil, err
		}
		envName := envSplitter[0]
		secretName, secretKey, err := parseSecretValue(envSplitter[1])
		if err != nil {
			return nil, fmt.Errorf("invalid env-secret format for %s: %w", envName, err)
		}
		envConverter = append(envConverter, openapi.EnvironmentVariable{
			Name: *openapi.NewNullableString(&envName),
			Secret: *openapi.NewNullableEnvironmentVariableSecret(
				&openapi.EnvironmentVariableSecret{
					Name: *openapi.NewNullableString(&secretName),
					Key:  *openapi.NewNullableString(&secretKey),
				}),
		})
	}
	return envConverter, nil
}

func parseSecretValue(value string) (secretName, secretKey string, err error) {
	parts := strings.Split(value, ",")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("expected format: secret-name,key=secret-key, got: %s", value)
	}
	secretName = parts[0]
	if secretName == "" {
		return "", "", fmt.Errorf("secret name cannot be empty")
	}
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "key=") {
			secretKey = strings.TrimPrefix(part, "key=")
			break
		}
	}
	if secretKey == "" {
		return "", "", fmt.Errorf("key is required, expected format: secret-name,key=secret-key")
	}
	return secretName, secretKey, nil
}

func GenerateImagePullSecretsList(variables []string) ([]openapi.ImagePullSecret, error) {
	var out []openapi.ImagePullSecret
	for _, v := range variables {
		am := parser.NewAttributeMapper(v, parser.NewStrictMapParser(imagePullSecretsParserName, imagePullSecretsParserSet))
		if err := am.Parse(); err != nil {
			return nil, err
		}
		parsed := am.GetMap()
		credType := parsed["type"]
		name := parsed["name"]
		if name == "" {
			return nil, fmt.Errorf("required flag(s) not set: name")
		}
		if _, err := imagePullMyCredentialsTypeFromFlag(credType); err != nil {
			return nil, err
		}
		ips := openapi.NewImagePullSecret()
		ips.SetName(name)
		ips.SetUserCredential(true)
		out = append(out, *ips)
	}
	return out, nil
}

func SetImagePullSecrets(variables []string, specFunc func([]openapi.ImagePullSecret)) error {
	secrets, err := GenerateImagePullSecretsList(variables)
	if err != nil {
		return err
	}
	specFunc(secrets)
	return nil
}

func parseCredentialKeyValuePairs(value string, allowedKeys mapset.Set[string]) (map[string]string, error) {
	am := parser.NewAttributeMapper(value, parser.NewStrictMapParser(credentialsParserName, allowedKeys))
	if err := am.Parse(); err != nil {
		return nil, err
	}
	return am.GetMap(), nil
}

func parseCredentialKeyValuePairsAllowValueContainEqual(value string, allowedKeys mapset.Set[string]) (map[string]string, error) {
	am := parser.NewAttributeMapper(value, parser.NewStrictMapParserAllowValueContainEqual(credentialsParserName, allowedKeys))
	if err := am.Parse(); err != nil {
		return nil, err
	}
	return am.GetMap(), nil
}

func userCredentialEnvVar(envName, credName, key string) openapi.EnvironmentVariable {
	return openapi.EnvironmentVariable{
		Name: *openapi.NewNullableString(&envName),
		UserCredential: *openapi.NewNullableEnvironmentVariableUserCredential(
			&openapi.EnvironmentVariableUserCredential{
				Name: *openapi.NewNullableString(&credName),
				Key:  *openapi.NewNullableString(&key),
			}),
	}
}

func buildUserCredentialEnvList(
	variables []string,
	parseFn func(string) (map[string]string, error),
	requiredKeys []string,
	keyFn func(map[string]string) string,
) ([]openapi.EnvironmentVariable, error) {
	var out []openapi.EnvironmentVariable
	for _, v := range variables {
		parsed, err := parseFn(v)
		if err != nil {
			return nil, err
		}
		if err := parser.CheckRequiredNonEmptyMap(parsed, v, requiredKeys...); err != nil {
			return nil, err
		}
		out = append(out, userCredentialEnvVar(
			parsed[parser.PersonalCredentialAttributeName],
			parsed[parser.PersonalCredentialAttributeCredName],
			keyFn(parsed)))
	}
	return out, nil
}

func GenerateUserCredentialGenericEnvList(variables []string) ([]openapi.EnvironmentVariable, error) {
	return buildUserCredentialEnvList(
		variables,
		func(v string) (map[string]string, error) {
			return parseCredentialKeyValuePairsAllowValueContainEqual(v, credentialGenericParserSet)
		},
		[]string{parser.PersonalCredentialAttributeName, parser.PersonalCredentialAttributeCredName, parser.PersonalCredentialAttributeKey},
		func(parsed map[string]string) string { return parsed[parser.PersonalCredentialAttributeKey] },
	)
}

func GenerateUserCredentialEnvList(variables []string, defaultKey string) ([]openapi.EnvironmentVariable, error) {
	return buildUserCredentialEnvList(
		variables,
		func(v string) (map[string]string, error) {
			return parseCredentialKeyValuePairs(v, credentialDockerNgcParserSet)
		},
		[]string{parser.PersonalCredentialAttributeName, parser.PersonalCredentialAttributeCredName},
		func(map[string]string) string { return defaultKey },
	)
}
