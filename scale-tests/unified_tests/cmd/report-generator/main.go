/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/output"
	reportsgenerator "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/report-generator"
)

type Config struct {
	TestResultsFile string                      `split_words:"true" required:"true"`
	OutputFile      string                      `split_words:"true" default:"metrics_report.json"`
	Type            reportsgenerator.ReportType `split_words:"true" default:"workload_scale"`
}

var c Config

func main() {
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("Failed to process environment variables: %v", err)
	}

	testResults, err := output.ReadTestResults(c.TestResultsFile)
	if err != nil {
		log.Fatalf("Failed to read test results: %v", err)
	}
	generator, err := reportsgenerator.GetReportGenerator(c.Type)
	if err != nil {
		log.Fatalf("Failed to create report generator: %v", err)
	}
	report, err := generator.GenerateReport(testResults)
	if err != nil {
		log.Fatalf("Failed to generate report: %v", err)
	}

	if err := output.WriteReport(*report, c.OutputFile); err != nil {
		log.Fatalf("Failed to write report: %v", err)
	}
}
