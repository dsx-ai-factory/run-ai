// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

const (
	// QuestionsAbsent means neither supported root questions filename exists.
	QuestionsAbsent = "absent"
	// QuestionsUsable means the entire declaration passed supported validation.
	QuestionsUsable = "usable"
	// QuestionsUnusable means a declaration exists but no partial form is returned.
	QuestionsUnusable = "unusable"
)

// Warning describes ignored metadata, advisory checks or an unusable declaration.
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

// Result contains normalized form metadata and the digest of the inspected bytes.
// It exposes neither the raw questions document nor the complete chart defaults.
type Result struct {
	ChartName       string           `json:"chartName"`
	ChartVersion    string           `json:"chartVersion"`
	ArchiveSHA256   string           `json:"archiveSHA256"`
	HasQuestions    bool             `json:"hasQuestions"`
	QuestionsStatus string           `json:"questionsStatus"`
	Groups          []map[string]any `json:"groups"`
	Questions       []map[string]any `json:"questions"`
	Warnings        []Warning        `json:"warnings"`
}
