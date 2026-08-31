/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dynamic_schema_upgrades

import (
	"fmt"
	imageparser "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/common/image-parser"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type mpsServerUpgrade struct{}

func newMpsServerUpgrader() *mpsServerUpgrade {
	return &mpsServerUpgrade{}
}

func (u *mpsServerUpgrade) shouldPerformUpgrade(resource *unstructured.Unstructured) bool {
	spec, ok := resource.Object["spec"]
	if !ok {
		return false
	}
	specObj, ok := spec.(map[string]interface{})
	if !ok {
		return false
	}
	mpsServerConf, ok := specObj["mps-server"]
	if !ok {
		return false
	}
	mpsServerConfObj, ok := mpsServerConf.(map[string]interface{})
	if !ok {
		return false
	}
	image, ok := mpsServerConfObj["image"]
	if !ok {
		return false
	}

	_, ok = image.(string)
	if !ok {
		return false
	}
	return true
}

func (u *mpsServerUpgrade) extractImageStringAfterValidation(resource *unstructured.Unstructured) string {
	return resource.Object["spec"].(map[string]interface{})["mps-server"].(map[string]interface{})["image"].(string)
}

func (u *mpsServerUpgrade) extractImageType(resource *unstructured.Unstructured) (*common.Image, error) {
	imageString := u.extractImageStringAfterValidation(resource)
	imageObj, err := imageparser.ParseImageString(imageString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mps-server image, err:[%v]", err)
	}
	return imageObj, nil
}

func (u *mpsServerUpgrade) updateRunaiConfig(resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	imageType, err := u.extractImageType(resource)
	if err != nil {
		return nil, err
	}
	spec := resource.Object["spec"].(map[string]interface{})
	mpsServerConf := spec["mps-server"].(map[string]interface{})
	mpsServerConf["image"] = imageType
	return resource, nil
}

func (u *mpsServerUpgrade) upgradeHandlerIdentifier() string {
	return "mps-server-image"
}
