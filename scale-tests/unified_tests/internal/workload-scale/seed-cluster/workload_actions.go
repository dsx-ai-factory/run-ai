/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package seed_cluster

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	projects "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/projects/seed_cluster"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	workload_scale "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/workload-scale"
	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"golang.org/x/sync/semaphore"
	"k8s.io/utils/ptr"
)

const (
	SampleReportsFileName = "/tmp/workload_raw_data.json"
	EntityName            = "workload"
)

type WorkloadConfig struct {
	WorkloadType                workloadclient.WorkloadType
	WorkloadCount               int
	ProjectsCount               int
	ControlPlaneQPS             int
	WorkloadCreateTimeOut       time.Duration
	ClusterID                   string
	NodepoolCount               int
	GpuPerWorkload              float64
	ProjectLoadMethod           LoadMethod
	NumberOfProjectsToHeavyLoad int
	MonitorWaitTime             time.Duration
	BatchSize                   int
	WorkloadsCommands           []string
	KwokEnabled                 bool
	GpuProduct                  []string

	// The following fields are relevant for distributed workloads only
	WorkloadWorkerCount  int
	DistributedFramework workloadclient.DistributedFramework

	// Node chaos configuration
	NodeChaosEnabled      bool
	NodeChaosCount        int
	NodeChaosDownDuration time.Duration
	NodeChaosNodeSelector string
	NodeChaosStartDelay   time.Duration
	NodeChaosDelayBetween time.Duration
}

type workloadCommonCreationParams struct {
	Name            string
	ProjectID       string
	ClusterID       string
	Command         string
	Image           string
	ImagePullPolicy workloadclient.ImagePullPolicy
	Compute         *workloadclient.WorkspaceSpecSpecAllOfCompute
	Labels          []workloadclient.Label
}

type WorkloadSampler struct {
	config                   *WorkloadConfig
	clients                  *sampler_actions.Clients
	workload                 sync.Map
	projects                 *projects.ProjectSampler
	sem                      *semaphore.Weighted
	wg                       sync.WaitGroup
	Pods                     []*workload_scale.PodK8sData
	mgr                      ctrl.Manager
	k8sWorkloadsReachedState bool
	managerCancel            context.CancelFunc
	creationDuration         time.Duration
	workloadsPerSecond       float64
	nodeChaosMetrics         *NodeChaosMetrics
	failedCreationCount      int
}

func NewWorkloadSampler(clusterConfig *scale.ClusterConfig, workloadConfig *WorkloadConfig,
	clients *sampler_actions.Clients,
	projects *projects.ProjectSampler) scale.Entity {
	workloadSamplerEntity := &WorkloadSampler{
		config:   workloadConfig,
		clients:  clients,
		workload: sync.Map{},
		projects: projects,
		sem:      semaphore.NewWeighted(int64(clusterConfig.ControlPlaneQPS)),
		Pods:     make([]*workload_scale.PodK8sData, 0),
	}
	ctx, _ := InitCtrlManager(clients.KubeClient, workloadSamplerEntity, projects.GetProjectsPrefix())
	ctx, cancel := context.WithCancel(ctx)
	workloadSamplerEntity.managerCancel = cancel
	return workloadSamplerEntity
}

