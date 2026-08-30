package kubeConfig

import (
	"context"
	"fmt"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	dwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateIW(ctx context.Context, k8sClient client.Client, interactiveWorkload iwv2a1.InteractiveWorkload) error {
	err := k8sClient.Update(ctx, &interactiveWorkload)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return err
	}
	return nil
}

func GetIW(ctx context.Context, k8sClient client.Client, namespace string, interactiveWorkloadName string) (iwv2a1.InteractiveWorkload, error) {
	interactiveWorkload := iwv2a1.InteractiveWorkload{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: interactiveWorkloadName}, &interactiveWorkload)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return interactiveWorkload, err
	}
	return interactiveWorkload, nil
}

func ListIW(ctx context.Context, k8sClient client.Client, namespace string) (iwv2a1.InteractiveWorkloadList, error) {
	interactiveWorkloadList := iwv2a1.InteractiveWorkloadList{}
	err := k8sClient.List(ctx, &interactiveWorkloadList, client.InNamespace(namespace))
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return interactiveWorkloadList, err
	}
	return interactiveWorkloadList, nil
}

func UpdateTW(ctx context.Context, k8sClient client.Client, trainingWorkload twv2a1.TrainingWorkload) error {
	err := k8sClient.Update(ctx, &trainingWorkload)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return err
	}
	return nil
}

func GetTW(ctx context.Context, k8sClient client.Client, namespace string, trainingWorkloadName string) (twv2a1.TrainingWorkload, error) {
	trainingWorkload := twv2a1.TrainingWorkload{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: trainingWorkloadName}, &trainingWorkload)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return trainingWorkload, err
	}
	return trainingWorkload, nil
}

func ListTW(ctx context.Context, k8sClient client.Client, namespace string) (twv2a1.TrainingWorkloadList, error) {
	trainingWorkloadList := twv2a1.TrainingWorkloadList{}
	err := k8sClient.List(ctx, &trainingWorkloadList, client.InNamespace(namespace))
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return trainingWorkloadList, err
	}
	return trainingWorkloadList, nil
}

func UpdateDW(ctx context.Context, k8sClient client.Client, distributedWorkload dwv2a1.DistributedWorkload) error {
	err := k8sClient.Update(ctx, &distributedWorkload)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return err
	}
	return nil
}

func GetDW(ctx context.Context, k8sClient client.Client, namespace string, distributedWorkloadName string) (dwv2a1.DistributedWorkload, error) {
	distributedWorkload := dwv2a1.DistributedWorkload{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: distributedWorkloadName}, &distributedWorkload)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return distributedWorkload, err
	}
	return distributedWorkload, nil
}

func ListDW(ctx context.Context, k8sClient client.Client, namespace string) (dwv2a1.DistributedWorkloadList, error) {
	distributedWorkloadList := dwv2a1.DistributedWorkloadList{}
	err := k8sClient.List(ctx, &distributedWorkloadList, client.InNamespace(namespace))
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return distributedWorkloadList, err
	}
	return distributedWorkloadList, nil
}

func GetWLR(ctx context.Context, k8sClient client.Client, namespace string, workloadRunnerName string) (dromosv1alpha1.WorkloadRunner, error) {
	workloadRunner := dromosv1alpha1.WorkloadRunner{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: workloadRunnerName}, &workloadRunner)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return workloadRunner, err
	}
	return workloadRunner, nil
}

func ListWLR(ctx context.Context, k8sClient client.Client, namespace string) (dromosv1alpha1.WorkloadRunnerList, error) {
	workloadRunnerList := dromosv1alpha1.WorkloadRunnerList{}
	err := k8sClient.List(ctx, &workloadRunnerList, client.InNamespace(namespace))
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return workloadRunnerList, err
	}
	return workloadRunnerList, nil
}

// ListProjects lists the kai.resources projects.
func ListProjects(ctx context.Context, k8sClient client.Client) (kaiorg.ProjectList, error) {
	projectList := kaiorg.ProjectList{}
	err := k8sClient.List(ctx, &projectList)
	if err != nil {
		zap.S().Errorf("err %+v ", err)
		return kaiorg.ProjectList{}, err
	}
	return projectList, nil
}

func ProjectNameOfNamespace(ctx context.Context, k8sClient client.Client, namespace string, workloadLabels map[string]string) (string, error) {
	if workloadLabels == nil {
		zap.S().Warnf("workloadLabels is nil, getting workload project using namespace %s", namespace)
	} else {
		workloadProject := workloadLabels[constants.RunaiProjectLabel]
		if workloadProject == "" {
			zap.S().Warnf("runai-project label not found on workload, getting workload project using namespace %s", namespace)
		} else {
			return workloadProject, nil
		}
	}

	var ns corev1.Namespace
	err := k8sClient.Get(ctx, client.ObjectKey{Name: namespace}, &ns)
	if err != nil {
		return "", err
	}

	if ns.Labels[constants.RunaiQueueLabel] == "" {
		return "", fmt.Errorf("runai-queue label not found on namespace %s", namespace)
	}

	return ns.Labels[constants.RunaiQueueLabel], nil
}
