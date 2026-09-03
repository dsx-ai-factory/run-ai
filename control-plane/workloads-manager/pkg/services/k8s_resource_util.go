/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/constants"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

var (
	defaultNodePool = []string{composer.DefaultNodePoolName}
)

// K8sResourceUtil provides utilities for handling Kubernetes resources
type K8sResourceUtil struct {
}

type resourceMetadataUpdateOptions struct {
	trustedConvertedFromNative bool
}

// NewK8sResourceUtil creates a new instance of K8sResourceUtil
func NewK8sResourceUtil() *K8sResourceUtil {
	return &K8sResourceUtil{}
}

// ValidateK8sResource validates that the provided resource map represents a valid Kubernetes resource
func (k *K8sResourceUtil) ValidateK8sResource(resource map[string]interface{}) *api.Error {
	if resource == nil {
		return api.NewValidationError("resource cannot be nil")
	}

	// Check for required Kubernetes resource fields
	apiVersion, ok := resource[constants.FieldAPIVersion].(string)
	if !ok || apiVersion == "" {
		return api.NewValidationError("missing required field: apiVersion")
	}

	kind, ok := resource[constants.FieldKind].(string)
	if !ok || kind == "" {
		return api.NewValidationError("missing required field: kind")
	}

	return nil
}

// IsValidK8sResource checks if the provided resource map represents a valid Kubernetes resource
func (k *K8sResourceUtil) IsValidK8sResource(resource map[string]interface{}) bool {
	return k.ValidateK8sResource(resource) == nil
}

// UpdateResourceMetadata adds workload ID to the resource metadata, namespace, and labels
func (k *K8sResourceUtil) UpdateResourceMetadata(ctx context.Context, resource map[string]interface{}, workloadId string, project *openapi.Project) (map[string]interface{}, error) {
	return k.updateResourceMetadataWithOptions(ctx, resource, workloadId, project, resourceMetadataUpdateOptions{})
}

func (k *K8sResourceUtil) updateResourceMetadataWithOptions(
	ctx context.Context,
	resource map[string]interface{},
	workloadId string,
	project *openapi.Project,
	options resourceMetadataUpdateOptions,
) (map[string]interface{}, error) {
	// Create a deep copy to avoid modifying the original
	modifiedResource := k.deepCopyMap(resource)

	// Get or create metadata
	metadata, ok := modifiedResource[constants.FieldMetadata].(map[string]interface{})
	if !ok {
		metadata = make(map[string]interface{})
		modifiedResource[constants.FieldMetadata] = metadata
	}

	// Set namespace from project
	namespace := k.getProjectNamespace(project)
	if namespace != "" {
		metadata[constants.FieldNamespace] = namespace
	}

	// Add workload ID to labels
	labels, ok := metadata[constants.FieldLabels].(map[string]interface{})
	if !ok {
		labels = make(map[string]interface{})
		metadata[constants.FieldLabels] = labels
	}
	labels[WorkloadLabel] = "true"
	labels[ProjectIdLabel] = project.Id

	// Add workload ID to annotations for additional metadata
	annotations, ok := metadata[constants.FieldAnnotations].(map[string]interface{})
	if !ok {
		annotations = make(map[string]interface{})
		metadata[constants.FieldAnnotations] = annotations
	}
	annotations[WorkloadIdAnnotation] = workloadId
	err := k.updateTopologyAnnotations(ctx, modifiedResource, annotations, project, options)
	if err != nil {
		return nil, err
	}

	return modifiedResource, nil
}

func (k *K8sResourceUtil) updateTopologyAnnotations(
	ctx context.Context,
	resource map[string]interface{},
	annotations map[string]interface{},
	project *openapi.Project,
	options resourceMetadataUpdateOptions,
) error {
	segmentTarget, targetErr := k.pyTorchSegmentAnnotationTarget(resource)
	if targetErr != nil {
		return targetErr
	}
	if segmentTarget != nil {
		return k.applySegmentTopologyDefaults(ctx, resource, annotations, project, segmentTarget, options)
	}
	return applyDefaultTopologyAnnotations(ctx, annotations, project)
}

func applyDefaultTopologyAnnotations(ctx context.Context, annotations map[string]interface{}, project *openapi.Project) error {
	if checkExistTopologyAnnotations(annotations) {
		return nil
	}

	nodePoolsNames := project.GetDefaultNodePools()
	if len(nodePoolsNames) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("project (%s) doesn't have default node pools set, using default node pools", project.Name)
		nodePoolsNames = defaultNodePool
	}

	actualNodePools, getErr := clusters.Get().GetNodePoolsInternalByCluster(ctx, project.ClusterId)
	if getErr != nil {
		return getErr
	}

	if actualNodePools == nil || len(actualNodePools.GetNodepools()) == 0 {
		return fmt.Errorf("no node pools found for cluster %s", project.ClusterId)
	}

	found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools.GetNodepools(), nodePoolsNames)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	networkTopology, err := topology.GetByName(ctx, clusters.Get(), topologyName)
	if err != nil {
		return err
	}

	annotations[topology.KAITopologyAnnotationKey] = topologyName
	annotations[topology.KAITopologyAnnotationPreferredKey] = networkTopology.Levels[len(networkTopology.Levels)-1]
	return nil
}

func checkExistTopologyAnnotations(annotations map[string]interface{}) bool {
	_, found := annotations[topology.KAITopologyAnnotationKey]
	if found {
		return true
	}
	_, found = annotations[topology.KAITopologyAnnotationPreferredKey]
	if found {
		return true
	}
	_, found = annotations[topology.KAITopologyAnnotationRequiredKey]
	return found
}

