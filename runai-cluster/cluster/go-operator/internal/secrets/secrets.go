/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secrets

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	InstallationConfigurationSecretName = "installation-configuration"
	SecretName                          = "whisper-%s-secret"
	deploymentSecretPath                = constants.SecretMountDirName
)

func MountToDeployment(ctx context.Context, reader client.Reader, d *appsv1.Deployment) error {
	secret := &corev1.Secret{}
	deploymentSecretName := generateSecretName(d.Name)
	err := getSecret(ctx, reader, deploymentSecretName, runaiv1.DefaultNamespace, secret)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// no data no set of secret
	if secret.Data == nil {
		return nil
	}

	// no data for deployment
	if _, ok := secret.Data[d.Name]; !ok {
		return nil
	}

	spec := &d.Spec.Template.Spec
	if spec.Volumes == nil {
		spec.Volumes = make([]corev1.Volume, 0)
	}

	spec.Volumes = append(spec.Volumes, corev1.Volume{
		Name: deploymentSecretName,
		VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
			SecretName: deploymentSecretName,
			Items: []corev1.KeyToPath{
				{
					Key:  d.Name,
					Path: constants.SecretMountFileName,
				},
			},
		}},
	})

	for i := range spec.Containers {
		if spec.Containers[i].VolumeMounts == nil {
			spec.Containers[i].VolumeMounts = make([]corev1.VolumeMount, 0)
		}

		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      deploymentSecretName,
			MountPath: deploymentSecretPath,
			ReadOnly:  true,
		})
	}

	return nil
}

func GetDeploymentSecret(ctx context.Context, reader client.Reader, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	deploymentSecretName := generateSecretName(name)
	err := getSecret(ctx, reader, deploymentSecretName, runaiv1.DefaultNamespace, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	secret.Name = deploymentSecretName
	secret.Namespace = runaiv1.DefaultNamespace
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	secret.Labels[constants.LabelName] = deploymentSecretName
	return secret, nil
}

func GetConfigurationSecret(ctx context.Context, reader client.Reader) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := getSecret(ctx, reader, InstallationConfigurationSecretName, runaiv1.DefaultNamespace, secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func getSecret(ctx context.Context, reader client.Reader, name, ns string, secret *corev1.Secret) error {
	err := reader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: ns,
	}, secret)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func generateSecretName(name string) string {
	return fmt.Sprintf(SecretName, name)
}
