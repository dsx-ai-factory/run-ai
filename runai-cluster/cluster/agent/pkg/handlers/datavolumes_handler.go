package handlers

import (
	"encoding/json"
	"reflect"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"

	"github.com/rs/zerolog/log"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/datavolumes_openapi"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DatavolumesGetLimit = 100
)

type DatavolumesHandler struct {
	*CommonHandler
}

type DatavolumeObjectInspector bool

func NewDatavolumesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		&v1alpha1.DataVolume{}, DatavolumeObjectInspector(true), HandlingOptions{})
	if err != nil {
		return nil, err
	}

	return &DatavolumesHandler{
		CommonHandler: commonHandler,
	}, nil
}

func (h *DatavolumesHandler) Handle() error {
	log.Debug().Msgf("Starting to sync datavolumes")

	backendDatavolumes, err := h.getBackendDatavolumes()
	if err != nil {
		return err
	}

	clusterDatavolumes := []client.Object{}
	hasErrorOccurred := false

	deletingCount := 0
	for _, backendDatavolume := range backendDatavolumes {
		if isBackendDatavolumeDeleting(backendDatavolume) {
			deletingCount++
			continue
		}

		clusterDatavolume, err := convertBackendToClusterDatavolume(&backendDatavolume)
		if err != nil {
			hasErrorOccurred = true
			continue
		}

		clusterDatavolumes = append(clusterDatavolumes, clusterDatavolume)
	}

	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterDatavolumes))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, deletingCount)

	return h.ReconcileObjects(clusterDatavolumes, hasErrorOccurred)
}

func (h *DatavolumesHandler) MetricLabel() string {
	return "data_volumes"
}

func (h *DatavolumesHandler) getBackendDatavolumes() ([]datavolumes_openapi.DatavolumeInternal, error) {
	request := h.backendClient.DatavolumesClient.DatavolumesApi.GetDatavolumesInternal(h.backendClient.Context())
	request = request.ClusterId(config.Config().ClusterUuid)
	request = request.Limit(DatavolumesGetLimit)

	result := []datavolumes_openapi.DatavolumeInternal{}

	var offset int32 = 0

	for {
		log.Debug().Msgf("Getting datavolumes from backend; offset: %d, limit: %d", offset, DatavolumesGetLimit)
		request = request.Offset(offset)

		backendDatavolumesResponse, res, err := h.backendClient.DatavolumesClient.DatavolumesApi.GetDatavolumesInternalExecute(
			request,
		)
		err = h.backendClient.HandleHttpResponse(res, err, "Get Datavolumes")
		if err != nil || backendDatavolumesResponse == nil {
			log.Error().Msgf("Had an error while getting datavolumes: %+v", err)
			return []datavolumes_openapi.DatavolumeInternal{}, err
		}

		result = append(result, backendDatavolumesResponse.Datavolumes...)

		if backendDatavolumesResponse.Next == nil || *backendDatavolumesResponse.Next == 0 {
			break
		}
		offset = *backendDatavolumesResponse.Next
	}

	return result, nil
}

func convertBackendToClusterDatavolume(backendDatavolume *datavolumes_openapi.DatavolumeInternal) (*v1alpha1.DataVolume, error) {
	datavolume, err := convertJsonDatavolumeFields(backendDatavolume)
	if err != nil {
		return nil, err
	}

	datavolume.Name = backendDatavolume.Name
	datavolume.Spec.DatavolumeId = backendDatavolume.Id

	if backendDatavolume.UseOriginalVolume.IsSet() && backendDatavolume.UseOriginalVolume.Get() != nil {
		datavolume.Spec.ShouldUseOriginalVolume = *backendDatavolume.UseOriginalVolume.Get()
	}

	return datavolume, nil
}

func isBackendDatavolumeDeleting(backendDv datavolumes_openapi.DatavolumeInternal) bool {
	return backendDv.Phase != nil && *backendDv.Phase == datavolumes_openapi.DATAVOLUMEPHASE_DELETING
}

/*
convertJsonDatavolumeFields converts backend datavolume to cluster datavolume, using json marshal/unmarshal.
Not all fields are matching in both objects - so other fields needs to be handled manually.
*/
func convertJsonDatavolumeFields(backendDatavolume *datavolumes_openapi.DatavolumeInternal) (*v1alpha1.DataVolume, error) {
	jsonBytes, err := json.Marshal(backendDatavolume)
	if err != nil {
		log.Error().Msgf("Failed to json marshal backend datavolume <%s>, error: %s",
			backendDatavolume.Id, err.Error())
		return nil, err
	}

	crdDatavolume := &v1alpha1.DataVolume{}
	err = json.Unmarshal(jsonBytes, &crdDatavolume.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json un-marshal crd datavolume <%s>, error: %s",
			backendDatavolume.Id, err.Error())
		return nil, err
	}

	return crdDatavolume, nil
}

func (d DatavolumeObjectInspector) DeepCopySpec(from, to client.Object) {
	fromDatavolume, ok := from.(*v1alpha1.DataVolume)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		log.Error().Msgf("Failed to convert from object to datavolume")
		return
	}
	toDatavolume, ok := to.(*v1alpha1.DataVolume)
	if !ok {
		log.Error().Msgf("Failed to convert to object to datavolume")
		return
	}

	fromDatavolume.Spec.DeepCopyInto(&toDatavolume.Spec)

	fromDatavolumeId, found := fromDatavolume.Labels[v1alpha1.DataVolumeIdLabel]
	if !found {
		return
	}

	if toDatavolume.Labels == nil {
		toDatavolume.Labels = map[string]string{}
	}
	toDatavolume.Labels[v1alpha1.DataVolumeIdLabel] = fromDatavolumeId
}

func (d DatavolumeObjectInspector) IsUpToDate(left, right client.Object) bool {
	leftDatavolume, ok := left.(*v1alpha1.DataVolume)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		log.Error().Msgf("Failed to convert left object to datavolume")
		return false
	}
	rightDatavolume, ok := right.(*v1alpha1.DataVolume)
	if !ok {
		log.Error().Msgf("Failed to convert right object to datavolume")
		return false
	}

	leftSpecJson, err := json.Marshal(leftDatavolume.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal datavolume spec, error: %s", err.Error())
		return false
	}
	rightSpecJson, err := json.Marshal(rightDatavolume.Spec)
	if err != nil {
		log.Error().Msgf("Failed to json marshal datavolume spec, error: %s", err.Error())
		return false
	}

	return reflect.DeepEqual(leftSpecJson, rightSpecJson)
}
