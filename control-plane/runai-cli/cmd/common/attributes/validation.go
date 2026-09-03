/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attributes

import (
	"errors"
)

const (
	MaxExclusiveCounter       = 1
	ErrorMoreThanMaxExclusive = "more than 1 of the attributes are set"
	ErrorNoneAttributes       = "none of the attributes are set"
)

func MutuallyExclusiveAtLeastOne(attributes ...interface{}) error {
	counter := getAttributesWithValues(attributes)
	if counter == 0 {
		return errors.New(ErrorNoneAttributes)
	}
	if counter > MaxExclusiveCounter {
		return errors.New(ErrorMoreThanMaxExclusive)
	}
	return nil
}

func getAttributesWithValues(attributes ...interface{}) int {
	var counter int
	for _, inters := range attributes {
		if inter, ok := inters.([]interface{}); ok {
			for _, att := range inter {
				switch attType := att.(type) {
				case bool:
					if attType {
						counter++
					}
				case string:
					if attType != "" {
						counter++
					}
				}
			}
		}
	}
	return counter
}
