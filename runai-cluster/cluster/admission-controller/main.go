/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/server"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	server := server.NewServer()
	err := server.Start(ctx)
	if err != nil {
		log.Printf("Failed to start: %v", err)
	}
}
