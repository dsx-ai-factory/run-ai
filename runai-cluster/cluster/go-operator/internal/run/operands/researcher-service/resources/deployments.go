/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"strconv"

	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeploymentName = "researcher-service"
)

type DeploymentResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewDeploymentResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *DeploymentResource {
	return &DeploymentResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (d *DeploymentResource) OperandName() string {
	return d.operandName
}

func (d *DeploymentResource) Object(ctx context.Context) (client.Object, error) {
	if !d.Enabled() {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DeploymentName,
				Namespace: d.runaiConfig.GetNamespace(),
			},
		}, nil
	}

	controlPlane := d.runaiConfig.Spec.Internal.Global.ControlPlane
	replicas := d.runaiConfig.Spec.ResearcherService.Replicas

	oauthProxy, oAuthProxyOK := controlPlane.Config.ResearcherService.GetOauthProxyOk()
	if !oAuthProxyOK {
		return nil, fmt.Errorf("oauth config is missing in the researcher-service config")
	} else {
		if _, ok := oauthProxy.GetOidcIssuerUrlOk(); !ok {
			return nil, fmt.Errorf("issuer-url is missing in researcher-service oauth config")
		}
	}

	deployment := &appsv1.Deployment{}

	err := d.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: d.runaiConfig.GetNamespace(),
	}, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	initDeployment(d.runaiConfig, deployment)

	deployment.Spec.Replicas = replicas

	return deployment, nil
}

func initDeployment(runaiConfig *runaiv1.RunaiConfig, deployment *appsv1.Deployment) {
	var (
		deploymentNamespace = runaiConfig.GetNamespace()
		global              = runaiConfig.Spec.Global
		config              = runaiConfig.Spec.ResearcherService
		controlPlane        = runaiConfig.Spec.Internal.Global.ControlPlane
		resources           = config.Resources
		image               = config.Image
		args                = config.Args
	)

	deployment.Name = DeploymentName
	deployment.Namespace = deploymentNamespace

	deployment.Spec.Replicas = ptr.To[int32](1)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": DeploymentName,
		},
	}

	deployment.Spec.Template.Labels = outils.InitMapIfNil(deployment.Spec.Template.Labels)
	deployment.Spec.Template.Labels["app"] = DeploymentName

	deployment.Spec.Template.Spec = v1.PodSpec{
		ServiceAccountName: DeploymentName,
		Affinity: outils.MergeAffinities(config.Affinity, runaiConfig.Spec.Global.Affinity,
			deployment.Spec.Selector.MatchLabels, runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		Tolerations: outils.GetTolerations(runaiConfig),
		Containers: []v1.Container{
			{
				Name: DeploymentName,
				Args: func() []string {
					arg := []string{}
					arg = append(arg, "--v="+strconv.Itoa(*args.Verbosity))
					arg = append(arg, "--auth-enabled="+strconv.FormatBool(*args.AuthEnabled))
					arg = append(arg, "--openshift="+strconv.FormatBool(runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift))
					if runaiConfig.Spec.Internal.Global.ClusterURL != "" {
						arg = append(arg, "--base-url", runaiConfig.Spec.Internal.Global.ClusterURL)
					}
					if logoutLink, ok := controlPlane.Config.ResearcherService.GetLogoutLinkOk(); ok {
						arg = append(arg, "--logout-link", *logoutLink)
					}
					if args.GkeOidcEnvoyHost != "" {
						arg = append(arg, "--gke-oidc-envoy-host", args.GkeOidcEnvoyHost)
					}

					rsControlPlaneConfig := controlPlane.Config.GetResearcherService()
					oauth := rsControlPlaneConfig.GetOauthProxy()
					arg = append(arg, "--issuer-url="+oauth.GetOidcIssuerUrl())

					if global.EnableWorkloadOwnershipProtection != nil && *global.EnableWorkloadOwnershipProtection {
						arg = append(arg, "--enable-workloads-ownership-protection")
					}

					arg = append(arg, "--cache-expiration-period="+args.CacheExpirationPeriod.String())

					return arg
				}(),
				SecurityContext: outils.GlobalSecurityContext(runaiConfig),
				Image:           fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag),
				ImagePullPolicy: image.PullPolicy,
				VolumeMounts:    outils.GlobalVolumeMounts(runaiConfig),
				Resources:       v1.ResourceRequirements(resources.ResearcherService),
				Ports: []v1.ContainerPort{
					{
						ContainerPort: 8080,
					},
				},
				Env: []v1.EnvVar{
					{
						Name:  "CLUSTER_NAME",
						Value: *runaiConfig.Spec.Internal.Global.ControlPlane.Config.ClusterName,
					},
					{
						Name:  "CLUSTER_DOMAIN",
						Value: runaiConfig.Spec.Internal.Global.ClusterURL,
					},
					{
						Name:  "SUBDOMAIN_SUPPORT",
						Value: strconv.FormatBool(ptr.Deref(global.SubdomainSupport, true)),
					},
					{
						Name:  "GODEBUG",
						Value: outils.FipsGodebugValue(runaiConfig, "off"),
					},
				},
			},
		},
		ImagePullSecrets: outils.GetGlobalImagePullSecrets(runaiConfig),
		Volumes:          outils.GlobalVolumes(runaiConfig),
	}
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

	conditions := deplopymentStatus.Conditions
	for _, condition := range conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == v1.ConditionStatus(metav1.ConditionTrue) {
			return nil
		}
	}

	return fmt.Errorf("deployment [%s] has not reached the Available condition", deployment.GetName())
}

func (d *DeploymentResource) Enabled() bool {
	return componentEnabled
}
