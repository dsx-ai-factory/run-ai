package services

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	"k8s.io/utils/ptr"
)

func NoContentResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusNoContent,
	}
}

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	apiError.Details = ""
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}

func ErrorToApiError(err error, errorCode int) api.ImplResponse {
	return api.ImplResponse{
		Code: errorCode,
		Body: err.Error(),
	}
}

func validateGetVolumesRequest(requestType api.DatavolumeRequestType, usableInProjectId string) *api.Error {
	switch requestType {
	case api.DATAVOLUMEREQUESTTYPE_ORIGINATED:
		if usableInProjectId != "" {
			return api.NewError(http.StatusBadRequest, "usableInProjectId is not allowed when requestType is Originated", nil)
		}
		return nil
	case api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT:
		if usableInProjectId == "" {
			return api.NewError(http.StatusBadRequest, "usableInProjectId is required when requestType is UsableInProject", nil)
		}
		return nil
	case "":
		return api.NewError(http.StatusBadRequest, "requestType is required", nil)
	}
	return api.NewError(http.StatusBadRequest, fmt.Sprintf("invalid requestType, must be one of: %v, %v", api.DATAVOLUMEREQUESTTYPE_ORIGINATED, api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT), nil)
}

func DatavolumeModelToApi(datavolume models.Datavolume) api.Datavolume {
	apiSharedScopes := []api.SharedScope{}
	if datavolume.SharedScopes != nil {
		apiSharedScopes = convutils.ConvertSliceTypes(datavolume.SharedScopes, sharedScopeModelToApi)
	}

	return api.Datavolume{
		Id:                         datavolume.ID.String(),
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
		ProjectName:                datavolume.ProjectName,
		DepartmentId:               datavolume.DepartmentId,
		ClusterId:                  datavolume.ClusterId.String(),
		CreatedAt:                  datavolume.CreatedAt,
		UpdatedAt:                  datavolume.UpdatedAt,
		CreatedBy:                  datavolume.CreatedBy,
		UpdatedBy:                  datavolume.UpdatedBy,

		SharedScopes: apiSharedScopes,

		Status: DatavolumeStatusModelToApi(datavolume.DatavolumeStatus),
	}
}

func DatavolumeModelToApiNoSharedScopes(datavolume models.Datavolume) api.DatavolumeNoSharedScopes {
	return api.DatavolumeNoSharedScopes{
		Id:                         datavolume.ID.String(),
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
		ProjectName:                datavolume.ProjectName,
		DepartmentId:               datavolume.DepartmentId,
		ClusterId:                  datavolume.ClusterId.String(),
		CreatedAt:                  datavolume.CreatedAt,
		UpdatedAt:                  datavolume.UpdatedAt,
		CreatedBy:                  datavolume.CreatedBy,
		UpdatedBy:                  datavolume.UpdatedBy,

		Status: DatavolumeStatusModelToApi(datavolume.DatavolumeStatus),
	}
}

func DatavolumeViewModelToApiNoSharedScopes(datavolume models.DatavolumeView) api.DatavolumeNoSharedScopes {
	return api.DatavolumeNoSharedScopes{
		Id:                         datavolume.ID.String(),
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
		UseOriginalVolume:          &datavolume.UseOriginalVolume,
		ProjectName:                datavolume.ProjectName,
		DepartmentId:               datavolume.DepartmentId,
		ClusterId:                  datavolume.ClusterId.String(),
		CreatedAt:                  datavolume.CreatedAt,
		UpdatedAt:                  datavolume.UpdatedAt,
		CreatedBy:                  datavolume.CreatedBy,
		UpdatedBy:                  datavolume.UpdatedBy,

		Status: DatavolumeStatusModelToApi(datavolume.DatavolumeStatus),
	}
}

func DatavolumeModelToApiInternal(datavolume models.Datavolume) api.DatavolumeInternal {
	apiSharedScopes := []api.SharedScope{}
	if datavolume.SharedScopes != nil {
		apiSharedScopes = convutils.ConvertSliceTypes(datavolume.SharedScopes, sharedScopeModelToApi)
	}

	return api.DatavolumeInternal{
		Id:                         datavolume.ID.String(),
		Name:                       datavolume.Name,
		Description:                datavolume.Description,
		ProjectId:                  datavolume.ProjectId,
		OriginPvcName:              datavolume.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolume.ShouldDeleteOriginalVolume,
		UseOriginalVolume:          &datavolume.UseOriginalVolume,
		ProjectName:                datavolume.ProjectName,
		DepartmentId:               datavolume.DepartmentId,
		ClusterId:                  datavolume.ClusterId.String(),
		CreatedAt:                  datavolume.CreatedAt,
		UpdatedAt:                  datavolume.UpdatedAt,
		CreatedBy:                  datavolume.CreatedBy,
		UpdatedBy:                  datavolume.UpdatedBy,
		Phase:                      datavolume.Phase,

		SharedScopes: apiSharedScopes,
	}
}

func DatavolumeCreateRequestToModel(datavolumeCreationFields api.DatavolumeCreationFields, tenantId uint,
	subjectId string, clusterId uuid.UUID, projectName string, departmentId string,
) models.Datavolume {
	return models.Datavolume{
		Name:                       datavolumeCreationFields.Name,
		Description:                datavolumeCreationFields.Description,
		ProjectId:                  datavolumeCreationFields.ProjectId,
		OriginPvcName:              datavolumeCreationFields.OriginPvcName,
		ShouldDeleteOriginalVolume: datavolumeCreationFields.ShouldDeleteOriginalVolume,
		UseOriginalVolume:          convutils.Deref(datavolumeCreationFields.UseOriginalVolume, false),

		ProjectName:  projectName,
		DepartmentId: departmentId,
		TenantId:     tenantId,
		ClusterId:    clusterId,

		BaseModel: models.BaseModel{
			CreatedBy: subjectId,
			UpdatedBy: subjectId,
		},

		DatavolumeStatus: models.DatavolumeStatus{
			Phase:        api.DATAVOLUMEPHASE_CREATING,
			PhaseMessage: "Creating datavolume in the cluster",
		},
	}
}

