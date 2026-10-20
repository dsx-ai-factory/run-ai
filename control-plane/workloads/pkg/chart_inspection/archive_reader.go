// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
)

const (
	chartMetadataFilename = "Chart.yaml"
	questionsYAMLFilename = "questions.yaml"
	questionsYMLFilename  = "questions.yml"
	chartValuesFilename   = "values.yaml"
	chartNameProperty     = "name"
	chartVersionProperty  = "version"
	tarPaddingBufferBytes = 32 << 10
	maxCompressedBytes    = 32 << 20
	maxExpandedBytes      = 64 << 20
	maxArchiveEntries     = 4096
	maxMetadataBytes      = 1 << 20
)

// ParseArchive reads bounded root metadata from a single-root gzip/tar archive.
// It returns errors for unsafe or malformed archives, but reports missing or
// unusable question declarations in api.AiApplicationChartInspectionResponse so callers can distinguish them.
func ParseArchive(data []byte) (*api.AiApplicationChartInspectionResponse, *api.Error) {
	result, err := parseArchive(data)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "invalid Helm chart archive", err)
	}
	return result, nil
}

func parseArchive(data []byte) (*api.AiApplicationChartInspectionResponse, *api.Error) {
	if len(data) > maxCompressedBytes {
		return nil, api.NewValidationError("chart archive exceeds 32 MiB")
	}
	source := bytes.NewReader(data)
	gz, err := gzip.NewReader(source)
	if err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("read chart gzip: %v", err))
	}
	defer gz.Close()
	gz.Multistream(false)
	limited := &io.LimitedReader{R: gz, N: maxExpandedBytes + 1}
	tarReader := tar.NewReader(limited)
	files := map[string][]byte{}
	root := ""
	for count := 0; ; count++ {
		header, err := tarReader.Next()
		if limited.N <= 0 {
			return nil, api.NewValidationError("expanded chart exceeds 64 MiB")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, api.NewValidationError(fmt.Sprintf("read chart tar: %v", err))
		}
		if count >= maxArchiveEntries {
			return nil, api.NewValidationError("chart archive exceeds 4096 entries")
		}
		name := header.Name
		if strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || path.Clean(name) != strings.TrimSuffix(name, "/") {
			return nil, api.NewValidationError("chart archive contains an unsafe path")
		}
		parts := strings.Split(strings.TrimSuffix(name, "/"), "/")
		for _, part := range parts {
			if part == "" || part == "." || part == ".." || strings.Contains(part, ":") {
				return nil, api.NewValidationError("chart archive contains an unsafe path")
			}
		}
		if root == "" {
			root = parts[0]
		}
		if parts[0] != root {
			return nil, api.NewValidationError("chart archive contains multiple roots")
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return nil, api.NewValidationError("chart archive contains a link or unsupported entry")
		}
		if header.Size < 0 || header.Size > limited.N {
			return nil, api.NewValidationError("expanded chart exceeds 64 MiB")
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if len(parts) != 2 {
			continue
		}
		file := parts[1]
		if file != chartMetadataFilename && file != questionsYAMLFilename && file != questionsYMLFilename && file != chartValuesFilename {
			continue
		}
		if _, exists := files[file]; exists {
			return nil, api.NewValidationError(fmt.Sprintf("duplicate root chart metadata %s", file))
		}
		if header.Size > maxMetadataBytes {
			return nil, api.NewValidationError(fmt.Sprintf("%s exceeds 1 MiB", file))
		}
		body, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, api.NewValidationError(fmt.Sprintf("read %s: %v", file, err))
		}
		files[file] = body
	}
	// Reading through EOF verifies the checksum, including when tar stops at its end marker.
	if err := consumeTarPadding(limited); err != nil {
		return nil, err
	}
	if limited.N <= 0 {
		return nil, api.NewValidationError("expanded chart exceeds 64 MiB")
	}
	if source.Len() != 0 {
		return nil, api.NewValidationError("chart archive contains trailing gzip data")
	}
	if _, yamlExists := files[questionsYAMLFilename]; yamlExists {
		if _, ymlExists := files[questionsYMLFilename]; ymlExists {
			return nil, api.NewValidationError("chart archive contains both root questions.yaml and questions.yml")
		}
	}
	metadata, exists := files[chartMetadataFilename]
	if !exists {
		return nil, api.NewValidationError("chart archive has no root Chart.yaml")
	}
	chart, metadataErr := yamlObject(metadata)
	if metadataErr != nil {
		return nil, api.NewValidationError(fmt.Sprintf("parse Chart.yaml: %s", metadataErr.Message))
	}
	name, nameOK := chart[chartNameProperty].(string)
	version, versionOK := chart[chartVersionProperty].(string)
	if !nameOK || strings.TrimSpace(name) == "" || !versionOK || strings.TrimSpace(version) == "" {
		return nil, api.NewValidationError("chart metadata requires a chart name and version")
	}
	sum := sha256.Sum256(data)
	out := &api.AiApplicationChartInspectionResponse{
		ChartName:       name,
		ChartVersion:    version,
		ArchiveSHA256:   hex.EncodeToString(sum[:]),
		QuestionsStatus: api.AIAPPLICATIONCHARTQUESTIONSSTATUS_ABSENT,
		Groups:          []map[string]any{},
		Questions:       []map[string]any{},
		Warnings:        []api.AiApplicationChartInspectionWarning{},
	}
	defaults := map[string]any{}
	if values, ok := files[chartValuesFilename]; ok {
		defaults, metadataErr = yamlObject(values)
		if metadataErr != nil {
			// Defaults are advisory; nil suppresses checks that require readable defaults.
			defaults = nil
			out.Warnings = append(out.Warnings, api.AiApplicationChartInspectionWarning{Code: "UNREADABLE_DEFAULTS", Message: "values.yaml could not be read for default checks", Path: "$"})
		}
	}
	questionsFile := questionsYAMLFilename
	questions, ok := files[questionsFile]
	if !ok {
		questionsFile = questionsYMLFilename
		questions, ok = files[questionsFile]
	}
	if !ok {
		return out, nil
	}
	out.HasQuestions = true
	out.QuestionsStatus = api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE
	declaration, declarationErr := yamlObject(questions)
	if declarationErr != nil {
		out.Warnings = append(out.Warnings, api.AiApplicationChartInspectionWarning{
			Code:    warningUnusableQuestions,
			Message: fmt.Sprintf("%s must contain valid YAML with an object root: %s", questionsFile, declarationErr.Message),
			Path:    "$",
		})
		return out, nil
	}
	parseDeclaration(declaration, defaults, out)
	for index := range out.Warnings {
		if out.Warnings[index].Code == warningUnusableQuestions {
			out.Warnings[index].Message = fmt.Sprintf("%s: %s", questionsFile, out.Warnings[index].Message)
		}
	}
	return out, nil
}

// Read through the gzip trailer to validate its checksum and reject payloads
// hidden after the tar end marker; legal zero padding is still accepted.
func consumeTarPadding(reader io.Reader) *api.Error {
	buffer := make([]byte, tarPaddingBufferBytes)
	for {
		count, err := reader.Read(buffer)
		for _, value := range buffer[:count] {
			if value != 0 {
				return api.NewValidationError("chart tar contains data after its end marker")
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return api.NewValidationError(fmt.Sprintf("read chart gzip trailer: %v", err))
		}
	}
}
