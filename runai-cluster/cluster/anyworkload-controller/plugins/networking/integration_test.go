/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"reflect"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type recordingFieldIndexer struct {
	fields []string
}

func (i *recordingFieldIndexer) IndexField(
	_ context.Context,
	_ client.Object,
	field string,
	_ client.IndexerFunc,
) error {
	i.fields = append(i.fields, field)
	return nil
}

func TestIntegrationSetupConfiguresOnlyNetworking(t *testing.T) {
	scheme := newTestScheme()
	cachedReader := fake.NewClientBuilder().WithScheme(scheme).Build()
	liveReader := fake.NewClientBuilder().WithScheme(scheme).Build()
	integration := &Integration{plugin: newPlugin(cachedReader, nil)}
	indexer := &recordingFieldIndexer{}

	if err := integration.Setup(context.Background(), liveReader, indexer); err != nil {
		t.Fatalf("setup networking integration: %v", err)
	}
	if integration.plugin.reader != cachedReader {
		t.Fatal("setup replaced the plugin's cached reader")
	}
	if integration.plugin.ownershipReader != liveReader {
		t.Fatal("setup did not inject the live ownership reader")
	}
	if integration.plugin.templatesReader != liveReader {
		t.Fatal("setup did not inject the live templates reader")
	}
	wantFields := []string{
		workloadRunnerNetworkingDependencyIndex,
		endpointSlicePodUIDIndex,
	}
	if !reflect.DeepEqual(indexer.fields, wantFields) {
		t.Fatalf("registered fields = %v, want %v", indexer.fields, wantFields)
	}
}

func TestIntegrationSetupFailsClosedBeforeBuild(t *testing.T) {
	integration := NewIntegration()
	if err := integration.Setup(context.Background(), nil, nil); err == nil {
		t.Fatal("setup without a built plugin must fail")
	}
}

func TestDisabledIntegrationSetupIsNoOp(t *testing.T) {
	integration := &Integration{disabled: true}
	if err := integration.Setup(context.Background(), nil, nil); err != nil {
		t.Fatalf("disabled networking setup returned an error: %v", err)
	}
}