func (w *WorkloadSampler) Create(ctx context.Context) error {
	if w.config.KwokEnabled {
		fmt.Println(fmt.Sprint("Configuring Kwok to work with workloads"))
		err := w.createKwokStageConfigurationToSimulateWorkloadPodsRun(ctx)
		if err != nil {
			return fmt.Errorf("failed to create kwok stage configuration: %w", err)
		}
	}

	w.startNodeChaos(ctx)

	fmt.Println(fmt.Sprintf("Creating %d Workloads", w.config.WorkloadCount))
	creationStartTime := time.Now()

	activeProjects := w.projects.GetActiveProjects()
	indexMapper := w.initializeIndexMapper(len(activeProjects))
	for i := 1; i <= w.config.WorkloadCount; i++ {
		w.wg.Add(1)
		projectIndex := indexMapper.GetIndex(i)
		projectSampleResponse := activeProjects[projectIndex]
		go w.createWorkload(ctx, i, projectSampleResponse)

	}

	w.wg.Wait()

	w.creationDuration = time.Since(creationStartTime)
	w.workloadsPerSecond = float64(w.config.WorkloadCount) / w.creationDuration.Seconds()
	fmt.Printf("All %d workloads created in: %v (%.2f workloads/sec)\n",
		w.config.WorkloadCount,
		w.creationDuration,
		w.workloadsPerSecond)

	w.failedCreationCount = 0
	var failureMessages []string
	w.workload.Range(func(k, v interface{}) bool {
		data, _ := v.(*workload_scale.WorkloadCreateSampleResponse)
		if !data.SuccessAtCreation {
			w.failedCreationCount++
			if len(failureMessages) < 5 {
				failureMessages = append(failureMessages, data.CreateFailMessage)
			}
		}
		return true
	})

	successCount := w.config.WorkloadCount - w.failedCreationCount
	if w.failedCreationCount > 0 {
		fmt.Printf("[WARNING] %d/%d workloads failed to create, continuing with %d successful workloads\n",
			w.failedCreationCount, w.config.WorkloadCount, successCount)
		fmt.Printf("[ERROR DETAILS] First %d failure messages:\n", len(failureMessages))
		for i, msg := range failureMessages {
			fmt.Printf("  %d: %s\n", i+1, msg)
		}
	}

	return nil
}

func (w *WorkloadSampler) GetName() string {
	return EntityName
}

func (w *WorkloadSampler) Delete(forceDelete bool) {
	filterString := []string{fmt.Sprintf("clusterId==%s", w.config.ClusterID)}
	if forceDelete {
		filterString = []string{}
	}
	countVal, err := w.Count(filterString)
	if err != nil {
		fmt.Printf("Failed to count workloads: %v\n", err)
		return
	}
	fmt.Printf("Going to Delete: %d Workloads\n", countVal)
	for i := 0; i < int(countVal); i += w.config.BatchSize {
		workloads, err := w.getWorkloads(filterString, i)
		if err != nil {
			fmt.Printf("Failed to get workloads: %v", err)
			return
		}
		workloadsIds := []string{}
		for _, workload := range workloads {
			workloadsIds = append(workloadsIds, workload.Id)
		}
		if len(workloadsIds) == 0 {
			continue
		}
		request := w.clients.WorkloadServiceClient.WorkloadsBatchAPI.BatchWorkloads(context.Background())
		batch := workloadclient.NewWorkloadBatch(workloadsIds, workloadclient.ACTION_DELETE)
		request = request.WorkloadBatch(*batch)
		request.Execute()
	}
	w.workload.Clear()
}

func (w *WorkloadSampler) Count(filter []string) (int64, error) {
	countWorkloadsRequest := w.clients.WorkloadServiceClient.WorkloadsAPI.CountWorkloads(context.Background())
	if len(filter) > 0 {
		countWorkloadsRequest = countWorkloadsRequest.FilterBy(filter)
	}
	count, _, err := countWorkloadsRequest.Execute()
	if err != nil {
		fmt.Printf("Failed to count workloads: %v\n", err)
		return 0, err
	}
	if count == nil {
		return 0, fmt.Errorf("failed to count workloads: count is nil")
	}
	return count.Count, nil
}

func (w *WorkloadSampler) NumberOfSuccessfulCreatedObjects() int {
	count := 0
	w.workload.Range(func(k, v interface{}) bool {
		data, _ := v.(*workload_scale.WorkloadCreateSampleResponse)
		if data.SuccessAtCreation {
			count++
		}
		return true
	})
	return count
}

func (w *WorkloadSampler) Monitor() error {
	retryNum := 20
	fmt.Printf("Monitoring %d Workloads, workload timeout:[%s]\n", w.config.WorkloadCount, w.config.WorkloadCreateTimeOut)
	start := time.Now()
	ticker := time.NewTicker(w.config.MonitorWaitTime)
	defer ticker.Stop()
	timeout := time.After(w.config.WorkloadCreateTimeOut)
	defer w.managerCancel()

	for {
		select {
		case <-ticker.C:
			if !w.k8sWorkloadsReachedState {
				expectedState, err := w.allK8sWorkloadsInClusterInExpectedPhase(context.Background())
				if err != nil {
					return err
				}
				if expectedState {
					fmt.Println(fmt.Sprintf("K8s cluster workload are in expected state , after %s", time.Since(start).String()))
					w.k8sWorkloadsReachedState = true
				}
			} else {
				if w.k8sWorkloadsReachedState {
					expectedState, err := w.allWorkloadsInExpectedState()
					if err != nil {
						if retryNum > 0 {
							fmt.Println(fmt.Sprintf("Failed to check CP cluster workload readiness, retry num:%d, err: %v, will retry", retryNum, err))
							retryNum--
							continue
						} else {
							return err
						}
					}
					if expectedState {
						fmt.Println(fmt.Printf("All CP workloads are in expected state, after %s", time.Since(start).String()))
						return nil
					}
				}
			}
		case <-timeout:
			return fmt.Errorf("workload timeout has reached not all workloads reached expected state")
		}
	}
}

