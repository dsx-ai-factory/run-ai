package testing

import (
	"context"
	"net/http"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	TestTenant               = "1"
	TestUser                 = "user@run.ai"
	FullResourcesDeservedGpu = float64(2)
)

func getUpdateResources() []api.ResourcesNullable {
	return []api.ResourcesNullable{
		{
			NodePool: &api.ResourcesNodePool{
				Id: "1",
				// no name on purpose - testing that it's not required
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:           ptr.To(float64(2)),
				Deserved:        float64(2),
				OverQuotaWeight: ptr.To(float64(1)),
			},
			Cpu: &api.ResourcesCpu{
				Limit:           ptr.To(float64(1000)),
				Deserved:        ptr.To(float64(1000)),
				OverQuotaWeight: ptr.To(float64(0)),
			},
			Memory: &api.ResourcesMemory{
				Limit:           ptr.To(float64(2000)),
				Deserved:        ptr.To(float64(2000)),
				OverQuotaWeight: ptr.To(float64(3)),
				Units:           api.UNITS_MIB,
			},
		},
		{
			NodePool: &api.ResourcesNodePool{
				Id: "2",
			},
			Gpu: api.ResourcesNullableGpu{
				Limit:           ptr.To(float64(0)),
				Deserved:        float64(0),
				OverQuotaWeight: ptr.To(float64(1)),
			},
		},
	}
}

func createDefaultApiResources(fullResources bool) []api.ResourcesNullable {
	return []api.ResourcesNullable{CreateNodePoolApiResources(fullResources, "1", "default")}
}

func CreateNodePoolApiResources(fullResources bool, nodePoolId string, nodePoolName string) api.ResourcesNullable {
	gpuLimit := float64(2)
	gpuDeserved := FullResourcesDeservedGpu
	gpuOqWeight := float64(2)
	nodePool := api.ResourcesNodePool{
		Id:   nodePoolId,
		Name: nodePoolName,
	}
	resources := api.ResourcesNullable{
		NodePool: &nodePool,
		Gpu: api.ResourcesNullableGpu{
			Limit:           &gpuLimit,
			Deserved:        gpuDeserved,
			OverQuotaWeight: &gpuOqWeight,
		},
	}
	if fullResources {
		limit := float64(2)
		deserved := FullResourcesDeservedGpu
		oqWeight := float64(2)
		units := api.UNITS_GB
		resources.Cpu = &api.ResourcesCpu{
			Limit:           &limit,
			Deserved:        &deserved,
			OverQuotaWeight: &oqWeight,
		}
		resources.Memory = &api.ResourcesMemory{
			Limit:           &limit,
			Deserved:        &deserved,
			OverQuotaWeight: &oqWeight,
			Units:           units,
		}
	}
	return resources
}

func createDefaultApiJobConfig() api.SchedulingRules {
	timeLimits := int32(1000)
	return api.SchedulingRules{
		InteractiveJobTimeLimitSeconds:           &timeLimits,
		InteractiveJobMaxIdleDurationSeconds:     &timeLimits,
		InteractiveJobPreemptIdleDurationSeconds: &timeLimits,
		TrainingJobMaxIdleDurationSeconds:        &timeLimits,
	}
}

func CleanTables() {
	db := dao_utils.DB()
	db.Where("1 = 1").Delete(&models.Resources{})
	db.Where("1 = 1").Delete(&models.Department{})
	db.Where("1 = 1").Delete(&models.Project{})
}

func SetContext(tenantId string, subjectId string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = addAuditToCtx(ctx)
	return ctx
}

func addAuditToCtx(ctx context.Context) context.Context {
	req, _ := http.NewRequest(http.MethodPost, "http://whatever.com", nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.1.1.1"

	al := audit.NewAuditLog().WithRequest(req).WithCtx(ctx)
	return audit.CtxWithAudit(ctx, al)
}
