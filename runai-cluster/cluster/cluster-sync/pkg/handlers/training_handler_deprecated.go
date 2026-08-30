package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/antihax/optional"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//TODO: delete when needed

type TrainingHandlerDeprecated struct {
	cachedReader       client.Reader
	runaiBackendClient *backendclient.BackendClient
}

func NewTrainingHandlerDeprecated(cachedReader client.Reader, runaiBackendClient *backendclient.BackendClient) *TrainingHandlerDeprecated {
	return &TrainingHandlerDeprecated{
		cachedReader:       cachedReader,
		runaiBackendClient: runaiBackendClient,
	}
}

func (handler *TrainingHandlerDeprecated) HandleUpdateItem(key client.ObjectKey) error {
	tw := &v2alpha1.TrainingWorkload{}
	err := handler.cachedReader.Get(context.Background(), key, tw)
	if err != nil {
		return err
	}

	return handler.HandleUpdate(tw)
}

func (handler *TrainingHandlerDeprecated) HandleAdd(workload *v2alpha1.TrainingWorkload) error {
	jobStatus := workload.Status.DeepCopyToJobStatus()
	statusToSend := parsers.WorkloadToStatusChangeRequest(&workload.ObjectMeta, jobStatus)
	if statusToSend == nil {
		return nil
	}

	res, err := handler.runaiBackendClient.Client.TrainingsApi.UpdateTrainingStatus(
		handler.runaiBackendClient.Context(),
		*statusToSend,
		config.Config().ClusterUuid,
	)
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("training %s not found in control plane, it has probably removed already from the DB", statusToSend.Id)
		return nil
	}
	return handler.runaiBackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Updated training status of %s to %s", statusToSend.Id, statusToSend.Status),
	)
}

func (handler *TrainingHandlerDeprecated) HandleDelete(workload *v2alpha1.TrainingWorkload) error {
	workloadId := workload.Annotations[parsers.AnnotationWorkloadId]
	if workloadId == "" {
		return nil
	}

	res, err := handler.runaiBackendClient.Client.TrainingsApi.DeleteTraining(
		handler.runaiBackendClient.Context(),
		config.Config().ClusterUuid,
		workloadId,
	)
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("training %s is already removed from control plane", workloadId)
		return nil
	}
	return handler.runaiBackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Deleted training %s", workloadId),
	)
}

func (handler *TrainingHandlerDeprecated) HandleUpdate(workload *v2alpha1.TrainingWorkload) error {
	// updating status of training is same as in addition of training
	return handler.HandleAdd(workload)
}

func (handler *TrainingHandlerDeprecated) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start trianing Sync")

	workloadList := &v2alpha1.TrainingWorkloadList{}
	opts := make([]client.ListOption, 0)
	kubeErr := handler.cachedReader.List(context.Background(), workloadList, opts...)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	syncToSend := make([]swagger.WorkloadStatus, 0, len(workloadList.Items))

	for _, workload := range workloadList.Items {
		jobStatus := workload.Status.DeepCopyToJobStatus()
		curStatus := parsers.WorkloadToStatusChangeRequest(&workload.ObjectMeta, jobStatus)
		if curStatus != nil {
			syncToSend = append(syncToSend, *curStatus)
		}
	}

	res, err := handler.runaiBackendClient.Client.TrainingsApi.SyncTrainingsStatus(
		handler.runaiBackendClient.Context(),
		config.Config().ClusterUuid,
		&swagger.TrainingsApiSyncTrainingsStatusOpts{
			Body: optional.NewInterface(syncToSend),
		},
	)

	err = handler.runaiBackendClient.HandleHttpResponse(res, err, "Sync workloads")
	if err != nil {
		zap.S().Errorf("%+v", err)
	}
}
