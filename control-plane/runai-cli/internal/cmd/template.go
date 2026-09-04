/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
)

const (
	WorkloadTypeReplacer      = "{{.WorkloadType}}"
	WorkloadNameReplacer      = "{{.WorkloadName}}"
	FrameworkReplacer         = "{{.Framework}}"
	FollowingReplacer         = "{{.Following}}"
	TypeReplacer              = "{{.Type}}"
	IndefiniteArticleReplacer = "{{.AOrAn}}"
	ExampleReplacer           = "{{.ExampleType}}"
	WorkloadFrameworkReplacer = WorkloadTypeReplacer + " " + FrameworkReplacer
	FrameworkWorkloadReplacer = FrameworkReplacer + " " + WorkloadTypeReplacer
	CommandTypeReplacer       = "{{.CommandType}}"
	FollowingWorkload         = "To track the workload's status, run 'runai {{.WorkloadType}} {{.Framework}} describe {{.WorkloadName}}'\n"
	FollowingFrameworkLess    = "To track the workload's status, run 'runai {{.WorkloadType}} describe {{.WorkloadName}}'\n"
	FollowingNim              = "To track the workload's status, run 'runai inference nim describe {{.WorkloadName}}'\n"

	Mod2                = 2
	IndefiniteArticleAn = "an"
	IndefiniteArticleA  = "a"

	InferencePodDescription            = "If no pod is specified, a pod from the latest ready revision will be used."
	DistributedInferencePodDescription = "If no pod is specified, a leader pod will be used."
	NimPodDescription                  = "If no pod is specified, a ready pod from the workload is used."
)

func GetInferencePodDescription() string {
	return InferencePodDescription
}

func GetDistributedInferencePodDescription() string {
	return DistributedInferencePodDescription
}

func GetNimPodDescription() string {
	return NimPodDescription
}

func GetFrameworkWorkloadReplacementTokens(workloadType api.WorkloadType, framework api.Framework, extra ...TokenPair) []TokenPair {
	aOrAn := IndefiniteArticleA
	if workloadType == api.Inference { // all workloads that should start with "an" indefinite article
		aOrAn = IndefiniteArticleAn
	}

	replacementTokens := []TokenPair{
		{Old: IndefiniteArticleReplacer, New: aOrAn},
		{Old: TypeReplacer, New: FrameworkWorkloadReplacer},
		{Old: CommandTypeReplacer, New: WorkloadFrameworkReplacer},
		{Old: WorkloadTypeReplacer, New: workloadType.CLIType()},
		{Old: FrameworkReplacer, New: framework.CLIType()},
		{Old: ExampleReplacer, New: framework.CLIType()},
		{Old: FollowingReplacer, New: FollowingWorkload},
	}

	return append(replacementTokens, extra...)
}

func GetWorkloadOnlyReplacementTokens(workloadType api.WorkloadType, extra ...TokenPair) []TokenPair {
	aOrAn := IndefiniteArticleA
	if workloadType == api.Inference { // all workloads that should start with "an" indefinite article
		aOrAn = IndefiniteArticleAn
	}

	replacementTokens := []TokenPair{
		{Old: IndefiniteArticleReplacer, New: aOrAn},
		{Old: TypeReplacer, New: workloadType.CLIType()},
		{Old: CommandTypeReplacer, New: workloadType.CLIType()},
		{Old: WorkloadTypeReplacer, New: workloadType.CLIType()},
		{Old: ExampleReplacer, New: workloadType.CLIType()},
		{Old: FollowingReplacer, New: FollowingFrameworkLess},
	}

	return append(replacementTokens, extra...)
}

// GetInferenceReplacementTokens returns replacement tokens for inference workload types
// with corrected CommandType and ExampleType paths.
// For DistributedInference: "inference distributed" and "distributed-inference"
// For Inference: "inference" and "inference"
func GetInferenceReplacementTokens(workloadType api.WorkloadType, extra ...TokenPair) []TokenPair {
	aOrAn := IndefiniteArticleA
	if workloadType == api.Inference {
		aOrAn = IndefiniteArticleAn
	}

	// Override CommandType and ExampleType for inference to get correct path
	commandPath := "inference"
	exampleName := "inference"
	if workloadType == api.DistributedInference {
		commandPath = "inference distributed"
		exampleName = "distributed-inference"
	}

	// Build tokens from scratch to ensure CommandType and ExampleType are correct
	replacementTokens := []TokenPair{
		{Old: IndefiniteArticleReplacer, New: aOrAn},
		{Old: TypeReplacer, New: workloadType.CLIType()},
		{Old: CommandTypeReplacer, New: commandPath},
		{Old: WorkloadTypeReplacer, New: workloadType.CLIType()},
		{Old: ExampleReplacer, New: exampleName},
		{Old: FollowingReplacer, New: FollowingFrameworkLess},
	}

	return append(replacementTokens, extra...)
}

