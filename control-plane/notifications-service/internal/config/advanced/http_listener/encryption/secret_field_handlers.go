/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package encryption

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/util"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/util/reflection"
	"github.com/xyproto/jpath"
	"gorm.io/datatypes"
)

const ShouldEncryptTag = "secret"

type PathHandler func(path *reflection.Path, encrypt EncrypterDecrypterer, toNode *jpath.Node) (any, error)

type SecretFieldHandlerer interface {
	RemoveSecrets(data datatypes.JSON) (map[string]interface{}, error)
	MergeSecrets(fromData datatypes.JSON, toData map[string]interface{}) (map[string]interface{}, error)
	DecryptSecrets(fromData datatypes.JSON) (datatypes.JSON, error)
	EncryptSecrets(fromData datatypes.JSON) (datatypes.JSON, error)
}

type SecretFieldHandler struct {
	secretPaths          []*reflection.Path
	encrypterDecrypterer *EncrypterDecrypterer
}

// NewSecretFieldHandler returns a TypePathExplorer, which searches for ShouldEncryptTag tags
func NewSecretFieldHandler(encrypt EncrypterDecrypterer, data ...interface{}) *SecretFieldHandler {
	pathExplorer := reflection.NewPathExplorer(
		func(field reflect.StructField) bool {
			return field.Tag.Get(ShouldEncryptTag) == "true"
		}, data...)

	return &SecretFieldHandler{
		secretPaths:          pathExplorer.Paths,
		encrypterDecrypterer: &encrypt,
	}
}

func (sfh *SecretFieldHandler) RemoveSecrets(data datatypes.JSON) (map[string]interface{}, error) {
	dataJson, err := jpath.New(data)
	if err != nil {
		return nil, err
	}

	for _, path := range sfh.secretPaths {
		if err = dataJson.DelKey(path.GetFullPath()); err != nil && err != jpath.ErrKeyNotFound {
			return nil, fmt.Errorf("failed to delete path: %s %w", path.GetFullPath(), err)
		}
	}

	return dataJson.Map(map[string]interface{}{}), nil
}

func (sfh *SecretFieldHandler) MergeSecrets(fromData datatypes.JSON, toData map[string]interface{}) (map[string]interface{}, error) {
	fromJson, err := jpath.New(fromData)
	if err != nil {
		return nil, err
	}

	toJson, err := util.MapToJPath(toData)
	if err != nil {
		return nil, err
	}

	for _, path := range sfh.secretPaths {
		mergeSecret(path, fromJson, toJson)
	}

	return toJson.Map(map[string]interface{}{}), nil
}

func (sfh *SecretFieldHandler) EncryptSecrets(data datatypes.JSON) (datatypes.JSON, error) {
	return sfh.handleSecretPaths(data, encryptSecret)
}

func (sfh *SecretFieldHandler) DecryptSecrets(data datatypes.JSON) (datatypes.JSON, error) {
	return sfh.handleSecretPaths(data, decryptSecret)
}

func (sfh *SecretFieldHandler) handleSecretPaths(data datatypes.JSON, ph PathHandler) (datatypes.JSON, error) {
	jData, err := jpath.New(data)
	if err != nil {
		return nil, err
	}

	for _, path := range sfh.secretPaths {
		err = handleSecretPath(path, jData, sfh.encrypterDecrypterer, ph)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(jData.Map(map[string]interface{}{}))
}

func mergeSecret(path *reflection.Path, fromJson *jpath.Node, toJson *jpath.Node) {
	fullPath := path.GetFullPath()
	fromNode := fromJson.GetNode(fullPath)
	if fromNode == nil || fromNode.Interface() == nil {
		log.Debug().Msgf("secret Field not exist %s", path.GetFullPath())
		return
	}

	toNodeParent := toJson.GetNode(path.Parent)
	if toNodeParent == nil || toNodeParent.Interface() == nil {
		return
	}

	toNode := toJson.GetNode(fullPath)
	if toNode == nil || toNode.Interface() == nil {
		toNodeParent.Set(path.Field, fromNode.Interface())
		log.Debug().Msgf("secret Field added from existing data %s", path.GetFullPath())
		return
	}
	log.Debug().Msgf("secret was not changed %s", path.GetFullPath())
}

func handleSecretPath(path *reflection.Path, dataJson *jpath.Node, encrypter *EncrypterDecrypterer, ph PathHandler) error {
	nodeParent := dataJson.GetNode(path.Parent)
	if nodeParent == nil || nodeParent.Interface() == nil {
		return nil
	}

	fullPath := path.GetFullPath()
	node := dataJson.GetNode(fullPath)
	if node != nil && node.Interface() != nil {
		handledData, err := ph(path, *encrypter, node)
		if err != nil {
			return err
		}
		nodeParent.Set(path.Field, handledData)
	}
	return nil
}

func encryptSecret(path *reflection.Path, encrypt EncrypterDecrypterer, toNode *jpath.Node) (any, error) {
	if path.Type == reflect.Struct {
		return encrypt.EncryptStruct(toNode.Interface())
	}

	return encrypt.Encrypt(toNode.Interface().(string))
}

func decryptSecret(path *reflection.Path, encrypt EncrypterDecrypterer, toNode *jpath.Node) (any, error) {
	if path.Type == reflect.Struct {
		return encrypt.DecryptStruct(toNode.Interface().(string))
	}

	return encrypt.Decrypt(toNode.Interface().(string))
}
