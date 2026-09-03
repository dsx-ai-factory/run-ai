/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	multierror "github.com/hashicorp/go-multierror"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CoreV1ApiVersion = "v1"

	RbacGroup = "rbac.authorization.k8s.io"

	NamespaceKind      = "Namespace"
	ProjectKind        = "Project"
	DepartmentKind     = "Department"
	QueueKind          = "Queue"
	LimitRangeKind     = "LimitRange"
	SecretKind         = "Secret"
	ServiceAccountKind = "ServiceAccount"
	ClusterRoleKind    = "ClusterRole"

	AppLabel                  = "app"
	OcpClusterMonitoringLabel = "openshift.io/cluster-monitoring"

	// ManagedByLabel marks the per-project RoleBindings this controller owns. Other controllers
	// set the Project as owner of their own RoleBindings too, so ownership alone can't tell them
	// apart - only bindings carrying this label may be pruned.
	ManagedByLabel        = "app.kubernetes.io/managed-by"
	ProjectControllerName = "project-controller"

	//--------------------------- Replication Labels -------------------------------------------------------------------
	// Two groups are involved: (a) type labels  (b) scope labels
	//
	// Type labels:
	// - run.ai/asset          [asset sync owned resources]
	// - run.ai/resource       [user owned resources, only for secrets]
	//
	// Scope labels:
	// - run.ai/tenant-wide    [asset sync owned only]
	// - run.ai/cluster-wide   [user owned only]
	// - run.ai/department     [both asset sync and user owned]
	//------------------------------------------------------------------------------------------------------------------

	DockerRegistryAssetKind = "docker-registry"
	PasswordAssetKind       = "password"
	AccessKeyAssetKind      = "access-key"
	GenericSecretAssetKind  = "generic"

	FailedGetNamespaceStatusMsg             = "Failed to get namespace for project"
	ProjectsNamespaceNotFoundStatusMsg      = "Can't find namespace for project"
	QueueLabelMissingFromNamespaceStatusMsg = "label missing from namespace"
	FailedUpdateNamespaceStatusMsg          = "Failed to update namespace for project"
	FailedCreateNamespaceStatusMsg          = "Failed to create namespace for project"

	AgentNamespaceVersion = "v2"

	DefaultServiceAccountName      = "default"
	DefaultLimitRangeConfigMapName = "default-limit-range"

	CpuResourceName      = "cpu"
	MemoryResourceName   = "memory"
	CpuDefaultRequest    = "cpuDefaultRequest"
	MemoryDefaultRequest = "memoryDefaultRequest"
	CpuDefaultLimit      = "cpuDefaultLimit"
	MemoryDefaultLimit   = "memoryDefaultLimit"
	CpuMaxLimit          = "cpuMaxLimit"
	MemoryMaxLimit       = "memoryMaxLimit"

	LogProjectTag         = "Project"
	LogDepartmentTag      = "Department"
	LogNamespaceTag       = "Namespace"
	LogQueueTag           = "Queue"
	LogQueueNumTag        = "Queues"
	LogProjectStatusTag   = "ProjectStatus"
	LogSecretTag          = "Secret"
	LogLimitRangeTag      = "LimitRange"
	LogRoleBindingTag     = "RoleBinding"
	LogImagePullSecretTag = "ImagePullSecret"
	LogPvcTag             = "PersistentVolumeClaim"
	LogConfigMapTag       = "ConfigMap"

	LogClusterWideTag            = "ClusterWideReplicator"
	LogClusterWideGVKSpecificTag = "ClusterWideReplicatorGVKSpecific"
	LogScopeWideLabelKeyTag      = "ScopeWideLabelKey"
	LogScopeWideLabelValTag      = "ScopeWideLabelVal"
	LogGvkTag                    = "GroupVersionKind"

	GvkDeleteBlockersConfigMapName = "project-delete-blockers"

	ForceDeleteAnnotation = "kai/force-delete"
)

var (
	TrueRef               = true
	OrphanDeletePolicyRef = metav1.DeletePropagationOrphan
)

type WithLoggerAndCli struct {
	Client client.Client
	Log    logr.Logger
}

func (common WithLoggerAndCli) GetNamespace(namespaceName string) (namespace corev1.Namespace, err error) {
	err = common.Client.Get(
		context.Background(),
		client.ObjectKey{Name: namespaceName},
		&namespace,
	)
	return namespace, err
}

