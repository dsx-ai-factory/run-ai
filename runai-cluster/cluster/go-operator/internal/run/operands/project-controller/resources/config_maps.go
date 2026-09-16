/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

// ConfigMapForRunaiConfig
// Note to Dev: if you are changing anything in this struct, please make sure to update
// the script in "cluster/new-service-template/hooks/post_gen_project.py" (add_project_based_rbac_if_needed func)
func ConfigMapForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		name      = "runai-project-controller-rolebindings-plugin"
		namespace = runaiConfig.Namespace

		config = runaiConfig.Spec.ProjectController
	)

	configMap := &v1.ConfigMap{}

	// Get the existing config map if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	configMap.ObjectMeta.Name = name
	configMap.ObjectMeta.Namespace = namespace

	configMap.Data = func() map[string]string {
		data := map[string]string{}

		if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift {
			data["openshift-jupyter-notebook.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-jupyter-notebook-scc",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-jupyter-notebook-scc",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind: utils.ResourceStructType(&v1.ServiceAccount{}),
							Name: "default",
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()

			data["openshift-s3.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-openshift-s3-scc",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-openshift-s3-scc",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind: utils.ResourceStructType(&v1.ServiceAccount{}),
							Name: "default",
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()

			data["openshift-hostpath.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-user-job",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-user-job",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind: utils.ResourceStructType(&v1.ServiceAccount{}),
							Name: "default",
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()

			data["openshift-prometheus.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-prometheus-k8s",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-prometheus-k8s",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      "prometheus-k8s",
							Namespace: consts.OpenshiftMonitoringNamespace,
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()
		}

		if *config.ClusterWideSecret {
			data["project-secret.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-project-controller-cluster-secret-per-project",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-project-controller-cluster-secret-per-project",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      operands.RunaiProjectControllerName,
							Namespace: namespace,
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()
		}
		if *config.ClusterWideConfigMap {
			data["project-configmap.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-project-controller-cluster-configmap-per-project",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-project-controller-cluster-configmap-per-project",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      operands.RunaiProjectControllerName,
							Namespace: namespace,
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()
		}

		if *config.ClusterWidePvc {
			data["project-pvc.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-project-controller-cluster-pvc-per-project",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-project-controller-cluster-pvc-per-project",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      operands.RunaiProjectControllerName,
							Namespace: namespace,
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()
		}

		if *config.LimitRange {
			data["limit-range.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-project-controller-limit-range-per-project",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-project-controller-limit-range-per-project",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      operands.RunaiProjectControllerName,
							Namespace: namespace,
						},
					},
				}

				roleBindingYAML, err := yaml.Marshal(roleBinding)
				if err != nil {
					panic(err)
				}

				return string(roleBindingYAML)
			}()
		}

		data["accessrule-controller.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-accessrule-controller-project",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-accessrule-controller-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      operands.RunaiAccessRuleControllerName,
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["researcher-service-rw.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "researcher-service-rw",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "researcher-service-rw",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "researcher-service",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["cluster-api-project.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-api-project",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "cluster-api-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "cluster-api",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["runai-scheduler-rw.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-scheduler-rw",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-scheduler-rw",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "runai-scheduler",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["runai-project-controller-project.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-project-controller-project",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-project-controller-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      operands.RunaiProjectControllerName,
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["runai-admission-controller-project.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-admission-controller-project",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-admission-controller-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "runai-admission-controller",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["runai-binder-project.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-binder-project",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-binder-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "binder",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["status-updater.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "status-updater",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "status-updater",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "status-updater",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["pod-group-controller.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-group-controller",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "pod-group-controller",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "podgroup-controller",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["numa-placement-exporter.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "numa-placement-exporter",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "numa-placement-exporter",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "numa-placement-exporter",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["external-workload-integrator.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "external-workload-integrator",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-external-workload-integrator-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "external-workload-integrator",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["cluster-api.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-api",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-cluster-api-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "cluster-api",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["shared-objects-controller.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "shared-objects-controller",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-shared-objects-controller-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "shared-objects-controller",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["advice-restart-controller.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "advice-restart-controller",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-advice-restart-controller-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "advice-restart-controller",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["cluster-sync.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-sync",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "cluster-sync-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "cluster-sync",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["runai-agent.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-agent",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-agent-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "runai-agent",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["runai-agent-additional-workload-gvks.yaml"] = func() string {
			crudClusterRoleName := runaiConfig.Spec.ExternalWorkloadIntegrator.RIDRbacConfig.ClusterRoleCrudName

			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-agent-additional-workload-gvks",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     crudClusterRoleName,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "runai-agent",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		data["anyworkload-controller.yaml"] = func() string {
			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/%s",
						rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
					Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "anyworkload-controller",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: rbacv1.SchemeGroupVersion.Group,
					Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
					Name:     "runai-anyworkload-controller-project",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
						Name:      "anyworkload-controller",
						Namespace: namespace,
					},
				},
			}

			roleBindingYAML, err := yaml.Marshal(roleBinding)
			if err != nil {
				panic(err)
			}

			return string(roleBindingYAML)
		}()

		return data
	}()

	return configMap, nil
}

func ConfigMapStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	configMap, err := ConfigMapForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if configMap.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
