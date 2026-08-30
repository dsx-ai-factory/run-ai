package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubeClient "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client"
	admission_utils "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/utils"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/priority"
)

const (
	runaiProjectEnvVar = "RUNAI_PROJECT"

	nonPreemptibleValue = "non-preemptible"

	nodeAffinityInteractiveSpecField = "nodeAffinityInteractive"
	nodeAffinityTrainSpecField       = "nodeAffinityTrain"
)

// NodeAffinityTypesFromAnnotation returns the node types a workload of this kind may run
// on, read from the kai.resources Project, where the agent writes them JSON-encoded.
func NodeAffinityTypesFromAnnotation(project *unstructured.Unstructured, interactive bool) []string {
	annotationKey := commonconstants.NodeAffinityTrainAnnotation
	if interactive {
		annotationKey = commonconstants.NodeAffinityInteractiveAnnotation
	}

	encoded, found := project.GetAnnotations()[annotationKey]
	if !found || encoded == "" {
		return nil
	}

	var nodeTypes []string
	if err := json.Unmarshal([]byte(encoded), &nodeTypes); err != nil {
		log.Printf("error: failed to parse annotation %s of project %s: %v",
			annotationKey, project.GetName(), err)
		return nil
	}

	return nodeTypes
}

func GetDesiredProjectName(client kubeClient.Client, namespace string, containers []v1.Container) string {
	// normally, when the job is part of a runai namespace, we expect the project name to be embedded
	// in the runai queue label of the namespace
	projectName, _ := NamespaceToProject(client, namespace, nil)

	// in special cases (kubeflow notebook, for example) the project can be found in
	// environment variable RUNAI_PROJECT
	if projectName == "" && containers != nil {
		_, projectName = admission_utils.GetEnvVarFromAnyContainer(containers, runaiProjectEnvVar)
	}

	return projectName
}

func NamespaceToProject(client kubeClient.Client, namespaceName string, meta *metav1.ObjectMeta) (string, bool) {
	namespace, err := client.GetNamespace(namespaceName)
	if err != nil {
		panic(err.Error())
	}

	project, found := namespace.Labels["runai/queue"]

	if !found && meta != nil && meta.Labels != nil {
		project = meta.Labels["project"]
		found = project != ""
	}

	return project, found
}

func GetProjectResource(client kubeClient.Client, projectName string) (*unstructured.Unstructured, error) {
	project, err := client.GetProject(projectName)
	if err != nil {
		if errors.IsNotFound(err) {
			err = fmt.Errorf(
				"project %s does not exist. Use a valid Project name or contact your administrator to create a new Project",
				projectName)
		} else {
			err = fmt.Errorf("error: could not get project %s, err: %w", projectName, err)
		}
		log.Printf("%s", err.Error())
		return nil, err
	}
	return project, nil
}

func IsKnownPriorityClassName(priorityClassName string) bool {
	return slices.Contains(
		[]string{priority.InteractivePriorityClassName, priority.InteractivePreemptiblePriorityClassName,
			priority.TrainPriorityClassName, priority.InferencePriorityClassName,
		},
		priorityClassName)
}

// IsInteractiveByPriority decides whether a workload should receive the project's interactive node
// affinity (vs. the training one). A workload is interactive when it is non-preemptible.
// The signal is resolved with the following precedence:
//  1. the preemptibility label (kai.scheduler/preemptibility), which since v2.25 may be set
//     independently of the priority class;
//  2. the priorityClassName label, resolved to its priority class value (>= threshold => interactive);
//  3. the priority class name from the pod spec, resolved the same way;
//  4. otherwise train (preemptible), preserving the pre-fix default.
func IsInteractiveByPriority(reader client.Reader, objectLabels map[string]string,
	specPriorityClassName string) bool {
	if preemptibility := objectLabels[commonconstants.LabelPreemptibilityName]; preemptibility != "" {
		return preemptibility == nonPreemptibleValue
	}

	priorityClassName := objectLabels[commonconstants.LabelPriorityClassName]
	if priorityClassName == "" {
		priorityClassName = specPriorityClassName
	}
	if priorityClassName == "" {
		return false
	}

	return getPriorityClassValue(reader, priorityClassName) >= priority.PriorityPreemptionThreshold
}

func getPriorityClassValue(reader client.Reader, priorityClassName string) int32 {
	priorityClass, err := priority.GetPriorityClass(context.Background(), priorityClassName, reader)
	if err == nil && priorityClass != nil {
		return priorityClass.Value
	}

	// Fail closed to train (preemptible): an unresolved/typo priority class must not be classified
	// as interactive by borrowing the cluster default priority.
	log.Printf("Failed to get priorityClass <%s> (%v); defaulting to train priority",
		priorityClassName, err)
	return priority.TrainPriorityValue
}

func IsPreemptible(priorityClassName string) bool {
	return slices.Contains(
		[]string{priority.TrainPriorityClassName, priority.InteractivePreemptiblePriorityClassName},
		priorityClassName)
}
