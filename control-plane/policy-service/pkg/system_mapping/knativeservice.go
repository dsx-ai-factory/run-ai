/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemServingKnativeDevServiceV1Mapping is the system seed for
// serving.knative.dev/Service/v1 (spec selectors without scope).
//
// Knative Service wraps a standard corev1.PodSpec under .spec.template.spec
// (via RevisionSpec inline embedding). Pod/container/metadata selectors match
// apps/Deployment/v1 paths; there is no Deployment-style selector field.
func BuildSystemServingKnativeDevServiceV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("serving.knative.dev"), Kind: "Service", Version: "v1"}, buildServingKnativeDevServiceV1SpecSelectors()
}

func buildServingKnativeDevServiceV1SpecSelectors() map[string]api.SpecSelector {
	containerArgsInfo := map[string]api.MappingFieldType{
		"containerName": api.MAPPINGFIELDTYPE_STRING,
	}

	return map[string]api.SpecSelector{
		"workloadMetadata": {
			Description: convutils.Ref("the Knative Service metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		"podSpec": {
			Description: convutils.Ref("the pod spec under .spec.template.spec"),
			Query: `(
			  .spec.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec.template.spec",
			},
		},

		"podMetadata": {
			Description: convutils.Ref("the revision template metadata (annotations and labels) under .spec.template.metadata"),
			Query: `(
			  .spec.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec.template.metadata",
			},
		},

		"allContainers": {
			Description: convutils.Ref("all app and init containers"),
			Query: `(
			  .spec.template.spec.containers[]?,
			  .spec.template.spec.initContainers[]?
			)
			| select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers"),
			Query: `(
			  .spec.template.spec.initContainers[]?
			)
			| select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers"),
			Query: `(
			  .spec.template.spec.containers[]?
			)
			| select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"appContainerByName": {
			Description: convutils.Ref("app container with a given name"),
			Query: `(
			  .spec.template.spec.containers[]?
			)
			| select(.name == $ARGS.named.containerName)`,
			ArgsInfo: containerArgsInfo,
			FieldGroups: []string{
				SystemPodSpecContainerFgName,
				SystemPodSpecAndContainerFgName,
			},
		},

		"initContainerByName": {
			Description: convutils.Ref("init container with a given name"),
			Query: `(
			  (.spec.template.spec.initContainers // [])[]
			)
			| select(.name == $ARGS.named.containerName)`,
			ArgsInfo: containerArgsInfo,
			FieldGroups: []string{
				SystemPodSpecContainerFgName,
				SystemPodSpecAndContainerFgName,
			},
		},
	}
}
