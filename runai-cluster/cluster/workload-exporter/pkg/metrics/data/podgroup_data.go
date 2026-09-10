/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/cluster_relations"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	kaiconsts "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type TopOwnerMetadata struct {
	Name    string `yaml:"name"`
	UID     string `yaml:"uid"`
	Group   string `yaml:"group"`
	Version string `yaml:"version"`
	Kind    string `yaml:"kind"`
}

func GetTopOwnerMetadata(podGroup *kaiv2alpha2.PodGroup) (*TopOwnerMetadata, error) {
	topOwnerMetadataString, found := podGroup.Annotations[kaiconsts.TopOwnerMetadataKey]
	if found {
		return unmarshallMetadata(topOwnerMetadataString)
	}
	topOwnerMetadataString, found = podGroup.Annotations[cluster_relations.TopOwnerAnnotationNameBC]
	if found {
		return unmarshallMetadata(topOwnerMetadataString)
	}

	return nil, fmt.Errorf("top owner annotation not found for podgroup %s/%s",
		podGroup.Namespace, podGroup.Name)
}

func GetProjectNameForPodGroup(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, readerClient client.Reader) (string, error) {
	queue, err := getQueueOfPodGroup(ctx, podGroup, readerClient)
	if err != nil {
		return "", err
	}

	projectName := getProjectNameOfQueueFromLabels(queue.Labels)
	if projectName == "" {
		return "", fmt.Errorf("could not find %s label on queue %s", constants.RunaiProjectLabel, queue.Name)
	}
	return projectName, nil
}

func getQueueOfPodGroup(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, readerClient client.Reader) (*kaiv2.Queue, error) {
	return getQueue(ctx, podGroup.Spec.Queue, readerClient)
}

func getQueue(ctx context.Context, queueName string, readerClient client.Reader) (*kaiv2.Queue, error) {
	var queue kaiv2.Queue
	err := readerClient.Get(ctx, client.ObjectKey{Name: queueName}, &queue)
	if err != nil {
		return nil, fmt.Errorf("could not get queue %s: %w", queueName, err)
	}

	return &queue, nil
}

func getProjectNameOfQueueFromLabels(labels map[string]string) string {
	if projectName, found := labels[constants.RunaiProjectLabel]; found {
		return projectName
	}
	return ""
}

func unmarshallMetadata(metadataString string) (*TopOwnerMetadata, error) {
	var metadata TopOwnerMetadata
	err := yaml.Unmarshal([]byte(metadataString), &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}
