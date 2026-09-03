/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kube_client

import (
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	once       sync.Once
	restConfig *rest.Config
	k8sClient  client.Client
)

func InitKubeClient() {
	var err error

	conf := runai_config.Get[*config.AssetsServiceConfig]()
	restConfig, err = rest.InClusterConfig()
	if err == nil {
		logging.Infof("in-cluster kube config loaded successfully")
	} else {
		logging.Warnf("load in-cluster config failed: %v", err)

		if conf.KubeConfig == "" {
			logging.Fatalf("KUBECONFIG variable is not provided or empty")
		} else {
			logging.Infof("loading KUBECONFIG from file: %s", conf.KubeConfig)

			kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				clientcmd.NewDefaultClientConfigLoadingRules(),
				&clientcmd.ConfigOverrides{})

			restConfig, err = kubeConfig.ClientConfig()
			if err != nil {
				logging.Fatalf("%v", err)
			}
		}
	}

	restConfig.QPS = float32(conf.K8sClientConfigQPS)
	restConfig.Burst = conf.K8sClientConfigBurst

	k8sClient, err = client.New(restConfig, client.Options{})
	if err != nil {
		logging.Fatalf("%v", err)
	}
}

func KubeClient() client.Client {
	once.Do(func() {
		if k8sClient == nil {
			InitKubeClient()
		}
	})
	return k8sClient
}

func SetKubeClientForTest(fakeClient client.Client) {
	k8sClient = fakeClient
}
