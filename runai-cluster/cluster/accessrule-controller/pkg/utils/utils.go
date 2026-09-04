/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"github.com/rs/zerolog/log"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsItemInList[T any](list []T, item T) bool {
	return indexOfFirstItemInList(list, item) > -1
}

func DeleteFirstFromList[T any](list []T, itemToRemove T) []T {
	indexToRemove := indexOfFirstItemInList(list, itemToRemove)

	if indexToRemove >= len(list) || indexToRemove == -1 {
		return list
	}
	if indexToRemove < len(list)-1 {
		copy(list[indexToRemove:], list[indexToRemove+1:]) // Shift a[i+1:] left one index.
	}

	list = list[:len(list)-1] // Truncate slice.
	return list
}

func indexOfFirstItemInList[T any](list []T, item T) int {
	for i, element := range list {
		if reflect.DeepEqual(element, item) {
			return i
		}
	}
	return -1
}

func GetServiceAccount(ctx context.Context, k8sClient client.Client, name, namespace string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{}
	serviceAccountKey := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting ServiceAccount <%s/%s>, error: %s",
			namespace, name, err.Error())
		return nil, err
	}

	return serviceAccount, nil
}
