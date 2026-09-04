/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ri

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CRDInstalled reports whether the legacy ResourceInterface CRD exists on the
// cluster. The migration is a no-op when it does not (fresh installs).
func CRDInstalled(ctx context.Context, c client.Client) (bool, error) {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := c.Get(ctx, types.NamespacedName{Name: CRDName}, crd)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteCRD removes the legacy ResourceInterface CRD. Run only after all RIs
// have been removed. NotFound is treated as success.
func DeleteCRD(ctx context.Context, c client.Client) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	crd.SetName(CRDName)
	err := c.Delete(ctx, crd)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
