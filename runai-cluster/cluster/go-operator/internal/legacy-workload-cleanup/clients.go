/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"fmt"

	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Clients struct {
	ManagementClient        client.Client
	WorkloadClient          client.Client
	WorkloadDynamic         dynamic.Interface
	ManagementAuthorization authorizationclient.AuthorizationV1Interface
	WorkloadAuthorization   authorizationclient.AuthorizationV1Interface
}

func NewClientsFromConfig(ctx context.Context, base *rest.Config) (*Clients, error) {
	managementClient, managementDynamic, managementAuthorization, err := newClusterClients(base)
	if err != nil {
		return nil, fmt.Errorf("build management cluster clients: %w", err)
	}

	config := &runaiv1.RunaiConfig{}
	key := client.ObjectKey{Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName}
	if err := managementClient.Get(ctx, key, config); err != nil {
		return nil, fmt.Errorf("get RunaiConfig %s: %w", key, err)
	}

	remoteURL := ""
	if config.Spec.Internal != nil && config.Spec.Internal.Global != nil {
		remoteURL = config.Spec.Internal.Global.NMC.RemoteAPIServerURL
	}
	if remoteURL == "" {
		return &Clients{
			ManagementClient:        managementClient,
			WorkloadClient:          managementClient,
			WorkloadDynamic:         managementDynamic,
			ManagementAuthorization: managementAuthorization,
			WorkloadAuthorization:   managementAuthorization,
		}, nil
	}

	workloadClient, workloadDynamic, workloadAuthorization, err := newClusterClients(remoteConfigForURL(base, remoteURL))
	if err != nil {
		return nil, fmt.Errorf("build workload cluster clients: %w", err)
	}
	return &Clients{
		ManagementClient:        managementClient,
		WorkloadClient:          workloadClient,
		WorkloadDynamic:         workloadDynamic,
		ManagementAuthorization: managementAuthorization,
		WorkloadAuthorization:   workloadAuthorization,
	}, nil
}

func newClusterClients(config *rest.Config) (client.Client, dynamic.Interface, authorizationclient.AuthorizationV1Interface, error) {
	typedClient, err := client.New(config, client.Options{Scheme: schemeinit.Scheme()})
	if err != nil {
		return nil, nil, nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	return typedClient, dynamicClient, clientset.AuthorizationV1(), nil
}

func remoteConfigForURL(base *rest.Config, remoteURL string) *rest.Config {
	remote := rest.CopyConfig(base)
	remote.Host = remoteURL
	return remote
}

func RunCommand(ctx context.Context) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("load in-cluster Kubernetes configuration: %w", err)
	}
	clients, err := NewClientsFromConfig(ctx, config)
	if err != nil {
		return err
	}
	return NewRunner(
		clients.ManagementClient,
		clients.WorkloadClient,
		clients.WorkloadDynamic,
		clients.ManagementAuthorization,
		clients.WorkloadAuthorization,
	).Run(ctx)
}
