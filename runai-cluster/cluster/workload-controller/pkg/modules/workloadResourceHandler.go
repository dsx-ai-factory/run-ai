/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package modules

import (
	"context"
	"fmt"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

func ClearDanglingResources(ctx context.Context, k8sClient client.Client, danglingList []client.Object) error {
	log := logger.FromContext(ctx)
	var wg sync.WaitGroup

	log.Info("checking for dangling resource started...")
	defer log.Info("checking for dangling resource ended...")

	if len(danglingList) == 0 {
		log.Info("no dangling resource found.")
		return nil
	}

	// Create a channel to receive errors from goroutines
	errCh := make(chan error, len(danglingList))
	wg.Add(len(danglingList))

	for _, dangling := range danglingList {
		go danglingResource(ctx, k8sClient, dangling, &wg, errCh)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) == 0 {
		return nil
	}

	for err := range errCh {
		log.Error(err, "workload dangling failed")
	}
	return fmt.Errorf("workload clear dangling resource failed")
}

func danglingResource(ctx context.Context, k8sClient client.Client, object client.Object, wg *sync.WaitGroup, errCh chan<- error) {
	log := logger.FromContext(ctx)
	defer wg.Done()
	log.Info("found dangling resource - deleting", "resourceName", object.GetName(), "resourceKind", object.GetObjectKind())
	err := k8sClient.Delete(ctx, object)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			errCh <- err
		}
		return
	}
}