func internalDatavolumeCreateRequestToModel(internalDatavolumeCreationFields api.InternalDatavolumeCreationFields, tenantId uint,
	clusterId uuid.UUID, departmentName string,
) models.Datavolume {
	return models.Datavolume{
		Name: internalDatavolumeCreationFields.Name,

		Scope:          internalDatavolumeCreationFields.Scope,
		DepartmentId:   internalDatavolumeCreationFields.DepartmentId,
		DepartmentName: departmentName,
		ClusterId:      clusterId,
		TenantId:       tenantId,

		ShouldDeleteOriginalVolume: false,
		UseOriginalVolume:          true,
		OriginPvcName:              internalDatavolumeCreationFields.OriginPvcName,
		BaseModel: models.BaseModel{
			CreatedBy: "runai",
			UpdatedBy: "runai",
		},

		DatavolumeStatus: models.DatavolumeStatus{
			Phase:        api.DATAVOLUMEPHASE_CREATING,
			PhaseMessage: "Creating datavolume in the cluster",
		},

		Internal: true,
	}
}

func DatavolumeStatusApiToModel(datavolumeStatus api.DatavolumeStatusInternal) models.DatavolumeStatus {
	return models.DatavolumeStatus{
		Phase:             datavolumeStatus.Phase,
		PhaseMessage:      datavolumeStatus.PhaseMessage,
		Conditions:        convutils.ConvertSliceTypes(datavolumeStatus.Conditions, conditionApiToModel),
		DatavolumePvcName: datavolumeStatus.DatavolumePvcName,
		DatavolumePvName:  datavolumeStatus.DatavolumePvName,
	}
}

func DatavolumeStatusModelToApi(datavolumeStatus models.DatavolumeStatus) api.DatavolumeStatus {
	return api.DatavolumeStatus{
		Phase:             datavolumeStatus.Phase,
		PhaseMessage:      datavolumeStatus.PhaseMessage,
		Conditions:        convutils.ConvertSliceTypes(datavolumeStatus.Conditions, conditionModelToApi),
		DatavolumePvcName: datavolumeStatus.DatavolumePvcName,
		DatavolumePvName:  datavolumeStatus.DatavolumePvName,
	}
}

func conditionApiToModel(condition api.Condition) models.Condition {
	return models.Condition{
		Type:               condition.Type,
		Status:             condition.Status,
		Message:            condition.Message,
		Reason:             condition.Reason,
		LastTransitionTime: condition.LastTransitionTime,
	}
}

func conditionModelToApi(condition models.Condition) api.Condition {
	return api.Condition{
		Type:               condition.Type,
		Status:             condition.Status,
		Message:            condition.Message,
		Reason:             condition.Reason,
		LastTransitionTime: condition.LastTransitionTime,
	}
}

func datavolumeMetaToAncestors(datavolume *models.Datavolume) *authorization_utils.Ancestors {
	return &authorization_utils.Ancestors{
		TenantId:     strconv.FormatUint(uint64(datavolume.TenantId), 10),
		ClusterUuid:  ptr.To(datavolume.ClusterId.String()),
		DepartmentId: &datavolume.DepartmentId,
		ProjectId:    &datavolume.ProjectId,
	}
}

func datavolumeStatusForSyncToModelWithStatus(datavolumeStatus api.DatavolumeStatusForSync) (models.Datavolume, error) {
	id, err := uuid.Parse(datavolumeStatus.DatavolumeId)
	if err != nil {
		return models.Datavolume{}, fmt.Errorf("failed to parse datavolume id <%s>, err: %v",
			datavolumeStatus.DatavolumeId, err)
	}

	return models.Datavolume{
		BaseModel: models.BaseModel{
			ID: id,
		},
		DatavolumeStatus: models.DatavolumeStatus{
			Phase:             datavolumeStatus.Phase,
			PhaseMessage:      datavolumeStatus.PhaseMessage,
			Conditions:        convutils.ConvertSliceTypes(datavolumeStatus.Conditions, conditionApiToModel),
			DatavolumePvcName: datavolumeStatus.DatavolumePvcName,
			DatavolumePvName:  datavolumeStatus.DatavolumePvName,
		},
	}, nil
}

func parseFromToUuids(fromStr string, toStr string) (*uuid.UUID, *uuid.UUID, *api.Error) {
	var from *uuid.UUID = nil
	if fromStr != "" {
		fromField, err := uuid.Parse(fromStr)
		if err != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("From field must be a valid uuid. <%s>", fromStr), err)
		}
		from = &fromField
	}

	var to *uuid.UUID = nil
	if toStr != "" {
		toField, err := uuid.Parse(toStr)
		if err != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("To field must be a valid uuid. <%s>", toStr), err)
		}
		to = &toField
	}

	return from, to, nil
}

func setDefaultLimit(limit int32) int32 {
	if limit == 0 {
		limit = 200
	}
	limit = int32(math.Min(float64(limit), 500))
	limit = int32(math.Max(float64(limit), 1))

	return limit
}

func convertSliceToMap[T comparable](s []T) map[T]bool {
	result := map[T]bool{}
	for _, item := range s {
		result[item] = true
	}
	return result
}
