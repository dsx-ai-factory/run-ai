/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package create

import (
	"errors"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

const (
	genericSecretItemHint  = "genericSecret --item must contain key and value. Example: key=NAME,value=VALUE"
	dockerRegistryItemHint = "dockerRegistry --item must contain user, password, and url. Example: user=USER,password=PASS,url=URL"
	ngcApiKeyItemHint      = "ngcApiKey --item must contain ngc-api-key. Example: ngc-api-key=YOUR_KEY"
)

var genericSecretItemSet = mapset.NewSet(ItemKeyGenericSecretKey, ItemKeyGenericSecretVal)

func Create(txn *transaction.Txn, cpUrl, token, typeFlag, name, description string, items []string) error {
	credType, err := list.CredentialTypeFromFlag(typeFlag)
	if err != nil {
		return err
	}
	if credType == "" {
		return fmt.Errorf("--type is required for create")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("SECRET_NAME is required and cannot be empty")
	}

	apiClient, err := client.NewClient(cpUrl, token)
	if err != nil {
		return err
	}
	ctx := txn.GetCtx()

	switch credType {
	case list.GenericSecret:
		pairs, err := ParseGenericSecretItems(items)
		if err != nil {
			return err
		}
		_, err = apiClient.CreateGenericSecret(ctx, name, description, pairs)
		return err
	case list.DockerRegistry:
		user, password, url, err := ParseDockerRegistryItems(items)
		if err != nil {
			return err
		}
		_, err = apiClient.CreateDockerRegistry(ctx, name, description, user, password, url)
		return err
	case list.NgcApiKey:
		key, err := ParseNgcApiKeyItems(items)
		if err != nil {
			return err
		}
		_, err = apiClient.CreateNgcApiKey(ctx, name, description, key)
		return err
	default:
		return fmt.Errorf("unsupported credential type: %s", credType)
	}
}

func wrapItemParseError(err error, hint string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, parser.ErrParseValueIsEmpty) {
		return fmt.Errorf("%s", hint)
	}
	s := err.Error()
	if strings.Contains(s, parser.ErrMsgDictionaryMustContainAllKeys) ||
		strings.Contains(s, parser.ErrMsgDictionaryDoesNotContainKey) ||
		strings.Contains(s, parser.ErrMsgMissingMandatoryFields) {
		return fmt.Errorf("%s", hint)
	}
	return err
}

func ParseGenericSecretItems(items []string) ([]client.KeyValue, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one --item with key and value is required for genericSecret")
	}
	pairs := make([]client.KeyValue, 0, len(items))
	for _, item := range items {
		am := parser.NewAttributeMapper(item, parser.NewStrictMapParser("genericSecret", genericSecretItemSet))
		if err := am.Parse(); err != nil {
			return nil, wrapItemParseError(err, genericSecretItemHint)
		}
		m := am.GetMap()
		if err := parser.CheckRequiredNonEmptyMap(m, "genericSecret --item", ItemKeyGenericSecretKey, ItemKeyGenericSecretVal); err != nil {
			return nil, wrapItemParseError(err, genericSecretItemHint)
		}
		pairs = append(pairs, client.KeyValue{
			Key:   m[ItemKeyGenericSecretKey],
			Value: m[ItemKeyGenericSecretVal],
		})
	}
	return pairs, nil
}

var dockerRegistryItemSet = mapset.NewSet(ItemKeyDockerRegistryUser, ItemKeyDockerRegistryPass, ItemKeyDockerRegistryURL)

func ParseDockerRegistryItems(items []string) (string, string, string, error) {
	if len(items) == 0 {
		return "", "", "", fmt.Errorf("%s", dockerRegistryItemHint)
	}
	am := parser.NewAttributeMapper(strings.Join(items, ","), parser.NewStrictMapParser("dockerRegistry", dockerRegistryItemSet))
	if err := am.Parse(); err != nil {
		return "", "", "", wrapItemParseError(err, dockerRegistryItemHint)
	}
	m := am.GetMap()
	if err := parser.CheckRequiredNonEmptyMap(m, "dockerRegistry --item", ItemKeyDockerRegistryUser, ItemKeyDockerRegistryPass, ItemKeyDockerRegistryURL); err != nil {
		return "", "", "", wrapItemParseError(err, dockerRegistryItemHint)
	}
	return m[ItemKeyDockerRegistryUser],
		m[ItemKeyDockerRegistryPass],
		m[ItemKeyDockerRegistryURL],
		nil
}

var ngcApiKeyItemSet = mapset.NewSet(ItemKeyNgcApiKey)

func ParseNgcApiKeyItems(items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("%s", ngcApiKeyItemHint)
	}
	am := parser.NewAttributeMapper(strings.Join(items, ","), parser.NewStrictMapParser("ngcApiKey", ngcApiKeyItemSet))
	if err := am.Parse(); err != nil {
		return "", wrapItemParseError(err, ngcApiKeyItemHint)
	}
	m := am.GetMap()
	if err := parser.CheckRequiredNonEmptyMap(m, "ngcApiKey --item", ItemKeyNgcApiKey); err != nil {
		return "", wrapItemParseError(err, ngcApiKeyItemHint)
	}
	return m[ItemKeyNgcApiKey], nil
}
