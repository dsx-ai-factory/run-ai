/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package image_pull_secret

import (
	"context"
	"encoding/base64"
	"strings"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateImagePullSecret(k8s client.Client, namespace, secretName, imagePullSecretData string) error {
	if strings.TrimSpace(imagePullSecretData) == "" {
		return nil
	}

	decodedImagePullSecretData, err := base64.StdEncoding.DecodeString(imagePullSecretData)
	if err != nil {
		return err
	}

	imagePullSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      secretName,
		},
	}

	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(imagePullSecret), imagePullSecret)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	imagePullSecret.Type = corev1.SecretTypeDockerConfigJson
	imagePullSecret.Data = map[string][]byte{
		corev1.DockerConfigJsonKey: decodedImagePullSecretData,
	}

	if imagePullSecret.UID == "" {
		err = k8s.Create(context.TODO(), imagePullSecret)
		if err != nil {
			return err
		}
	} else {
		err = k8s.Update(context.TODO(), imagePullSecret)
		if err != nil {
			return err
		}
	}

	return nil
}