func (k *K8sResourceUtil) GetGVK(resource map[string]any) (api.Gvk, error) {
	if resource == nil {
		return api.Gvk{}, fmt.Errorf("resource cannot be nil")
	}

	apiVersion, _ := resource[constants.FieldAPIVersion].(string)
	kind, _ := resource[constants.FieldKind].(string)

	group, version := k.parseGroupVersion(apiVersion)

	return api.Gvk{
		Group:   convutils.Ref(group),
		Version: version,
		Kind:    kind,
	}, nil
}

// GetResourceInfo extracts basic information from a Kubernetes resource
func (k *K8sResourceUtil) GetResourceInfo(resource map[string]any) (*ResourceInfo, error) {
	if resource == nil {
		return nil, fmt.Errorf("resource cannot be nil")
	}

	gvk, err := k.GetGVK(resource)
	if err != nil {
		return nil, err
	}

	apiVersion, _ := resource[constants.FieldAPIVersion].(string)

	metadata, _ := resource[constants.FieldMetadata].(map[string]interface{})
	name, _ := metadata[constants.FieldName].(string)
	namespace, _ := metadata[constants.FieldNamespace].(string)

	// Convert labels and annotations to string maps
	labels := k.convertToStringMap(metadata[constants.FieldLabels])
	annotations := k.convertToStringMap(metadata[constants.FieldAnnotations])

	return &ResourceInfo{
		APIVersion:  apiVersion,
		Kind:        gvk.Kind,
		Name:        name,
		Namespace:   namespace,
		Group:       convutils.Deref(gvk.Group, ""),
		Version:     gvk.Version,
		Labels:      labels,
		Annotations: annotations,
	}, nil
}

// getProjectNamespace determines the namespace for a project
func (k *K8sResourceUtil) getProjectNamespace(project *openapi.Project) string {
	// Use requested namespace if available
	if project.HasRequestedNamespace() {
		requestedNamespace := project.GetRequestedNamespace()
		if requestedNamespace != "" {
			return requestedNamespace
		}
	}

	if project.Status.Namespace.IsSet() && project.Status.Namespace.Get() != nil {
		return *project.Status.Namespace.Get()
	}

	return fmt.Sprintf("%s%s", RunaiNamespacePrefix, project.GetName())
}

// parseGroupVersion parses the group and version from an apiVersion string
func (k *K8sResourceUtil) parseGroupVersion(apiVersion string) (group, version string) {
	if apiVersion == "" {
		return "", ""
	}

	parts := strings.Split(apiVersion, "/")
	if len(parts) == 1 {
		// Core API (e.g., "v1")
		return "", parts[0]
	} else if len(parts) == 2 {
		// Group API (e.g., "apps/v1")
		return parts[0], parts[1]
	}

	// Invalid format, return as-is
	return "", apiVersion
}

// convertToStringMap converts a map[string]interface{} to map[string]string
func (k *K8sResourceUtil) convertToStringMap(input interface{}) map[string]string {
	if input == nil {
		return nil
	}

	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil
	}

	result := make(map[string]string)
	for key, value := range inputMap {
		if strValue, ok := value.(string); ok {
			result[key] = strValue
		}
	}

	return result
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func (k *K8sResourceUtil) deepCopyMap(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return nil
	}

	copy := make(map[string]interface{})
	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = k.deepCopyMap(v)
		case []interface{}:
			copy[key] = k.deepCopySlice(v)
		default:
			copy[key] = v
		}
	}
	return copy
}

// deepCopySlice creates a deep copy of a []interface{}
func (k *K8sResourceUtil) deepCopySlice(original []interface{}) []interface{} {
	if original == nil {
		return nil
	}

	copy := make([]interface{}, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[i] = k.deepCopyMap(v)
		case []interface{}:
			copy[i] = k.deepCopySlice(v)
		default:
			copy[i] = v
		}
	}
	return copy
}

// FromJSON converts JSON bytes to map[string]interface{}
func (k *K8sResourceUtil) FromJSON(jsonBytes []byte) (map[string]interface{}, error) {
	var resource map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &resource); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return resource, nil
}

// ToJSON converts map[string]interface{} to JSON bytes
func (k *K8sResourceUtil) ToJSON(resource map[string]interface{}) ([]byte, error) {
	if resource == nil {
		return nil, fmt.Errorf("resource cannot be nil")
	}

	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %v", err)
	}
	return jsonBytes, nil
}

// ResourceInfo contains basic information about a Kubernetes resource
type ResourceInfo struct {
	APIVersion  string            `json:"apiVersion"`
	Kind        string            `json:"kind"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Group       string            `json:"group"`
	Version     string            `json:"version"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// HasWorkloadId checks if a resource has a workload ID in its labels
func (r *ResourceInfo) HasWorkloadId() bool {
	if r.Annotations == nil {
		return false
	}
	_, exists := r.Annotations[WorkloadIdAnnotation]
	return exists
}

// GetWorkloadId returns the workload ID from the resource labels
func (r *ResourceInfo) GetWorkloadId() string {
	if r.Annotations == nil {
		return ""
	}
	return r.Annotations[WorkloadIdAnnotation]
}

// ToGvk converts ResourceInfo to api.Gvk
func (r *ResourceInfo) ToGvk() api.Gvk {
	return api.Gvk{
		Group:   convutils.Ref(r.Group),
		Version: r.Version,
		Kind:    r.Kind,
	}
}

// GVKkey returns a string representation of the resource
func (r *ResourceInfo) GVKkey() string {
	if r.Namespace != "" {
		return fmt.Sprintf(ResourceKeyFormat, r.Namespace, r.Name, r.Kind)
	}
	return fmt.Sprintf(ResourceKeyFormatNoNS, r.Name, r.Kind)
}
