package services

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func HandleProjectDeleteRequest(ctx context.Context, event *events.ProjectDeleteRequest) error {
	logging.Infof("Handling project deleted request: %v", event)
	if event.LeaveOrphans {
		return nil
	}

	err := dao.DeleteWorkloadsByScope(ctx, api.SCOPETYPE_PROJECT, event.ProjectId)
	if err != nil {
		return err
	}
	err = dao.DeleteAssetsByScope(ctx, api.SCOPETYPE_PROJECT, event.ProjectId)
	return err

}

func HandleClusterDeleteEvent(ctx context.Context, event *events.ClusterDelete) error {
	logging.Infof("Handling cluster deleted request: %v", event)
	err := dao.DeleteWorkloadsByScope(ctx, api.SCOPETYPE_CLUSTER, event.ClusterId)
	if err != nil {
		return err
	}
	err = dao.DeleteAssetsByScope(ctx, api.SCOPETYPE_CLUSTER, event.ClusterId)
	return err
}

func HandleTenantDeleteEvent(ctx context.Context, event *events.TenantDelete) error {
	logging.Infof("Handling tenant deleted request: %v", event)
	// The tenant ID comes from the event, not the context,
	// since a tenant deletion request wouldn't originate from within the tenant itself.
	err := dao.DeleteWorkloadsByScope(ctx, api.SCOPETYPE_TENANT, event.TenantId)
	if err != nil {
		return err
	}
	err = dao.DeleteAssetsByScope(ctx, api.SCOPETYPE_TENANT, event.TenantId)
	return err
}

func HandlePVCEvents(ctx context.Context, pvcEvents *events.PvcEvents) error {
	tenantId := auth.TenantIdOf(ctx)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Received %d PVC events to handle", len(pvcEvents.Events))

	claimNames := make([]string, 0, len(pvcEvents.Events))
	for _, event := range pvcEvents.Events {
		claimNames = append(claimNames, event.ClaimName)
	}

	// The events coming from the k8s-objects-tacker contain only the scope type and scope id, and claim name.
	// We need to list the assets by claim names to get the asset ids in order to create an event for the asset.
	pvcAssets, err := dao.ListPvcAssetsByClaimNames(ctx, nil, claimNames)
	if err != nil {
		return fmt.Errorf("failed to list pvc assets by claim names: %v", err)
	}

	pvcAssetsByClaimName := make(map[string][]*api.PvcAsset)
	for _, pvcAsset := range pvcAssets {
		if pvcAsset.Spec.ClaimName == nil {
			continue
		}
		if _, found := pvcAssetsByClaimName[*pvcAsset.Spec.ClaimName]; !found {
			pvcAssetsByClaimName[*pvcAsset.Spec.ClaimName] = make([]*api.PvcAsset, 0)
		}

		pvcAssetsByClaimName[*pvcAsset.Spec.ClaimName] = append(pvcAssetsByClaimName[*pvcAsset.Spec.ClaimName], &pvcAsset)
	}

	eventModels := make([]*dbmodel.Event, 0, len(pvcAssets))
	for _, event := range pvcEvents.Events {
		assetId := ""

		// When handling PVC events, the same claim name can exist in multiple scopes
		// (for example, the same PVC name might exist in different clusters, projects, or namespaces).
		// This means that matching by claim name alone is not enough to find the correct asset.
		// To connect the event to the right asset, we also need to check the scope type and scope id
		// from the event. The logic below first checks if there's exactly one asset with the given claim name.
		// If there are multiple assets with the same name, it goes through each one and compares their
		// scope type and scope id with the event. This way we can tell them apart and find the correct match.
		if len(pvcAssetsByClaimName[event.ClaimName]) == 1 {
			assetId = pvcAssetsByClaimName[event.ClaimName][0].Meta.Id
		} else {
			for _, pvcAsset := range pvcAssetsByClaimName[event.ClaimName] {
				scopeInfo := ScopeInfoOfMeta(&pvcAsset.Meta)
				if string(scopeInfo.Scope) == string(event.ScopeType) && scopeInfo.ScopeId == event.ScopeId {
					assetId = pvcAsset.Meta.Id
					break
				}
			}
			if assetId == "" {
				logging.Warnf("Failed to find asset id for pvc event with claim name: %s", event.ClaimName)
				continue
			}
		}

		eventModel := dbmodel.NewEventModelFromPVCEvent(tenantId, assetId, &event)
		eventModels = append(eventModels, &eventModel)
	}

	if len(eventModels) != 0 {
		dbErr := dao.UpsertEvents(ctx, eventModels)
		if dbErr != nil {
			return fmt.Errorf("failed to upsert events: %v", dbErr)
		}
	}

	return nil
}
