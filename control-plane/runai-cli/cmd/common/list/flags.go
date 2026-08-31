/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/spf13/cobra"
)

const (
	YamlStr              = "yaml"
	JsonStr              = "json"
	TableStr             = "table"
	TableNoHeadersStr    = "no-headers"
	RawStr               = "raw"
	ColumnsStr           = "columns"
	SetDefaultColumnsStr = "set-default-columns"
)

func FlagOutputType(cmd *cobra.Command, output *Output) {
	cmd.Flags().BoolVar(&output.ofJson, JsonStr, false, "Output structure JSON")
	cmd.Flags().BoolVar(&output.ofYaml, YamlStr, false, "Output structure YAML")
	cmd.Flags().BoolVar(&output.ofTable, TableStr, false, "Output structure table")
	cmd.Flags().BoolVar(&output.ofTableNoHeaders, TableNoHeadersStr, false, "Output structure table without headers")
	cmd.MarkFlagsMutuallyExclusive(JsonStr, YamlStr, TableStr)
}

func FlagOutputTypeWithColumns(cmd *cobra.Command, output *Output) {
	cmd.Flags().BoolVar(&output.ofRaw, RawStr, false, "Return the unprocessed response from the API call")
	cmd.Flags().StringSliceVar(&output.ofColumns, ColumnsStr, nil, "Specify which columns to display (comma-separated list)")
	cmd.Flags().BoolVar(&output.ofSetDefaultColumns, SetDefaultColumnsStr, false, "[Experimental] Set the columns flag value to the default output")
}
