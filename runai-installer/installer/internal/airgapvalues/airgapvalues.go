/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package airgapvalues

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
)

// Image is one mirrored image with its lock group and source tag.
type Image struct {
	Group        string
	GroupVersion string
	Name         string
	Tag          string
}

const (
	controlPlaneGroup = "runai-control-plane"
	clusterGroup      = "runai-cluster"
	kaiGroup          = "kai-scheduler"
	krmGroup          = "kai-resource-management"
	kartaGroup        = "karta"

	// ControlPlaneValuesFile is the generated control-plane values filename.
	ControlPlaneValuesFile = "control-plane-airgap-values.yaml"
	// ClusterValuesFile is the generated cluster values filename.
	ClusterValuesFile = "cluster-airgap-values.yaml"
)

// ControlPlane builds the values that install the control-plane chart from the mirror.
func ControlPlane(registry string, images []Image, chartValues map[string]any) (map[string]any, error) {
	v := values{}
	v.set("global.installationType", "airgap")
	v.set("global.image.registry", artifacts.RegistryPrefix(registry, controlPlaneGroup))
	v.set("global.image.tag", groupVersion(images, controlPlaneGroup))
	// The cluster chart is installed from the bundle, not downloaded.
	v.set("clusterService.config.clusterInstallation.helmRegistryUrl", "")
	if err := applyThirdParty(v, chartValues, controlPlaneGroup, registry, images); err != nil {
		return nil, err
	}
	return v, nil
}

// Cluster builds the values that install the cluster chart from the mirror.
func Cluster(registry string, images []Image, chartValues map[string]any) (map[string]any, error) {
	v := values{}
	v.set("global.image.registry", artifacts.RegistryPrefix(registry, clusterGroup))
	// KRM and the KAI subchart it bundles each compose their own images, and a
	// top-level global.registry would leak into both, so each gets its own key.
	// KRM deliberately exposes no global.registry, hence image.registry here.
	v.set("kai-resource-management.image.registry", artifacts.RegistryPrefix(registry, krmGroup))
	v.set("kai-resource-management.kai-scheduler.global.registry",
		artifacts.RegistryPrefix(registry, kaiGroup))
	for _, image := range images {
		if image.Group == kartaGroup && image.Name == "operator" {
			setFullReferenceImage(v, "karta.image", artifacts.RegistryPrefix(registry, kartaGroup), image)
		}
	}
	if err := applyThirdParty(v, chartValues, clusterGroup, registry, images); err != nil {
		return nil, err
	}
	return v, nil
}

func applyThirdParty(v values, chartValues map[string]any, group, registry string, images []Image) error {
	entries, ok := chartValues["thirdPartyImages"].(map[string]any)
	if !ok || len(entries) == 0 {
		return fmt.Errorf("%s chart has no thirdPartyImages", group)
	}
	for name, entry := range entries {
		image, found := imageByName(images, group, name)
		if !found {
			continue
		}
		mirror := artifacts.RegistryPrefix(registry, group)
		path, referenced := entry.(string)
		if !referenced {
			if _, inline := entry.(map[string]any); !inline {
				return fmt.Errorf("thirdPartyImages.%s must be an image or values path", name)
			}
			setRegistryImage(v, "thirdPartyImages."+name, mirror, image)
			continue
		}
		node, err := valueAtPath(chartValues, path)
		if err != nil {
			return err
		}
		imageValues, ok := node.(map[string]any)
		if !ok {
			return fmt.Errorf("%s is not an image", path)
		}
		if imageValues["registry"] == nil {
			setFullReferenceImage(v, path, mirror, image)
		} else {
			setRegistryImage(v, path, mirror, image)
		}
		repository := fmt.Sprint(imageValues["repository"])
		if strings.HasPrefix(repository, "bitnami") {
			subchart, _, _ := strings.Cut(path, ".")
			v.set(subchart+".global.security.allowInsecureImages", true)
		}
	}
	return nil
}

func imageByName(images []Image, group, name string) (Image, bool) {
	for _, image := range images {
		if image.Group == group && image.Name == name {
			return image, true
		}
	}
	return Image{}, false
}

func valueAtPath(tree map[string]any, path string) (any, error) {
	var value any = tree
	for _, key := range strings.Split(path, ".") {
		node, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("values path %s is invalid", path)
		}
		value, ok = node[key]
		if !ok {
			return nil, fmt.Errorf("values path %s does not exist", path)
		}
	}
	return value, nil
}

func setRegistryImage(v values, path, registry string, image Image) {
	v.set(path+".registry", registry)
	v.set(path+".repository", image.Name)
	v.set(path+".tag", image.Tag)
}

func setFullReferenceImage(v values, path, registry string, image Image) {
	v.set(path+".repository", registry+"/"+image.Name)
	v.set(path+".tag", image.Tag)
}

func groupVersion(images []Image, group string) string {
	for _, image := range images {
		if image.Group == group {
			return image.GroupVersion
		}
	}
	return ""
}
