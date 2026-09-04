/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/runaipublic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	k8sClient client.Client
}

func NewClient(k8sClient client.Client) *Client {
	return &Client{
		k8sClient: k8sClient,
	}
}

func (c *Client) Get(ctx context.Context, extension extensions.Type) (*extensions.Extension, error) {
	handler := runaipublic.NewRunaiPublicHandler(ctx, c.k8sClient)
	enabled, suspendable, err := c.get(handler, extension)
	if err != nil {
		return nil, err
	}
	return createExtensionStatus(extension, enabled, suspendable), nil
}

func (c *Client) get(handler *runaipublic.Handler, extension extensions.Type) (bool, bool, error) {
	enabled, _, suspendable, err := handler.IsDistributedJobAvailable(extension.String())
	if err != nil {
		return false, false, err
	}
	return enabled, suspendable, nil
}

func (c *Client) List(ctx context.Context, extensionsGroup []extensions.Type) (extensions.Group, error) {
	group := make(extensions.Group, len(extensionsGroup))
	handler := runaipublic.NewRunaiPublicHandler(ctx, c.k8sClient)
	for _, extension := range extensionsGroup {
		enabled, suspension, err := c.get(handler, extension)
		if err != nil {
			return nil, err
		}
		group[extension] = *createExtensionStatus(extension, enabled, suspension)
	}
	return group, nil
}

func createExtensionStatus(extension extensions.Type, enabled, suspendable bool) *extensions.Extension {
	capabilities := make(map[extensions.CapabilityType]bool, 1)
	capabilities[Suspension] = suspendable

	return &extensions.Extension{
		Type: extension,
		Status: extensions.Status{
			Enabled:      enabled,
			Capabilities: capabilities,
		},
	}
}
