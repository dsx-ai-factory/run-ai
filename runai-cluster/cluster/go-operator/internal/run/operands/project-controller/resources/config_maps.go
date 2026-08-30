package resources

import (
	"context"
	"fmt"

	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
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

		if *runaiConfig.Spec.JobController.Enabled {
			data["job-controller.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-job-controller-project",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-job-controller-project",
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

		if *runaiConfig.Spec.WorkloadController.Enabled {
			data["workload-controller-rw.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "workload-controller-rw",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "workload-controller-rw",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      "workload-controller",
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

		if *runaiConfig.Spec.JobController.Enabled {
			data["runai-job-controller-project.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "runai-job-controller-project",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-job-controller-project",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      "runai-job-controller",
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

		if *runaiConfig.Spec.InferenceWorkloadController.Enabled {
			data["inference-workload-controller.yaml"] = func() string {
				roleBinding := &rbacv1.RoleBinding{
					TypeMeta: metav1.TypeMeta{
						APIVersion: fmt.Sprintf("%s/%s",
							rbacv1.SchemeGroupVersion.Group, rbacv1.SchemeGroupVersion.Version),
						Kind: utils.ResourceStructType(&rbacv1.RoleBinding{}),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "inference-workload-controller",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Group,
						Kind:     utils.ResourceStructType(&rbacv1.ClusterRole{}),
						Name:     "runai-inference-workload-controller-project",
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      utils.ResourceStructType(&v1.ServiceAccount{}),
							Name:      "inference-workload-controller",
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

// blockersConfigMapKey is the single ConfigMap entry holding the YAML list of blockers.
const blockersConfigMapKey = "blockers.yaml"

// BlockersConfigMapForRunaiConfig builds the "project-delete-blockers" ConfigMap read by the
// project-controller at startup to construct its generic project-deletion blockers. The
// runai blocker set (runaiDeletionBlockers) reproduces the historical
// Workload/DataVolume/Secret/PVC blockers. The project-controller is pointed at this
// namespace via --project-delete-blockers-namespace (see buildContainerArgs).
func BlockersConfigMapForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		name      = clustercommonconstants.GvkDeleteBlockersConfigMapName
		namespace = runaiConfig.Namespace
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

	// A single ConfigMap entry holding the YAML list of blockers the project-controller
	// parses at startup. Blockers sharing a displayName are grouped into one project
	// condition by the controller.
	encoded, err := yaml.Marshal(runaiDeletionBlockers())
	if err != nil {
		return nil, fmt.Errorf("marshalling project-delete-blockers: %w", err)
	}
	configMap.Data = map[string]string{blockersConfigMapKey: string(encoded)}

	return configMap, nil
}

func BlockersConfigMapStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	configMap, err := BlockersConfigMapForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if configMap.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}

// blocker mirrors the YAML shape the project-controller parses from the project-delete-blockers
// ConfigMap (json tags == the wire contract): a flat list where blockers sharing a displayName
// are grouped into one project condition by the controller. It is defined locally so the
// go-operator has no build-time dependency on project-controller — the only contract between
// them is these YAML field names (the same arrangement as the rolebindings ConfigMap, where
// go-operator marshals RoleBinding objects the controller decodes).
type blocker struct {
	DisplayName   string                `json:"displayName"`
	Group         string                `json:"group"`
	Version       string                `json:"version"`
	Kind          string                `json:"kind"`
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// runaiDeletionBlockers is the runai-specific set of project-deletion blockers. It reproduces
// the historically hard-coded Workload/DataVolume/Secret/PVC blockers and is the single source
// of truth for the runai packaging. Blockers sharing a displayName (Workloads, Secrets) are
// grouped by the project-controller into a single project condition.
func runaiDeletionBlockers() []blocker {
	const (
		runaiGroup = "run.ai"
		v2alpha1   = "v2alpha1"
		v1alpha1   = "v1alpha1"
	)
	assetKinds := []string{"password", "access-key", "generic", "docker-registry"}

	// projectScopeExpressions excludes department / cluster-wide / tenant-wide scoped
	// resources, matching common.getProjectScopeLabelSelector in project-controller.
	projectScopeExpressions := []metav1.LabelSelectorRequirement{
		{Key: clustercommonconstants.DepartmentScopeLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
		{Key: clustercommonconstants.ClusterWideResourceLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
		{Key: clustercommonconstants.TenantScopeLabel, Operator: metav1.LabelSelectorOpDoesNotExist},
	}

	projectScopeWithType := func(typeLabel string) *metav1.LabelSelector {
		exprs := append(projectScopeExpressions, metav1.LabelSelectorRequirement{
			Key: typeLabel, Operator: metav1.LabelSelectorOpIn, Values: assetKinds,
		})
		return &metav1.LabelSelector{MatchExpressions: exprs}
	}

	return []blocker{
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "TrainingWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "InferenceWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "DistributedWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v2alpha1, Kind: "InteractiveWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v1alpha1, Kind: "ExternalWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: v1alpha1, Kind: "DistributedInferenceWorkload"},
		{DisplayName: "Workloads", Group: runaiGroup, Version: "v1", Kind: "RunaiJob"},
		{DisplayName: "DataVolumes", Group: runaiGroup, Version: v1alpha1, Kind: "DataVolume"},
		{DisplayName: "Secrets", Group: "", Version: "v1", Kind: "Secret",
			LabelSelector: projectScopeWithType(clustercommonconstants.RunaiResourceLabel)},
		{DisplayName: "Secrets", Group: "", Version: "v1", Kind: "Secret",
			LabelSelector: projectScopeWithType(clustercommonconstants.RunaiAssetLabel)},
		{DisplayName: "Pvcs", Group: "", Version: "v1", Kind: "PersistentVolumeClaim",
			LabelSelector: &metav1.LabelSelector{MatchExpressions: projectScopeExpressions}},
	}
}
