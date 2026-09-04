/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"fmt"
	"strings"
)

const allRecords = "1=1"

func listItemsForQuery(ids []string) string {
	base := "("
	for _, id := range ids {
		base += fmt.Sprintf(" '%s',", id)
	}
	return strings.TrimSuffix(base, ",") + ")"
}
