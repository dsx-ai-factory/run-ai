/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/client-go/rest"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runai-crd-upgrader/pkg/upgrader"
)

func main() {
	crdsDir := flag.String("crds-dir", "/crds", "directory containing CRD YAML files")
	establishedTimeout := flag.Duration("established-timeout", 5*time.Minute, "maximum time to wait for all CRDs to become Established")
	pollInterval := flag.Duration("poll-interval", time.Second, "interval between CRD Established checks")
	flag.Parse()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("load in-cluster Kubernetes configuration: %v", err)
	}
	// Raise above client-go's 5 QPS / 10 burst default: this job applies and then
	// polls CRDs, and the default client-side limiter throttles the
	// Established checks (APF protects the server).
	config.QPS = 50
	config.Burst = 100
	config.Timeout = 30 * time.Second

	client, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("create Kubernetes API extensions client: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	u := upgrader.New(
		client.CustomResourceDefinitions(),
		*crdsDir,
		*establishedTimeout,
		*pollInterval,
		log.Default(),
	)
	if err := u.Run(ctx); err != nil {
		log.Fatalf("upgrade CRDs: %v", err)
	}
}
