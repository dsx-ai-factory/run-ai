/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pkg

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/controller"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/validation"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RuntimeRules struct {
	K8sClient         client.WithWatch             // the k8's client with watch
	QueuesManager     *queuesManager.QueuesManager // the queues manager
	PrometheusQuerier prometheus.Querier           // Prometheus client
	Recorder          record.EventRecorder         // K8S event Recorder
	EventList         []int
	WorkloadValidator *validation.WorkloadValidator
	SupportedJobKinds map[string]client.Object
}

func NewRuntimeRules(k8sClient client.WithWatch, qm *queuesManager.QueuesManager, prometheusClient prometheus.Querier,
	recorder record.EventRecorder, eventList []int, workloadValidator *validation.WorkloadValidator) *RuntimeRules {
	// Copy controller.DistributedJobCRDsSupportedMap rather than aliasing it —
	// the package-level map would otherwise pick up additions from every
	// constructor call (notably the WorkloadRunner entry below, which would leak
	// across test cases).
	supportedJobKinds := make(map[string]client.Object, len(controller.DistributedJobCRDsSupportedMap)+2)
	for k, v := range controller.DistributedJobCRDsSupportedMap {
		supportedJobKinds[k] = v
	}
	supportedJobKinds[constants.KindRunaiJob] = &v1.RunaiJob{}

	supportedJobKinds[constants.KindWorkloadRunner] = &dromosv1alpha1.WorkloadRunner{}

	return &RuntimeRules{
		K8sClient:         k8sClient,
		QueuesManager:     qm,
		PrometheusQuerier: prometheusClient,
		Recorder:          recorder,
		EventList:         eventList,
		WorkloadValidator: workloadValidator,
		SupportedJobKinds: supportedJobKinds,
	}
}

func (rr *RuntimeRules) Start(ctx context.Context) {
	zap.S().Infof("Starting runtime-rules")
	workloadCache := util.WorkloadCache{}

	zap.S().Infof("Initializing queue manager with cluster data (projects + workloads)")
	rr.InitQueueManager(ctx, &workloadCache)

	zap.S().Infof("Starting watch clients")
	rr.startWatchClients(ctx, &workloadCache)
}

func (rr *RuntimeRules) InitQueueManager(ctx context.Context, workloadCache *util.WorkloadCache) {
	// list all cluster existing projects
	projectList, err := kubeConfig.ListProjects(ctx, rr.K8sClient)
	if err != nil {
		zap.S().Errorf("Error while list Projects  %+v \n", err)
		return
	}

	// for each project add it to the queue manager
	// and push all the project workloads to the project queue
	for i := range projectList.Items {
		project := &projectList.Items[i]
		if !util.ProjectHasLimits(project) {
			// in case of no time limit rules, skip
			zap.S().Debugf("Project %s has no time limit rules, skip ", project.Name)
			continue
		}
		projectQueue := rr.QueuesManager.Add(ctx, workloadCache, rr.K8sClient, rr.PrometheusQuerier, project, rr.Recorder, rr.EventList, rr.WorkloadValidator)
		if projectQueue == nil {
			zap.S().Errorf("Error while adding project (%s) to the queue manager", project.Name)
			return
		}

		err := queuesManager.PushProjectWorkloadsToProjectQueue(ctx, workloadCache, rr.K8sClient, rr.PrometheusQuerier, project.Name, project.Status.Namespace, projectQueue, rr.SupportedJobKinds)
		if err != nil {
			zap.S().Errorf("Error while adding project workloads to the project queue  %+v \n", err)
			return
		}
	}
}

func (rr *RuntimeRules) startWatchClients(ctx context.Context, workloadCache *util.WorkloadCache) {
	zap.S().Debug("Starting watch clients \n")

	workloadHandler := handlers.NewWorkloadHandler(rr.K8sClient, rr.QueuesManager, rr.PrometheusQuerier, workloadCache, rr.SupportedJobKinds)
	projectController := controllers.NewProjectController(
		rr.K8sClient,
		rr.K8sClient,
		rr.QueuesManager,
		rr.PrometheusQuerier,
		workloadCache,
		rr.Recorder,
		rr.EventList,
		rr.WorkloadValidator,
		rr.SupportedJobKinds,
	)
	autoDeleteCompletedController := controllers.NewAutoDeleteCompletedController(ctx, rr.K8sClient, rr.Recorder, rr.SupportedJobKinds)

	stopper := make(chan struct{})
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer func() {
		close(stopper)
		cancel()
	}()
	defer runtime.HandleCrash()

	if _, ok := rr.SupportedJobKinds[constants.KindWorkloadRunner]; ok {
		workloadRunnerController := controllers.NewWorkloadController(rr.K8sClient, workloadHandler, constants.KindWorkloadRunner)
		go workloadRunnerController.Start(ctxWithCancel, stopper)
	}

	go projectController.Start(ctxWithCancel, stopper)
	go autoDeleteCompletedController.Start(ctxWithCancel, stopper)

	<-stopper
}
