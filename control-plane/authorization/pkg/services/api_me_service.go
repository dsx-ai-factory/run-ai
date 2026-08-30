package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
)

// MeApiService is a service that implements the logic for the MeApiServicer
type MeApiService struct{}

func NewMeApiService() api.MeAPIServicer {
	return &MeApiService{}
}

func (m *MeApiService) GetMeAccessRules(ctx context.Context, limit int32, offset int32, search string) (api.ImplResponse, error) {
	tenantId, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	accessRules, err := dao.DbGetMeAccessRules(ctx, tenantId, subjectId, groups, limit, offset, search)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	countAccessRules := int32(len(*accessRules))
	accessRulesResponse := api.MeAccessRulesResponse{
		AccessRules: api.MeAccessRulesResponseAccessRules{
			TotalRecords:   countAccessRules,
			DisplayRecords: countAccessRules,
			AccessRules:    []api.MeAccessRule{},
		},
	}
	for _, accessRule := range *accessRules {
		accessRulesResponse.AccessRules.AccessRules = append(accessRulesResponse.AccessRules.AccessRules, AccessRuleDaoToMeApi(&accessRule))
	}
	return api.Response(http.StatusOK, accessRulesResponse), nil
}

func (m *MeApiService) CountMeAccessRules(ctx context.Context, search string) (api.ImplResponse, error) {
	tenantId, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	count, dbErr := dao.DbCountMeAccessRules(ctx, tenantId, subjectId, groups, search)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountAccessRules200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}
