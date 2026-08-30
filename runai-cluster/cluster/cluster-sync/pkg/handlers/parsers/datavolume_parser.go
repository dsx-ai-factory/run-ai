package parsers

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/utils/ptr"

	swagger "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

var dvCrdPhaseToApiPhase = map[v1alpha1.DataVolumePhase]swagger.DatavolumePhase{
	v1alpha1.DataVolumeInitializing: swagger.DATAVOLUMEPHASE_CREATING,
	v1alpha1.DataVolumeReady:        swagger.DATAVOLUMEPHASE_READY,
	v1alpha1.DataVolumeNotReady:     swagger.DATAVOLUMEPHASE_NOT_READY,
	v1alpha1.DataVolumeDeleting:     swagger.DATAVOLUMEPHASE_DELETING,
	"":                              swagger.DATAVOLUMEPHASE_UNKNOWN,
}

func ParseDatavolumeToUpdateStatusRequest(datavolumeStatus *v1alpha1.DataVolumeStatus) swagger.DatavolumeStatusInternal {
	return swagger.DatavolumeStatusInternal{
		Phase:             getApiPhase(datavolumeStatus.Phase),
		PhaseMessage:      getStrPtrOrNil(datavolumeStatus.Message),
		DatavolumePvcName: getStrPtrOrNil(datavolumeStatus.DatavolumePvcName),
		DatavolumePvName:  getStrPtrOrNil(datavolumeStatus.DatavolumePvName),
	}
}

func ParseDatavolumeToDeleteStatusRequest() swagger.DatavolumeStatusInternal {
	return swagger.DatavolumeStatusInternal{
		// the phase is disregarded in the delete request
		Phase:     swagger.DATAVOLUMEPHASE_READY,
		IsDeleted: *swagger.NewNullableBool(ptr.To(true)),
	}
}

func ParseDatavolumesToStatusesForSync(datavolumes []v1alpha1.DataVolume) []swagger.DatavolumeStatusForSync {
	datavolumeStatuses := []swagger.DatavolumeStatusForSync{}
	for _, dv := range datavolumes {
		datavolumeStatuses = append(datavolumeStatuses, swagger.DatavolumeStatusForSync{
			Phase:             getApiPhase(dv.Status.Phase),
			PhaseMessage:      getStrPtrOrNil(dv.Status.Message),
			DatavolumePvcName: getStrPtrOrNil(dv.Status.DatavolumePvcName),
			DatavolumePvName:  getStrPtrOrNil(dv.Status.DatavolumePvName),
			DatavolumeId:      dv.Spec.DatavolumeId,
		})
	}
	return datavolumeStatuses
}

func getApiPhase(phase v1alpha1.DataVolumePhase) swagger.DatavolumePhase {
	apiPhase, ok := dvCrdPhaseToApiPhase[phase]
	if !ok {
		apiPhase = swagger.DATAVOLUMEPHASE_UNKNOWN
	}
	return apiPhase
}

func getStrPtrOrNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}
