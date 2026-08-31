/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package s3

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name      = "name"
	Bucket    = "bucket"
	Path      = "path"
	AccessKey = "accesskey"
	Key       = "key"
	Secret    = "secret"
	Url       = "url"

	ParserName = "S3Parser"
)

var (
	S3 = []string{Name, Bucket, Path, AccessKey, Key, Secret, Url}
)

// name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](S3...)))
}
