/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeploymentName      = "cluster-redis"
	valkeyConfFileName  = "valkey.conf"
	valkeyConfMountPath = "/usr/local/etc/valkey"
	valkeyDataMountPath = "/data"
	valkeyImageUID      = int64(999)
)

type DeploymentResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewDeploymentResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *DeploymentResource {
	return &DeploymentResource{operandName: operandName, k8sReader: k8sReader, runaiConfig: runaiConfig}
}

func (d *DeploymentResource) OperandName() string { return d.operandName }

func (d *DeploymentResource) Object(ctx context.Context) (client.Object, error) {
	var (
		deploymentNamespace = d.runaiConfig.GetNamespace()
		config              = d.runaiConfig.Spec.ClusterRedis
		image               = config.Image
		replicas            = config.Replicas
	)
	runAsUser := ptr.To(valkeyImageUID)
	if d.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runaiv1internal.KubernetesFlavorOpenShift {
		runAsUser = nil
	}
	deployment := &appsv1.Deployment{}
	err := d.k8sReader.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: deploymentNamespace}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	deployment.Name = DeploymentName
	deployment.Namespace = deploymentNamespace

	deployment.Spec.Replicas = replicas
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": DeploymentName},
	}

	deployment.Spec.Template.Name = DeploymentName
	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName

	deployment.Spec.Template.Spec = v1.PodSpec{
		Affinity: outils.MergeAffinities(config.Affinity, d.runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, d.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations:                  outils.GetTolerations(d.runaiConfig),
		AutomountServiceAccountToken: ptr.To(false),
		Containers: []v1.Container{
			{
				Name: DeploymentName,
				SecurityContext: &v1.SecurityContext{
					AllowPrivilegeEscalation: ptr.To(false),
					RunAsNonRoot:             ptr.To(true),
					RunAsUser:                runAsUser,
					ReadOnlyRootFilesystem:   ptr.To(true),
				},
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				Resources:       v1.ResourceRequirements(*config.Resources),
				Command:         []string{"valkey-server"},
				Args:            []string{fmt.Sprintf("%s/%s", valkeyConfMountPath, valkeyConfFileName)},
				VolumeMounts: []v1.VolumeMount{
					{Name: "valkey-config", MountPath: valkeyConfMountPath},
					{Name: "valkey-data", MountPath: valkeyDataMountPath},
				},
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(d.runaiConfig),
		Volumes: []v1.Volume{
			{
				Name: "valkey-config",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: DeploymentName,
						Items:      []v1.KeyToPath{{Key: valkeyConfFileName, Path: valkeyConfFileName}},
					},
				},
			},
			{
				Name: "valkey-data",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	return deployment, nil
}

func (d *DeploymentResource) Status(ctx context.Context) error {
	resource, err := d.Object(ctx)
	if err != nil {
		return err
	}
	deployment := resource.(*appsv1.Deployment)
	if deployment.GetUID() == "" {
		return fmt.Errorf("deployment [%s] is missing", deployment.GetName())
	}
	deplopymentStatus := deployment.Status
	if len(deplopymentStatus.Conditions) == 0 {
		return fmt.Errorf("deployment [%s] has no conditions", deployment.GetName())
	}
	for _, condition := range deplopymentStatus.Conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return nil
		}
	}
	return fmt.Errorf("deployment [%s] has not reached the Available condition", deployment.GetName())
}

func (d *DeploymentResource) Enabled() bool { return true }
