package usage_info

import (
	"context"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func PatchUsageTimes(ctx context.Context, projectId int32, workloadType string,
	patchingInfo map[string]response_patching.ResponsePatchingPtrs,
) *api.Error {
	//
	//   some sanity validations
	//
	if patchingInfo == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("patchingInfo nil, cannot patch usage times")
		return nil
	}
	if len(patchingInfo) == 0 {
		return nil
	}

	if projectId <= 0 || workloadType == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("projectId and workloadType are required for usage time patching")
		return nil
	}

	//
	//   make a list of ids for all the assets we want to check
	//
	assetIdsToCheck := make([]string, 0, len(patchingInfo))

	for assetId := range patchingInfo {
		assetIdsToCheck = append(assetIdsToCheck, assetId)
	}

	//
	//   load the usage times of the assets from the DB
	//
	usageTimes, dbErr := dao.ListUsageTimes(ctx, workloadType, projectId, assetIdsToCheck)
	if dbErr != nil {
		return dbErr
	}

	//
	//   patch all the information that was loaded from the DB into the response
	//
	for _, usageTime := range usageTimes {
		patchPtr := patchingInfo[usageTime.AssetId]
		usageTime := usageTime.WorkloadCreationTime

		*patchPtr.UsageTimes = &api.UsageTimesInfo{
			LastUsedByWorkload: &usageTime,
		}
	}

	return nil
}
