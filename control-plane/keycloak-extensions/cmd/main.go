/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/keycloak-extensions/internal/keycloakinit"
)

func main() {
	if err := keycloakinit.New().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "keycloak-extensions-init: %v\n", err)
		os.Exit(1)
	}
}
