/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package thirdparty

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Image struct {
	Name       string
	Repository string
	Tag        string
}

func FromChart(chart string) ([]Image, error) {
	path := filepath.Join(chart, "values.yaml")
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	images, err := fromValues(contents)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return images, nil
}

func fromValues(contents []byte) ([]Image, error) {
	var values map[string]any
	if err := yaml.Unmarshal(contents, &values); err != nil {
		return nil, err
	}
	entries, ok := values["thirdPartyImages"].(map[string]any)
	if !ok || len(entries) == 0 {
		return nil, fmt.Errorf("thirdPartyImages must be a non-empty map")
	}
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	images := make([]Image, 0, len(names))
	for _, name := range names {
		node := entries[name]
		if path, ok := node.(string); ok {
			var err error
			node, err = valueAtPath(values, path)
			if err != nil {
				return nil, fmt.Errorf("thirdPartyImages.%s: %w", name, err)
			}
		}
		image, ok := node.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("thirdPartyImages.%s must be an image or values path", name)
		}
		repository, err := imageField(image, "repository")
		if err != nil {
			return nil, fmt.Errorf("thirdPartyImages.%s: %w", name, err)
		}
		tag, err := imageField(image, "tag")
		if err != nil {
			return nil, fmt.Errorf("thirdPartyImages.%s: %w", name, err)
		}
		if image["registry"] != nil {
			registry, err := imageField(image, "registry")
			if err != nil {
				return nil, fmt.Errorf("thirdPartyImages.%s: %w", name, err)
			}
			repository = registry + "/" + repository
		}
		images = append(images, Image{Name: name, Repository: repository, Tag: tag})
	}
	return images, nil
}

func valueAtPath(values map[string]any, path string) (any, error) {
	var value any = values
	for _, key := range strings.Split(path, ".") {
		node, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("values path %q is invalid", path)
		}
		value, ok = node[key]
		if !ok {
			return nil, fmt.Errorf("values path %q does not exist", path)
		}
	}
	return value, nil
}

func imageField(image map[string]any, field string) (string, error) {
	value := fmt.Sprint(image[field])
	if value == "" || value == "<nil>" {
		return "", fmt.Errorf("%s is required", field)
	}
	return value, nil
}

func (image Image) Source() string {
	return image.Repository + ":" + image.Tag
}
