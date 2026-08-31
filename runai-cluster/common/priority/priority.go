/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package priority

import (
	"context"

	v1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultPriority = 50

	InferencePriorityClassName = "inference"
	InferencePriorityValue     = 125

	InteractivePriorityClassName = "build"
	InteractivePriorityValue     = 100

	InteractivePreemptiblePriorityClassName = "interactive-preemptible"
	InteractivePreemptiblePriorityValue     = 75

	TrainPriorityClassName = "train"
	TrainPriorityValue     = 50

	PriorityPreemptionThreshold = 100 // Priority value above which a pod is considered preemptible
)

func GetDefaultPriority(ctx context.Context, readerClient client.Reader) (int32, error) {
	podGroupDefaultPriorityClass, err := GetGlobalDefaultPriorityClass(ctx, readerClient)
	if err != nil {
		return 0, err
	}

	if podGroupDefaultPriorityClass != nil {
		return podGroupDefaultPriorityClass.Value, nil
	}
	return defaultPriority, nil
}

func GetPriorityClass(ctx context.Context, priorityClassName string, readerClient client.Reader) (*v1.PriorityClass, error) {
	priorityClass := &v1.PriorityClass{}
	err := readerClient.Get(ctx, types.NamespacedName{Name: priorityClassName}, priorityClass)
	if err != nil {
		return nil, err
	}
	return priorityClass, nil
}

func GetGlobalDefaultPriorityClass(ctx context.Context, readerClient client.Reader) (*v1.PriorityClass, error) {
	priorityClassList := &v1.PriorityClassList{}
	err := readerClient.List(ctx, priorityClassList)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	for i := range priorityClassList.Items {
		pc := &priorityClassList.Items[i]
		if pc.GlobalDefault {
			return pc, nil
		}
	}
	return nil, nil
}

func IsPreemptible(ctx context.Context, readerClient client.Reader, priorityClassName string) (bool, error) {
	if priorityClassName == "" {
		// TODO: use default preemptible value from config
		return false, nil
	}

	priorityClass, err := GetPriorityClass(ctx, priorityClassName, readerClient)
	if err != nil {
		return false, err
	}

	return priorityClass.Value < PriorityPreemptionThreshold, nil
}