func GetNimReplacementTokens(extra ...TokenPair) []TokenPair {
	return append([]TokenPair{
		{Old: IndefiniteArticleReplacer, New: IndefiniteArticleA},
		{Old: TypeReplacer, New: "nim inference"},
		{Old: CommandTypeReplacer, New: "inference nim"},
		{Old: WorkloadTypeReplacer, New: "nim inference"},
		{Old: ExampleReplacer, New: "nim-inference"},
		{Old: FollowingReplacer, New: FollowingNim},
	}, extra...)
}

// TokenPair represents a replacement token pair (old -> new)
type TokenPair struct {
	Old string
	New string
}

func ToList(pairs []TokenPair) []string {
	list := make([]string, len(pairs)*Mod2)
	for i, pair := range pairs {
		list[Mod2*i] = pair.Old
		list[(Mod2*i)+1] = pair.New
	}
	return list
}

// BaseCommandTexts provides common functionality for command text templating
type BaseCommandTexts struct {
	shortTemplate     string
	longTemplate      string
	exampleTemplate   string
	baseErrorTemplate string
	successTemplate   string
	aliases           []string
	replacementTokens []TokenPair
}

// NewBaseCommandTexts creates a new BaseCommandTexts instance
func NewBaseCommandTexts(shortTemplate string, longTemplate string, exampleTemplate string, baseErrorTemplate string,
	successTemplate string, aliases []string, replacementTokens []TokenPair) *BaseCommandTexts {
	return &BaseCommandTexts{
		shortTemplate:     shortTemplate,
		longTemplate:      longTemplate,
		exampleTemplate:   exampleTemplate,
		baseErrorTemplate: baseErrorTemplate,
		successTemplate:   successTemplate,
		replacementTokens: replacementTokens,
		aliases:           aliases,
	}
}

// replaceTokens applies all token replacements to the given template
func (b *BaseCommandTexts) replaceTokens(template string) string {
	return internalReplace(template, ToList(b.replacementTokens))

}

func internalReplace(template string, tokenList []string) string {
	if template == "" {
		return ""
	}
	var after string
	// Initialize variables for replacement loop
	before := template

	// Keep replacing tokens until no more changes occur
	for {
		// Apply all replacements
		after = strings.NewReplacer(tokenList...).Replace(before)

		// If no changes were made in this iteration, we're done
		if after == before {
			break
		}
		// Update before for next iteration
		before = after
	}

	return after
}

// replaceTokens applies all token replacements to the given template
func (b *BaseCommandTexts) replaceTokensWithWorkload(template, workloadName string) string {
	newTokenList := append(b.replacementTokens, TokenPair{Old: WorkloadNameReplacer, New: workloadName})
	return internalReplace(template, ToList(newTokenList))
}

// GetShort returns the short description with tokens replaced
func (b *BaseCommandTexts) GetShort() string {
	return b.replaceTokens(b.shortTemplate)
}

// GetLong returns the long description with tokens replaced
func (b *BaseCommandTexts) GetLong() string {
	return b.replaceTokens(b.longTemplate)
}

// GetExample returns the example with tokens replaced
func (b *BaseCommandTexts) GetExample() string {
	return b.replaceTokens(b.exampleTemplate)
}

// GetBaseErrorMsg returns the base error message with tokens replaced
func (b *BaseCommandTexts) GetBaseErrorMsg() string {
	return b.replaceTokens(b.baseErrorTemplate)
}

// GetSuccessMsg returns the success message with tokens replaced
func (b *BaseCommandTexts) GetSuccessMsg(workload string) string {
	return b.replaceTokensWithWorkload(b.successTemplate, workload)
}

// GetAliases returns command aliases (empty by default)
func (b *BaseCommandTexts) GetAliases() []string {
	return b.aliases
}

// BaseSearchableCommandTexts extends BaseCommandTexts with searchable functionality
type BaseSearchableCommandTexts struct {
	BaseCommandTexts
	noResultsTemplate  string
	searchableTemplate string
}

// NewBaseSearchableCommandTexts creates a new BaseSearchableCommandTexts instance
func NewBaseSearchableCommandTexts(shortTemplate string, longTemplate string, exampleTemplate string,
	baseErrorTemplate string, successTemplate string, noResultsTemplate string,
	searchableTemplate string, replacementTokens []TokenPair, aliases []string) *BaseSearchableCommandTexts {
	return &BaseSearchableCommandTexts{
		BaseCommandTexts: BaseCommandTexts{
			shortTemplate:     shortTemplate,
			longTemplate:      longTemplate,
			exampleTemplate:   exampleTemplate,
			baseErrorTemplate: baseErrorTemplate,
			successTemplate:   successTemplate,
			replacementTokens: replacementTokens,
			aliases:           aliases,
		},
		noResultsTemplate:  noResultsTemplate,
		searchableTemplate: searchableTemplate,
	}
}

// GetNoResultsErrorMsg returns the no results error message with tokens replaced
func (b *BaseSearchableCommandTexts) GetNoResultsErrorMsg() string {
	return b.replaceTokens(b.noResultsTemplate)
}

// GetSearchableErrorMsg returns the searchable error message with tokens replaced
func (b *BaseSearchableCommandTexts) GetSearchableErrorMsg() string {
	return b.replaceTokens(b.searchableTemplate)
}
