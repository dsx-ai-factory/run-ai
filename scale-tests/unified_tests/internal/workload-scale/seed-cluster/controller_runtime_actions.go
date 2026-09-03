/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package seed_cluster

import (
	"context"
	"fmt"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InitCtrlManager(client client.Client, workloadSampler *WorkloadSampler, projectsPrefix string) (context.Context, error) {
	ctx := ctrl.SetupSignalHandler()
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		return nil, err
	}
	controller := NewK8sController(workloadSampler, projectsPrefix, client)
	err = controller.SetupWithManager(mgr)
	if err != nil {
		return nil, err
	}
	go func() {
		fmt.Println("Starting controller manager...")
		if err := mgr.Start(ctx); err != nil {
			log.Fatalf("Controller manager exited: %v", err)
			os.Exit(1)
		}
	}()
	return ctx, nil
}
