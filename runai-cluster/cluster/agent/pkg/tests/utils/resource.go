/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing_utils

import (
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"    //nolint:staticcheck
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NamespacedNameFor(resource client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}
}

func DeleteAndPollUntilDeleted(k8sClient client.Client, resource client.Object, intervals ...time.Duration) {
	deleteAndPollUntilDeletedInner(k8sClient, resource, false, intervals...)
}

func deleteAndPollUntilDeletedInner(k8sClient client.Client, resource client.Object, immediately bool, intervals ...time.Duration) {
	By(fmt.Sprintf("waiting for %s to be deleted...", resourceUniqueID(resource)))

	// deletedResourceUID := resource.GetUID()

	deleteOptions := &client.DeleteOptions{}
	var gracePeriodSeconds int64 = 0
	if immediately {
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	deleteFn := func() error {
		err := k8sClient.Delete(globalAPIContext, resource, deleteOptions)
		if err == nil || errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	ExpectWithOffset(1, deleteFn()).To(Succeed())

	actual := func() bool {
		err := k8sClient.Get(globalAPIContext, NamespacedNameFor(resource), resource)
		// return errors.IsNotFound(err) || resource.GetUID() != deletedResourceUID
		return errors.IsNotFound(err)
	}

	if len(intervals) == 0 {
		EventuallyWithOffset(1, actual, TimeoutUnit).Should(BeTrue())
	} else {
		EventuallyWithOffset(1, actual, intervals[0]).Should(BeTrue())
	}

	By(fmt.Sprintf("succesfully deleted resource %s", resourceUniqueID(resource)))
}

func ExpectCreateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("creating resource %s...", resourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Create(globalAPIContext, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully created resource %s", resourceUniqueID(resource)))
}

func ExpectUpdateResource(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("updating resource %s...", resourceUniqueID(resource)))

	ExpectWithOffset(1, k8sClient.Update(globalAPIContext, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully updated resource %s", resourceUniqueID(resource)))
}

func resourceUniqueID(resource client.Object) string {
	return fmt.Sprintf("[%s/%s/%s]",
		resource.GetNamespace(),
		resourceStructType(resource),
		resource.GetName())
}

func resourceStructType(resource client.Object) string {
	return reflect.TypeOf(resource).Elem().Name()
}