func (w *WorkloadSampler) GetSampleReportGenerator() scale.SampledEntity {
	return w
}

func (w *WorkloadSampler) GenerateSamplesReport() (*scale.SamplesReport, error) {
	success, err := w.allWorkloadsInExpectedState()
	if err != nil {
		return nil, err
	}
	generator := NewWorkloadSamplesReportsGenerator(w, success)
	report, err := generator.GenerateRawDataReport()
	if err != nil {
		return nil, err
	}
	report.UpdateMetaData(*w.config)

	chaosMetrics := w.GetNodeChaosMetrics()
	for k, v := range chaosMetrics.ToMetadataMap() {
		report.MetaData[k] = v
	}

	report.MetaData["workload_creation_failed_count"] = fmt.Sprintf("%d", w.failedCreationCount)
	report.MetaData["workload_creation_success_count"] = fmt.Sprintf("%d", w.config.WorkloadCount-w.failedCreationCount)

	return report, nil

}

func (w *WorkloadSampler) GetSamplesReportFileName() string {
	return SampleReportsFileName
}

func (w *WorkloadSampler) allWorkloadsInExpectedState() (bool, error) {
	clusterIdFilterString := fmt.Sprintf("clusterId==%s", w.config.ClusterID)
	for i := 0; i < w.config.WorkloadCount; i += w.config.BatchSize {
		workloads, err := w.getWorkloads([]string{clusterIdFilterString}, i)
		if err != nil {
			return false, err
		}
		for _, workload := range workloads {
			workloadInExpectedPhase := w.isWorkloadInExpectedPhase(workload.Name, string(workload.GetPhase()))
			if workloadInExpectedPhase == nil {
				continue
			}
			if !*workloadInExpectedPhase {
				return false, nil
			}
		}
	}
	return true, nil
}

func (w *WorkloadSampler) createWorkload(ctx context.Context, i int, projectSampleResponse *projects.ProjectResponse) {
	defer w.wg.Done()
	w.sem.Acquire(ctx, 1)
	defer w.sem.Release(1)

	projectName := projectSampleResponse.Name
	workloadName := fmt.Sprintf("workload-%d", i)
	workloadCommand := w.getCommand(i)

	var workloadID string
	var resp *http.Response
	var err error

	if w.config.NodeChaosEnabled {
		workloadID, resp, err = w.requestWorkloadCreationWithRetry(ctx, workloadName, projectSampleResponse, workloadCommand, i)
	} else {
		workloadID, resp, err = w.requestWorkloadCreation(ctx, workloadName, projectSampleResponse, workloadCommand, i)
	}

	response := workload_scale.WorkloadCreateSampleResponse{
		Name:             workloadName,
		ProjectName:      projectName,
		ExpectedEndPhase: w.getWorkloadExpectedEndPhase(workloadCommand),
	}

	var body string
	if resp != nil {
		body = sampler_actions.BestEffortRead(resp.Body)
	}
	if err != nil {
		response.SuccessAtCreation = false
		if resp == nil {
			response.CreateFailMessage = fmt.Sprintf("failed to create workload-%d: %v", i, err)
		} else {
			response.CreateFailMessage = fmt.Sprintf("failed to create workload-%d: %s: %v", i, body, err)
		}
	} else if resp.StatusCode != http.StatusAccepted {
		response.SuccessAtCreation = false
		response.CreateFailMessage = fmt.Sprintf("failed to create workload-%d, status code: %d: %s", i, resp.StatusCode, body)
	} else {
		response.SuccessAtCreation = true
		response.Id = workloadID
		response.CreateTime = scale.GetTimeInUtc(ptr.To(time.Now()))
	}

	w.workload.Store(workloadName, &response)
}

