/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package certificates

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmv1 "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeleteOldCertificates(ctx context.Context, cmv1Client cmv1.CertmanagerV1Interface, namespace string) error {
	var oldCertificates *certmanagerv1.CertificateList
	var err error
	if oldCertificates, err = cmv1Client.Certificates(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=runai",
	}); err != nil {
		if k8sErrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			return nil
		}
		return err
	}

	for _, cert := range oldCertificates.Items {
		if err := cmv1Client.Certificates(namespace).Delete(ctx, cert.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
