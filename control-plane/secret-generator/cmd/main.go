/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/secret-generator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/secret-generator/pkg/generator"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func main() {
	logging.Init(nil)
	logging.Infof("Starting secret-generator job")

	cfg := config.Init()

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		logging.Fatalf("failed to load in-cluster config: %v", err)
	}

	k8sClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		logging.Fatalf("failed to create Kubernetes client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logging.Warnf("Received signal %v, initiating shutdown", sig)
		cancel()
	}()

	gen := generator.NewSecretGenerator(k8sClient, cfg)
	if err := gen.Run(ctx); err != nil {
		logging.Errorf("Secret generation failed: %v", err)
		os.Exit(1)
	}

	logging.Infof("Secret-generator job completed successfully")
}