func (w *WorkloadSampler) requestWorkloadCreationWithRetry(ctx context.Context, workloadName string, projectSampleResponse *projects.ProjectResponse, workloadCommand string, i int) (string, *http.Response, error) {
	const retryInterval = 5 * time.Second
	const maxDuration = 1 * time.Hour
	startTime := time.Now()

	for {
		workloadID, resp, err := w.requestWorkloadCreation(ctx, workloadName, projectSampleResponse, workloadCommand, i)

		if !w.shouldRetryWorkloadCreation(resp, err) {
			return workloadID, resp, err
		}

		if time.Since(startTime) >= maxDuration {
			fmt.Printf("[Retry] Workload %s creation timed out after %s\n", workloadName, maxDuration)
			return workloadID, resp, err
		}

		fmt.Printf("[Retry] Workload %s creation failed (status=%d), retrying in %s\n",
			workloadName, w.getStatusCode(resp), retryInterval)
		time.Sleep(retryInterval)
	}
}

func (w *WorkloadSampler) shouldRetryWorkloadCreation(resp *http.Response, err error) bool {
	if err != nil && resp == nil {
		return true
	}
	if resp != nil && resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return true
	}
	return false
}

func (w *WorkloadSampler) getStatusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func (w *WorkloadSampler) requestWorkloadCreation(ctx context.Context, workloadName string, projectSampleResponse *projects.ProjectResponse, workloadCommand string, i int) (respWorkloadID string, resp *http.Response, err error) {
	var image string
	if w.config.KwokEnabled {
		image = "ubuntu:22.04"
	} else {
		image = "ubuntu"
	}
	p := workloadCommonCreationParams{
		Name:            workloadName,
		ProjectID:       projectSampleResponse.Id,
		ClusterID:       w.config.ClusterID,
		Command:         workloadCommand,
		Image:           image,
		ImagePullPolicy: workloadclient.IMAGEPULLPOLICY_IF_NOT_PRESENT,
		Compute: &workloadclient.WorkspaceSpecSpecAllOfCompute{
			CpuMemoryRequest:  *workloadclient.NewNullableString(ptr.To("1M")),
			GpuDevicesRequest: *workloadclient.NewNullableInt32(ptr.To(int32(w.config.GpuPerWorkload))),
			CpuCoreRequest:    *workloadclient.NewNullableFloat64(ptr.To(float64(0.05))),
		},
		Labels: []workloadclient.Label{
			{
				Name:  *workloadclient.NewNullableString(ptr.To("command")),
				Value: *workloadclient.NewNullableString(ptr.To(strings.ReplaceAll(workloadCommand, " ", ""))),
			},
		},
	}

	switch w.config.WorkloadType {
	case workloadclient.WORKLOADTYPE_TRAINING:
		return w.requestTrainingWorkloadCreation(ctx, p, i)
	case workloadclient.WORKLOADTYPE_DISTRIBUTED:
		return w.requestDistributedWorkloadCreation(ctx, p, i)
	default:
		return "", nil, fmt.Errorf("workload type %s is not supported", w.config.WorkloadType)
	}
}

func (w *WorkloadSampler) requestTrainingWorkloadCreation(ctx context.Context, p workloadCommonCreationParams, i int) (respWorkloadID string, resp *http.Response, err error) {
	creationRequest := &workloadclient.TrainingCreationRequest{
		Name:      p.Name,
		ProjectId: p.ProjectID,
		ClusterId: p.ClusterID,
		Spec: &workloadclient.TrainingSpecSpec{
			Command:         *workloadclient.NewNullableString(ptr.To(p.Command)),
			Image:           *workloadclient.NewNullableString(ptr.To(p.Image)),
			ImagePullPolicy: *workloadclient.NewNullableImagePullPolicy(ptr.To(p.ImagePullPolicy)),
			Compute:         *workloadclient.NewNullableWorkspaceSpecSpecAllOfCompute(p.Compute),
			Labels:          p.Labels,
		},
	}
	creationRequest.Spec.NodePools, creationRequest.Spec.Tolerations = w.generateWorkloadNodePoolConfig(i)
	tw, resp, err := w.clients.WorkloadServiceClient.TrainingsAPI.CreateTraining(ctx).TrainingCreationRequest(*creationRequest).Execute()
	if err != nil {
		return "", resp, err
	}

	return tw.WorkloadId, resp, nil
}

