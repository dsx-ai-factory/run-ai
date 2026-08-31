/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_relations

import (
	"context"
	"fmt"
	kaiconsts "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TopOwnerAnnotationNameBC  = "run.ai/top-owner-metadata"
	TopOwnerUidForPodGropPath = ".metadata.ownerreferenses.uid"
)

func GetPodGroupTopOwnerRef(podGroup *kaiv2alpha2.PodGroup) (*metadata.TopOwnerReference, error) {
	topOwnerAnnotation, hasAnnotation := podGroup.Annotations[kaiconsts.TopOwnerMetadataKey]
	if hasAnnotation {
		return unmarshallMetadata(topOwnerAnnotation)
	}
	topOwnerAnnotation, hasAnnotation = podGroup.Annotations[TopOwnerAnnotationNameBC]
	if hasAnnotation {
		return unmarshallMetadata(topOwnerAnnotation)
	}
	return nil, fmt.Errorf("has no top owner")
}

func getAllPodGroupsOfOwnerObj(ctx context.Context, ownerObject *unstructured.Unstructured, kubeClient client.Client) (
	kaiv2alpha2.PodGroupList, error) {
	topOwnerUid := string(ownerObject.GetUID())

	topOwnerPodGroups := kaiv2alpha2.PodGroupList{}
	err := kubeClient.List(ctx, &topOwnerPodGroups, client.MatchingFields{TopOwnerUidForPodGropPath: topOwnerUid})
	if err != nil {
		return kaiv2alpha2.PodGroupList{}, err
	}

	return topOwnerPodGroups, nil
}

func unmarshallMetadata(metadataString string) (*metadata.TopOwnerReference, error) {
	topOwnerRef := metadata.TopOwnerReference{}
	err := yaml.Unmarshal([]byte(metadataString), &topOwnerRef)
	if err != nil {
		return nil, err
	}
	return &topOwnerRef, nil
}
