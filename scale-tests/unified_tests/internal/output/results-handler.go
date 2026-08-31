/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package output

import (
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	"os"
)

func ReadTestResults(testResultsFile string) (*scale.SamplesReport, error) {
	data, err := os.ReadFile(testResultsFile)
	if err != nil {
		return nil, err
	}

	var testResults scale.SamplesReport
	if err := json.Unmarshal(data, &testResults); err != nil {
		return nil, err
	}

	return &testResults, nil
}

func WriteReport(report any, outputFile string) error {
	b, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	if err := os.WriteFile(outputFile, b, 0644); err != nil {
		return fmt.Errorf("failed to write report to file: %w", err)
	}
	return nil
}