func (w *WorkloadSampler) requestDistributedWorkloadCreation(ctx context.Context, p workloadCommonCreationParams, i int) (respWorkloadID string, resp *http.Response, err error) {
	creationRequest := &workloadclient.DistributedCreationRequest{
		Name:      p.Name,
		ProjectId: p.ProjectID,
		ClusterId: p.ClusterID,
		Spec: &workloadclient.DistributedSpecSpec{
			NumWorkers:           *workloadclient.NewNullableInt32(ptr.To(int32(w.config.WorkloadWorkerCount))),
			DistributedFramework: *workloadclient.NewNullableDistributedFramework(ptr.To(w.config.DistributedFramework)),
			Command:              *workloadclient.NewNullableString(ptr.To(p.Command)),
			Image:                *workloadclient.NewNullableString(ptr.To(p.Image)),
			ImagePullPolicy:      *workloadclient.NewNullableImagePullPolicy(ptr.To(p.ImagePullPolicy)),
			Compute:              *workloadclient.NewNullableWorkspaceSpecSpecAllOfCompute(p.Compute),
			Labels:               p.Labels,
		},
	}
	creationRequest.Spec.NodePools, creationRequest.Spec.Tolerations = w.generateWorkloadNodePoolConfig(i)
	dw, resp, err := w.clients.WorkloadServiceClient.DistributedAPI.CreateDistributed(ctx).DistributedCreationRequest(*creationRequest).Execute()
	if err != nil {
		return "", resp, err
	}

	return dw.WorkloadId, resp, nil
}

func (w *WorkloadSampler) initializeIndexMapper(rangeLen int) sampler_actions.IndexMapper {
	if w.config.ProjectLoadMethod == Uniform {
		return sampler_actions.NewUniformIndexMapper(rangeLen)
	}
	return sampler_actions.NewHeavyIndexMapper(rangeLen, w.config.NumberOfProjectsToHeavyLoad)
}

func (w *WorkloadSampler) getWorkloadExpectedEndPhase(command string) workloadclient.Phase {
	if command == "sleep infinity" {
		return workloadclient.PHASE_RUNNING
	}
	return workloadclient.PHASE_COMPLETED
}

func (w *WorkloadSampler) generateWorkloadNodePoolConfig(i int) (nodepools []string, tolerations []workloadclient.Toleration) {
	if w.config.NodepoolCount == 0 {
		return
	}

	nodepools = []string{fmt.Sprintf("nodepool-%d", i%w.config.NodepoolCount+1)}
	if !w.config.KwokEnabled && len(w.config.GpuProduct) > 0 {
		nodepools = []string{fmt.Sprintf("nodepool-%d", i%len(w.config.GpuProduct)+1)}
	}
	tolerations = []workloadclient.Toleration{
		{
			Key:      *workloadclient.NewNullableString(ptr.To("kwok.x-k8s.io/node")),
			Value:    *workloadclient.NewNullableString(ptr.To("fake")),
			Effect:   *workloadclient.NewNullableTolerationEffect(ptr.To(workloadclient.TOLERATIONEFFECT_NO_SCHEDULE)),
			Operator: *workloadclient.NewNullableTolerationOperator(ptr.To(workloadclient.TOLERATIONOPERATOR_EQUAL)),
		},
	}

	return
}

func (w *WorkloadSampler) getCommand(i int) string {
	return w.config.WorkloadsCommands[i%len(w.config.WorkloadsCommands)]
}

func (w *WorkloadSampler) getWorkloads(filter []string, i int) ([]workloadclient.Workload, error) {
	getWorkloadsRequest := w.clients.WorkloadServiceClient.WorkloadsAPI.GetWorkloads(context.Background())
	if len(filter) > 0 {
		getWorkloadsRequest = getWorkloadsRequest.FilterBy(filter)
	}
	getWorkloadsRequest = getWorkloadsRequest.Limit(int32(w.config.BatchSize))
	getWorkloadsRequest = getWorkloadsRequest.Offset(int32(i))
	workloads, _, err := getWorkloadsRequest.Execute()

	if err != nil {
		return nil, err
	}
	if workloads == nil {
		return nil, fmt.Errorf("failed to get workloads: workloads is nil")
	}
	return workloads.Workloads, nil
}

