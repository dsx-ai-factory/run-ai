/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package image_parser

import (
	"errors"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"k8s.io/utils/ptr"
	"regexp"
)

var (
	MultiLineRegexpStr   = fmt.Sprintf(`(?m)^[\n\r\s]+.*:\s*%s`, DockerImageRegExpStr)
	DockerImageRegExpStr = `([^\/|\n|:]*)\/(.*?):([^\s|"]*)$`
	MultiLineRegexp      = regexp.MustCompile(MultiLineRegexpStr)
	DockerImageRegExp    = regexp.MustCompile(DockerImageRegExpStr)
)

var (
	ErrReferenceInvalidFormat = errors.New("invalid reference format")
	ErrNameEmpty              = errors.New("repository name must have at least one component")
)

func ParseMultilineContent(content string) ([]*common.Image, error) {
	matches := MultiLineRegexp.FindAllStringSubmatch(content, -1)
	var images []*common.Image
	for _, match := range matches {
		image, err := extractImageRef(match, true)
		if err != nil {
			continue
		}
		images = append(images, image)
	}
	return images, nil
}

func ParseImageString(image string) (*common.Image, error) {
	matches := DockerImageRegExp.FindStringSubmatch(image)
	if matches == nil {
		if image == "" {
			return nil, ErrNameEmpty
		}
		return nil, ErrReferenceInvalidFormat
	}
	return extractImageRef(matches, true)
}

func extractImageRef(matches []string, setDefaults bool) (*common.Image, error) {
	if matches[1] == "" || matches[2] == "" || matches[3] == "" {
		return nil, ErrReferenceInvalidFormat
	}
	imageObj := &common.Image{}
	imageObj.Registry = ptr.To(matches[1])
	imageObj.Name = ptr.To(matches[2])
	imageObj.Tag = ptr.To(matches[3])
	if setDefaults {
		imageObj.SetDefaultsWhereNeeded(&common.Image{})
	}
	return imageObj, nil
}
