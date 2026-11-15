// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	fixtureYMLQuestionsFilename = "questions.yml"
	fixtureChartMetadataYAML    = "name: demo\nversion: 1.0.0\n"
	fixtureExampleChartName     = "example"
	fixtureWorkerNamePath       = "workers[0].name"
	fixtureChartName            = "app"
	fixtureRepositoryURL        = "https://example.com"
	fixtureQuestionsFilename    = "questions.yaml"
	fixtureChartMetadataPath    = "demo/Chart.yaml"
	fixtureUploadID             = "authorized-upload"
	fixtureValuesPath           = "demo/values.yaml"
	fixtureVersion              = "1.2.3"
	fixtureMalformedQuestions   = "questions: [\n"
	fixtureEmptyQuestions       = "questions: []\n"
	fixtureWorkerEnabledPath    = "workers.enabled"
)

func TestChartInspection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chart Inspection Suite")
}
