/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import "strings"

type DiffChecker struct {
	hasLookup     string
	missingLookup string
	lookup        []string
}

func NewDiffChecker(hasLookup, missingLookup string, lookup []string) *DiffChecker {
	dc := DiffChecker{
		hasLookup:     hasLookup,
		missingLookup: missingLookup,
		lookup:        lookup,
	}
	return &dc
}

func (dc *DiffChecker) Validate() bool {
	for _, lookup := range dc.lookup {
		found := dc.lineLookup(lookup)
		if !found {
			return false
		}
	}
	return true
}

func (dc *DiffChecker) lineLookup(lookup string) bool {
	return strings.Contains(dc.hasLookup, lookup) && !strings.Contains(dc.missingLookup, lookup)
}