func (w *WorkloadSampler) createKwokStageConfigurationToSimulateWorkloadPodsRun(ctx context.Context) error {
	for _, command := range w.config.WorkloadsCommands {
		if command == "sleep infinity" {
			continue
		}
		commandForStage := strings.ReplaceAll(command, " ", "")
		millisString := strings.ReplaceAll(commandForStage, "sleep", "")
		millisValue, _ := strconv.Atoi(millisString)
		stage := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "kwok.x-k8s.io/v1alpha1",
				"kind":       "Stage",
				"metadata": map[string]interface{}{
					"name": fmt.Sprintf("dynamic-pod-%s", commandForStage),
				},
				"spec": map[string]interface{}{
					"resourceRef": map[string]interface{}{
						"apiGroup": "v1",
						"kind":     "Pod",
					},
					"delay": map[string]interface{}{
						"durationMilliseconds": millisValue * 1000,
					},
					"selector": map[string]interface{}{
						"matchLabels": map[string]string{
							"command": commandForStage,
						},
						"matchExpressions": []map[string]interface{}{
							{
								"key":      ".status.phase",
								"operator": "In",
								"values":   []string{"Running"},
							},
						},
					},
					"next": map[string]interface{}{
						"statusTemplate": `{{ $now := Now }}
{{ $root := . }}
containerStatuses:
{{ range $index, $item := .spec.containers }}
{{ $origin := index $root.status.containerStatuses $index }}
- image: {{ $item.image | Quote }}
  name: {{ $item.name | Quote }}
  ready: false
  restartCount: 0
  started: false
  state:
    terminated:
      exitCode: 0
      finishedAt: {{ $now | Quote }}
      reason: Completed
      startedAt: {{ $now | Quote }}
{{ end }}
phase: Succeeded
`,
					},
				},
			},
		}
		err := sampler_actions.ApplyUnstructured(ctx, w.clients.DynamicClient, schema.GroupVersionResource{
			Group:    "kwok.x-k8s.io",
			Version:  "v1alpha1",
			Resource: "stages",
		}, stage)
		if err != nil {
			return fmt.Errorf("failed to apply kwok simulate-pods-running: %w", err)
		}

	}
	return nil
}

func (w *WorkloadSampler) allK8sWorkloadsInClusterInExpectedPhase(ctx context.Context) (bool, error) {
	for _, project := range w.projects.GetActiveProjects() {
		namespace := sampler_actions.ProjectToNs(project.Name)
		isInExpectedPhase, err := w.allK8sWorkloadsInNamespaceInExpectedPhase(ctx, namespace)
		if err != nil {
			return false, err
		}
		if !isInExpectedPhase {
			return false, nil
		}
	}
	return true, nil
}

func (w *WorkloadSampler) allK8sWorkloadsInNamespaceInExpectedPhase(ctx context.Context, namespace string) (bool, error) {
	trainingWorkloads, err := sampler_actions.GetK8sTrainingWorkloadFromCluster(ctx, namespace, w.clients.KubeClient)
	if err != nil {
		return false, err
	}
	distributedWorkloads, err := sampler_actions.GetK8sDistributedWorkloadFromCluster(ctx, namespace, w.clients.KubeClient)
	if err != nil {
		return false, err
	}
	for _, workload := range trainingWorkloads {
		workloadInExpectedPhase := w.isWorkloadInExpectedPhase(workload.Name, workload.Status.Phase)
		if workloadInExpectedPhase == nil {
			continue
		}
		if !*workloadInExpectedPhase {
			return false, nil
		}
	}
	for _, workload := range distributedWorkloads {
		workloadInExpectedPhase := w.isWorkloadInExpectedPhase(workload.Name, workload.Status.Phase)
		if workloadInExpectedPhase == nil {
			continue
		}
		if !*workloadInExpectedPhase {
			return false, nil
		}
	}
	return true, nil
}

func (w *WorkloadSampler) isWorkloadInExpectedPhase(workloadName, workloadPhase string) *bool {
	data, ok := w.workload.Load(workloadName)
	if !ok {
		return nil
	}
	workloadCreateSampleResponse := data.(*workload_scale.WorkloadCreateSampleResponse)
	if !workloadCreateSampleResponse.SuccessAtCreation {
		return nil
	}
	if workloadPhase == string(workloadCreateSampleResponse.ExpectedEndPhase) || workloadPhase == string(workloadclient.PHASE_FAILED) {
		return ptr.To(true)
	}
	return ptr.To(false)
}
