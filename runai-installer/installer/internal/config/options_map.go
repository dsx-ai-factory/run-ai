/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"strings"
)

type MapFlags map[string]string

func (m *MapFlags) String() string {
	return fmt.Sprintf("%v", *m)
}

func (m *MapFlags) Type() string { return "MapFlags" }

func (m *MapFlags) Set(value string) error {
	parts := strings.Split(value, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if strings.Contains(p, "=") {
			kv := strings.SplitN(p, "=", 2)
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			(*m)[key] = val
		} else {
			// boolean-like flag (presence indicates enabled); store empty string as value
			key := p
			(*m)[key] = ""
		}
	}
	return nil
}