func (common WithLoggerAndCli) GetQueue(queueName string) (queue kaiv2.Queue, err error) {
	err = common.Client.Get(
		context.Background(),
		client.ObjectKey{Name: queueName},
		&queue,
	)
	return queue, err
}

func (common WithLoggerAndCli) GetQueuesForProjectByLabel(
	projectName string,
) (queues kaiv2.QueueList, err error) {
	err = common.Client.List(
		context.Background(),
		&queues,
		client.MatchingLabels(map[string]string{config.Get().ProjectLabelKey: projectName}),
	)
	if err != nil {
		common.Log.Error(err, "Failed listing queues with project name label selector", LogProjectTag, projectName)
	}

	return queues, err
}

func (common WithLoggerAndCli) KaiProjectToNamespace(project *kaiv1alpha1.Project) (result string, err error) {
	_, result, err = KaiProjectToNamespace(project, common.Client, common.Log)
	return result, err
}

func KaiProjectToNamespace(project *kaiv1alpha1.Project, kubeclient client.Client, logger logr.Logger,
) (isNotFound bool, result string, err error) {
	if project.Status.Namespace != "" {
		return false, project.Status.Namespace, nil
	}

	return projectToNamespace(project.Name, kubeclient, logger)
}

// projectToNamespace is a utility function for converting project name to namespace. This is done by interrogating
// the namespace project label of the namespaces.
func projectToNamespace(project string, kubeclient client.Client, logger logr.Logger,
) (isNotFound bool, result string, err error) {
	labelCondition := map[string]string{
		config.Get().NamespaceProjectLabelKey: project,
	}

	cantFindText := fmt.Sprintf("Failed to determine the namespace of project %s", project)

	items := corev1.NamespaceList{}
	err = kubeclient.List(context.Background(), &items, client.MatchingLabels(labelCondition))
	if err != nil {
		logger.Error(err, cantFindText)
		return false, "", err
	}

	if len(items.Items) == 0 {
		err = fmt.Errorf("no namespace found to have %s label with the name of the project", config.Get().NamespaceProjectLabelKey)
		logger.Error(err, cantFindText)
		return true, "", err
	}
	if len(items.Items) != 1 {
		err = fmt.Errorf(
			"multiple namespaces found to have %s label with the name of the project (%s, %s, ...)",
			config.Get().NamespaceProjectLabelKey, items.Items[0].Name, items.Items[1].Name)
		logger.Error(err, cantFindText)
		return false, "", err
	}

	return false, items.Items[0].Name, nil
}

func IsManuallyOverridden(meta metav1.Object) (isOverridden bool) {
	return strings.ToLower(meta.GetLabels()[config.Get().ResourceManualOverrideLabelKey]) == "true"
}

// IsProjectOwner returns the index of the given Project in obj's OwnerReferences,
// or -1 if the Project does not own obj. It is the KAI replacement for the run.ai
// Project.IsOwner method.
func IsProjectOwner(project *kaiv1alpha1.Project, obj client.Object) int {
	for index, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.Kind == ProjectKind &&
			ownerRef.UID == project.UID &&
			ownerRef.Name == project.Name {
			return index
		}
	}
	return -1
}

// IsForceDelete reports whether the object requests force deletion via the
// kai/force-delete annotation. It is the KAI replacement for the former
// run.ai deletionRequest.Force spec flag.
func IsForceDelete(meta metav1.Object) bool {
	return strings.ToLower(meta.GetAnnotations()[ForceDeleteAnnotation]) == "true"
}

func ContainsString(term string, list []string) bool {
	return IndexOfString(term, list) > -1
}

// DeleteTerm replaces the term to delete, if found, with the last element of the slice and returns the slice with the last index truncated.
func DeleteTerm(term string, list []string) (result []string) {
	if list != nil {
		indexOfTerm := IndexOfString(term, list)
		list[indexOfTerm] = list[len(list)-1]
		list[len(list)-1] = ""
		result = list[:len(list)-1]
	}
	return result
}

func IndexOfString(term string, list []string) int {
	for i, element := range list {
		if element == term {
			return i
		}
	}
	return -1
}

func AppendErrIfNotNil(err error, innerErr error) error {
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
	}
	return err
}
