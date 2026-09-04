/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

const (
	limitAttribute        = "limit"
	offsetAttribute       = "offset"
	nextTokenAttribute    = "next-token"
	maxItemsAttribute     = "max-items"
	pageSizeAttribute     = "page-size"
	noPaginationAttribute = "no-pagination"

	limitDeprecationMsg  = "Warning: The `--limit` flag is deprecated. Please use `--page-size` instead."
	offsetDeprecationMsg = "Warning: The `--offset` flag is deprecated. Please use `--continue-token` instead."

	nextTokenUsage    = "set the token for requesting the next page" //nolint:gosec // just text
	maxItemsUsage     = "set the max number of items to return, default is all of the items"
	pageSizeUsage     = "set the single page size (max 500)"
	noPaginationUsage = "return a single page instead of full list, by default set to false"

	defaultFilterPageSize = 10
	defaultSinglePageSize = 1

	ErrNegativeOffset = "--offset must be a non-negative number, got %d"
)

type PageOptions struct {
	Offset       int32
	MaxItems     int32
	PageSize     int32
	NoPagination bool
	NextToken    structs.NextToken
}

func NewFilterPageOption() *PageOptions {
	return &PageOptions{
		PageSize:     defaultFilterPageSize,
		NoPagination: true,
	}
}

func NewSingleArtifactPageOption() *PageOptions {
	return &PageOptions{
		PageSize:     defaultSinglePageSize,
		NoPagination: true,
	}
}

func SetPageFlags(cmd *cobra.Command, options *PageOptions) {
	cmd.Flags().Int32Var(&options.MaxItems, limitAttribute, 0, consts.AttributeLimitUsage)
	cmd.Flags().Int32Var(&options.Offset, offsetAttribute, 0, consts.AttributeOffsetUsage)
	_ = cmd.Flags().MarkDeprecated(limitAttribute, limitDeprecationMsg)
	_ = cmd.Flags().MarkDeprecated(offsetAttribute, offsetDeprecationMsg)

	cmd.Flags().Var(cli.NewNextTokenFlag(&options.NextToken), nextTokenAttribute, nextTokenUsage)
	cmd.Flags().Int32Var(&options.MaxItems, maxItemsAttribute, 0, maxItemsUsage)
	cmd.Flags().Int32Var(&options.PageSize, pageSizeAttribute, 100, pageSizeUsage)
	cmd.Flags().BoolVar(&options.NoPagination, noPaginationAttribute, false, noPaginationUsage)

	cmd.MarkFlagsMutuallyExclusive(offsetAttribute, nextTokenAttribute)
	cmd.MarkFlagsMutuallyExclusive(offsetAttribute, maxItemsAttribute)
	cmd.MarkFlagsMutuallyExclusive(limitAttribute, maxItemsAttribute)
	cmd.MarkFlagsMutuallyExclusive(limitAttribute, nextTokenAttribute)
	cmd.MarkFlagsMutuallyExclusive(pageSizeAttribute, maxItemsAttribute)
}
