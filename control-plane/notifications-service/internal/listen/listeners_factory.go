/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package listen

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	http_listener "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/http"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/kevent"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/prom"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/smtp"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type KubeClientGetter interface {
	Get() kubernetes.Interface
}

type ListenerFactory struct {
	kubeClientGetter KubeClientGetter
}

func NewListenerFactory() Factory {
	return &ListenerFactory{
		kubeClientGetter: newKubeClientGetter(),
	}
}

func (lf *ListenerFactory) CreateListener(id string, eventCh chan *events.Event, configProvider ListeningConfigProvider) (Listener, error) {
	switch configProvider.GetListenerType() {
	case kevent.ListenerType:
		kubeClient := lf.kubeClientGetter.Get()
		return kevent.NewKubeEventListener(id, eventCh, kubeClient, configProvider.GetListeningConfig().KubeListener), nil
	case prom.ListenerType:
		return prom.NewPrometheusListener(id, eventCh, configProvider.GetListeningConfig().PrometheusListener), nil
	case http_listener.ListenerType:
		return http_listener.NewHttpListener(id, eventCh, configProvider.GetListeningConfig().HttpListenerConfig), nil
	case smtp.ListenerType:
		return smtp.NewSmtpListener(id, eventCh, configProvider.GetListeningConfig().SmtpListener), nil
	default:
		return nil, fmt.Errorf("listener %s not found", configProvider.GetListenerType())
	}
}

func newKubeClientGetter() *kubeClientGetter {
	return &kubeClientGetter{}
}

type kubeClientGetter struct{}

func (kcg *kubeClientGetter) Get() kubernetes.Interface {
	kubeConfig := ctrl.GetConfigOrDie()
	return kubernetes.NewForConfigOrDie(kubeConfig)
}
