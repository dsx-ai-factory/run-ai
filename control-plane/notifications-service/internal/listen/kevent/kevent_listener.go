/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kevent

import (
	"context"
	"regexp"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/spf13/viper"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/util"
	"github.com/rs/zerolog/log"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	maxEventAge  = 10
	ListenerType = "kube_events"
)

type KubeEventListenerConfig struct {
	RelevantObjects    []RelevantObjectConfig `mapstructure:"relevant_objects" json:"relevant_objects"`
	RelevantNamespaces []string               `mapstructure:"relevant_namespaces" json:"relevant_namespaces"`
}

func NewKubeEventListenerConfig(v *viper.Viper) *KubeEventListenerConfig {
	kubeEventsDefaults(v)
	cfg := KubeEventListenerConfig{}
	if err := v.Unmarshal(&cfg, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed unmarshalling configuration of route %q", v.GetString("Id"))
	}
	return &cfg
}

func kubeEventsDefaults(v *viper.Viper) {
	config.SetDefault(v, "relevant_objects", []RelevantObjectConfig{
		{
			Kind: "PodGroup",
			Reasons: []string{
				"Pending",
				"Running",
				"Suspended",
				"Succeeded",
				"Failed",
				"Error",
				"StartError",
				"CrashLoopBackOff",
				"UnexpectedAdmissionError",
				"CreateContainerConfigError",
				"Timeout",
			},
		},
	})
	config.SetDefault(v, "listener.kube_events.relevant_namespaces", []string{
		"runai-.*",
	})
}

type RelevantObjectConfig struct {
	Kind    string   `mapstructure:"kind" json:"kind"`
	Reasons []string `mapstructure:"reasons" json:"reasons"`
}

type KubeEventListener struct {
	id         string
	config     *KubeEventListenerConfig
	kubeClient kubernetes.Interface
	eventCh    chan *events.Event
}

func NewKubeEventListener(id string, eventCh chan *events.Event, kubeClient kubernetes.Interface, config *KubeEventListenerConfig) *KubeEventListener {
	ctrl.SetLogger(zap.New())

	return &KubeEventListener{
		id:         id,
		config:     config,
		kubeClient: kubeClient,
		eventCh:    eventCh,
	}
}

func (kel *KubeEventListener) Id() string {
	return kel.id
}

func (kel *KubeEventListener) Listen(ctx context.Context) {
	// Generating map for faster lookup
	relevantEventObjectKindReasonsMap := map[string]map[string]struct{}{}
	for _, relevantObject := range kel.config.RelevantObjects {
		relevantEventObjectKindReasonsMap[relevantObject.Kind] = util.ListToSet(relevantObject.Reasons)
	}
	relevantNamespacesRegex := util.AggregateRegexList(kel.config.RelevantNamespaces)

	factory := informers.NewSharedInformerFactoryWithOptions(kel.kubeClient, 0)
	eventInformer := factory.Events().V1().Events().Informer()

	eventHandler := kel.getFilteringResourceEventHandler(relevantNamespacesRegex, relevantEventObjectKindReasonsMap)
	_, err := eventInformer.AddEventHandler(eventHandler)
	if err != nil {
		log.Err(err).Msg("failed to add event handler for listener")
		return
	}

	eventInformer.Run(ctx.Done())
}

func (kel *KubeEventListener) getFilteringResourceEventHandler(relevantNamespacesRegex *regexp.Regexp, relevantEventObjectKindReasonsMap map[string]map[string]struct{}) cache.FilteringResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: getFilterFunc(relevantNamespacesRegex, relevantEventObjectKindReasonsMap),
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				kubeEvent := obj.(*eventsv1.Event)
				event := mapEventToServiceEvent(kubeEvent)
				kel.eventCh <- event
			},
		},
	}
}

func getFilterFunc(relevantNamespacesRegex *regexp.Regexp, relevantEventObjectKindReasonsMap map[string]map[string]struct{}) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		switch event := obj.(type) {
		case *eventsv1.Event:
			if _, ok := relevantEventObjectKindReasonsMap[event.Regarding.Kind]; !ok {
				return false
			}
			if _, ok := relevantEventObjectKindReasonsMap[event.Regarding.Kind][event.Reason]; !ok {
				return false
			}
			if !relevantNamespacesRegex.MatchString(event.Regarding.Namespace) {
				return false
			}
			if !eventAgeValid(event) {
				return false
			}
			return true
		default:
			return false
		}
	}
}

func mapEventToServiceEvent(kubeEvent *eventsv1.Event) *events.Event {
	event := events.NewEvent()
	event.UID = events.UID(kubeEvent.UID)
	event.RawEvent = kubeEvent

	event.Properties["source"] = "kubernetes"
	event.Properties["regarding"] = map[string]interface{}{
		"kind":      kubeEvent.Regarding.Kind,
		"name":      kubeEvent.Regarding.Name,
		"namespace": kubeEvent.Regarding.Namespace,
	}
	event.Properties["reason"] = kubeEvent.Reason
	event.Properties["note"] = kubeEvent.Note

	return event
}

func eventAgeValid(event *eventsv1.Event) bool {
	timestamp := event.CreationTimestamp.Time
	if timestamp.IsZero() {
		timestamp = event.EventTime.Time
	}
	eventAge := time.Since(timestamp)
	return eventAge <= time.Duration(maxEventAge*time.Second)
}
