/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
)

func IsItemInList[T any](list []T, item T) bool {
	return indexOfStringInList(list, item) > -1
}

func DeleteFromList[T any](list []T, itemToRemove T) []T {
	indexToRemove := indexOfStringInList(list, itemToRemove)

	if indexToRemove >= len(list) || indexToRemove == -1 {
		return list
	}
	if indexToRemove < len(list)-1 {
		copy(list[indexToRemove:], list[indexToRemove+1:]) // Shift a[i+1:] left one index.
	}

	list = list[:len(list)-1] // Truncate slice.
	return list
}

func indexOfStringInList[T any](list []T, item T) int {
	for i, element := range list {
		if reflect.DeepEqual(element, item) {
			return i
		}
	}
	return -1
}

func DeleteFromNodePoolList(list []v1alpha1.NodePool, nodePoolToRemove v1alpha1.NodePool) (result []v1alpha1.NodePool) {
	for _, nodePool := range list {
		if nodePool.Name != nodePoolToRemove.Name {
			result = append(result, nodePool)
		}
	}
	return result
}

func AppendErrIfNotNil(err error, innerErr error) error {
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}
	return err
}

func ParseLabelInt32(labels map[string]string, labelKey string) (int32, error) {
	valueStr, found := labels[labelKey]
	if !found {
		return -1, fmt.Errorf("label <%s> doesn't exist", labelKey)
	}

	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		log.Error().Msgf("Failed parsing label <%s> value <%s> to int32, err: %s", labelKey, valueStr, err.Error())
		return -1, err
	}
	return int32(value), nil
}

func ParseLabelBool(labels map[string]string, labelKey string) (bool, error) {
	valueStr, found := labels[labelKey]
	if !found {
		return false, fmt.Errorf("label <%s> doesn't exist", labelKey)
	}

	return valueStr == "true", nil
}

// Difference returns the elements in firstSlice that aren't in secondSlice.
func Difference[T comparable](firstSlice, secondSlice []T) []T {
	secondSliceSet := make(map[T]struct{}, len(secondSlice))
	for _, element := range secondSlice {
		secondSliceSet[element] = struct{}{}
	}

	var difference []T
	for _, element := range firstSlice {
		if _, found := secondSliceSet[element]; !found {
			difference = append(difference, element)
		}
	}
	return difference
}
