package utils

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ResourceStructType(resource client.Object) string {
	return reflect.TypeOf(resource).Elem().Name()
}

func NamespacedNameFor(resource client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}
}

func GetDefaultRunaiConfig() *runaiv1.RunaiConfig {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.SetDefaultsWhereNeeded()

	return runaiConfig
}

func ResourceUniqueID(resource client.Object) string {
	return fmt.Sprintf("[%s/%s/%s]",
		resource.GetNamespace(),
		ResourceStructType(resource),
		resource.GetName())
}

func ListProjectsNamespacesNames(ctx context.Context, k8sReader client.Reader) ([]string, error) {
	namespaces := &v1.NamespaceList{}

	err := k8sReader.List(ctx, namespaces, client.HasLabels{"runai/queue"})
	if err != nil {
		return nil, err
	}

	namespaceNames := []string{}

	for _, namespace := range namespaces.Items {
		namespaceNames = append(namespaceNames, namespace.Name)
	}

	return namespaceNames, nil
}

func VolumeMountsForHTTPProxy(runaiConfig *runaiv1.RunaiConfig) []v1.VolumeMount {
	global := runaiConfig.Spec.Global
	if global.HTTPProxy.Enabled && global.HTTPProxy.TLSCertSecret != nil {
		return []v1.VolumeMount{
			{
				MountPath: "/opt/runai/tls.crt",
				Name:      global.HTTPProxy.ConfigMapName,
				SubPath:   "tls.crt",
			},
		}
	}

	return nil
}

func VolumesForHTTPProxy(runaiConfig *runaiv1.RunaiConfig) []v1.Volume {
	global := runaiConfig.Spec.Global
	if global.HTTPProxy.Enabled && global.HTTPProxy.TLSCertSecret != nil {
		return []v1.Volume{
			{
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: global.HTTPProxy.ConfigMapName,
						},
					},
				},
				Name: global.HTTPProxy.ConfigMapName,
			},
		}
	}

	return nil
}

func PtrFrom[T any](v T) *T {
	return &v
}

func NamespaceSelectorRunaiQueue() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "runai/queue",
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
}

func GetDeploymentImage(ctx context.Context, kbClient client.Client, name, namespace string) (string, error) {
	deployment := &appsv1.Deployment{}

	err := kbClient.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, deployment)
	if err != nil {
		return "", err
	}

	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return "", fmt.Errorf("deployment %s has more than one container (%d)", name, len(containers))
	}

	return containers[0].Image, nil
}
func GetDeploymentArgs(deployment appsv1.Deployment) ([]string, error) {
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return nil, fmt.Errorf("deployment %s has more than one container (%d)", deployment.Name, len(containers))
	}
	return containers[0].Args, nil
}

func GetImageTag(image string) (string, error) {
	imageSplit := strings.Split(image, ":")
	tag := imageSplit[len(imageSplit)-1]
	if len(imageSplit) < 2 || tag == "" || strings.Contains(tag, "/") {
		return "", fmt.Errorf("cannot determine image '%s' tag", image)
	}

	return tag, nil
}

func GetCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func MapToArgFormat(m map[string]string) string {
	arg := ""
	for key, value := range m {
		arg += fmt.Sprintf("%s=%s,", key, value)
	}

	if strings.HasSuffix(arg, ",") {
		arg = arg[:len(arg)-1]
	}

	return arg
}
