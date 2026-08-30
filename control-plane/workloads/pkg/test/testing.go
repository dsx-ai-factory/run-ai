package testing

import (
	"context"

	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"k8s.io/utils/ptr"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	TestTenant           = "1"
	TestTenantId         = 1
	TestTenant2          = "2"
	ManagementTestTenant = "10"
	TestUser             = "user@run.ai"
	AnotherUser          = "another@run.ai"

	TestDefaultClusterVersion                = "9.9.9"
	TestClusterVersionForOwnershipProtection = "2.18.0"
	TestClusterVersionForUpdateInference     = "2.19.0"
	TestClusterVersionNativeSubmission       = "2.26.0"

	// Make sure that TestUuid1 < TestUuid2 < TestUuid3
	TestUuid1            = "2e4483c9-1ccc-4079-b9d6-943a93e1077b"
	TestUuid2            = "527eb033-11ca-4ce8-abd3-d36ea3c8f682"
	TestUuid3            = "8fa03b86-61a0-4f70-ba57-872903e5e8fa"
	TestUuid4            = "95427d89-56c5-42e3-834b-27c6ed040521"
	TestUuid5            = "96427d89-56c5-42e3-834b-27c6ed040521"
	TestUuid6            = "a6007d89-56c5-42e3-834b-27c6ed040521"
	TestUuid7            = "aa007d89-56c5-42e3-834b-27c6ed040521"
	ClusterUuid1         = "aea4a846-77fe-4631-9707-2d3b083e83a7"
	ClusterUuid2         = "99876b74-6c33-417c-bb94-7e112a7d4bad"
	ClusterUuid3         = "f9f1eb20-230d-49f7-b809-2fb89abd3417"
	ClusterUuid4         = "f9f1eb20-230d-49f7-b809-2fb89abd3418"
	All                  = "1 = 1"
	TrainCategory        = "Train"
	KnativeDependency    = "knative"
	MPIDependency        = "mpi"
	PyTorchDependency    = "pytorch"
	JaxDependency        = "jax"
	TensorflowDependency = "tensorflow"
	XGBoostDependency    = "xgboost"
	LWSDependency        = "lws"

	DistributedSuspendComponent = "suspend"
)

var Categories = map[string]string{
	"Workspace":   "Build",
	"Training":    "Train",
	"Distributed": "Train",
	"Inference":   "Deploy",
}

func SetContext(tenantId string, subjectId string, groups []string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, groups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{tenantId}})

	return ctx
}

func SetSystemContext(subjectId string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, ManagementTestTenant)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, []string{"Super-Admins"})
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey,
		&authorization_api.PermittedScopes{Tenants: []string{ManagementTestTenant}, System: ptr.To(true)})

	return ctx
}

func SetProjectContext(tenantId string, subjectId string, groups []string, projectId string) context.Context {
	ctx := SetContext(tenantId, subjectId, groups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{projectId}})

	return ctx
}

func SetProjectOnlyContext(tenantId string, subjectId string, groups []string, projectId string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, groups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{projectId}})

	return ctx
}

func SetContextWithAccept(tenantId string, subjectId string, groups []string, accept string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
	ctx = context.WithValue(ctx, runaictx.CfGroups, groups)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{tenantId}})
	ctx = context.WithValue(ctx, runaictx.CfAccept, accept)

	return ctx
}

func CreateTables() {
	config.Init()
	dao_utils.Init(config.Database())
}

func CleanTables() {
	gormDb := dao_utils.DB()
	gormDb.Unscoped().Delete(&models.PodElement{}, All)
	gormDb.Unscoped().Delete(&models.Container{}, All)
	gormDb.Unscoped().Delete(&models.Pod{}, All)
	gormDb.Unscoped().Delete(&models.WorkloadChild{}, All)
	gormDb.Unscoped().Delete(&models.WorkloadElement{}, All)
	gormDb.Unscoped().Delete(&models.Event{}, All)
	gormDb.Unscoped().Delete(&models.Workload{}, All)
	gormDb.Unscoped().Delete(&models.Revision{}, All)
	gormDb.Unscoped().Delete(&models.RevisionChild{}, All)
	gormDb.Unscoped().Delete(&models.ApplicationCrudSync{}, All)
	gormDb.Unscoped().Delete(&models.Application{}, All)

	gormDb.Unscoped().Delete(&models_crud.Workload{}, All)
	gormDb.Unscoped().Delete(&models_crud.WorkloadSyncUpdate{}, All)
	gormDb.Unscoped().Delete(&models_crud.Template{}, All)
	gormDb.Unscoped().Delete(&models_crud.TemplateAssets{}, All)
	gormDb.Unscoped().Delete(&models_crud.PolicyComplianceRecord{}, All)
	gormDb.Unscoped().Delete(&models_crud.Migrations{}, All)
}

func GetDefaultClustersCacheConfig() *checker.ClustersCacheConfig {
	return &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}
}

func BuildDisplayedCluster(clusterId string, version string, protection *bool) openapi.DisplayedCluster {
	return BuildDisplayedClusterWithCustomFrameworks(clusterId, version, protection, getAllFrameworksAvailableDependencies())
}

func BuildDisplayedClusterWithCustomFrameworks(clusterId string, version string, protection *bool, frameworksDependencies map[string]openapi.ClusterDependencyStatus) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)

	displayedStatus := openapi.NewClusterDisplayedStatus()

	dependenciesStatus := openapi.NewClusterDependenciesStatusWithDefaults()
	dependenciesStatus.SetOptional(frameworksDependencies)
	displayedStatus.SetDependencies(*dependenciesStatus)

	if protection != nil {
		clusterConfig := openapi.NewClusterReportedConfig()
		clusterConfig.SetWorkloadOwnershipProtection(*protection)
		displayedStatus.SetConfig(*clusterConfig)
	}

	result.SetStatus(*displayedStatus)

	return result
}

func BuildDisplayedClusterWithKnative(clusterId string, version string, available bool, scaleToZero bool, nodeType bool) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)

	displayedStatus := openapi.NewClusterDisplayedStatus()

	dependenciesStatus := openapi.NewClusterDependenciesStatusWithDefaults()
	dependenciesStatus.SetOptional(map[string]openapi.ClusterDependencyStatus{
		"knative": {
			Available: available,
			Components: &map[string]openapi.ClusterDependencyStatus{
				"knative": {
					Available: available,
					Components: &map[string]openapi.ClusterDependencyStatus{
						"enable-scale-to-zero": {
							Available: scaleToZero,
						},
						"kubernetes.podspec-nodeselector": {
							Available: nodeType,
						},
					},
				},
			},
		},
	})
	displayedStatus.SetDependencies(*dependenciesStatus)

	result.SetStatus(*displayedStatus)

	return result
}

func BuildDisplayedClusterWithSubdomain(clusterId string, version string, subdomainEnabled bool) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)

	displayedStatus := openapi.NewClusterDisplayedStatus()

	dependenciesStatus := openapi.NewClusterDependenciesStatusWithDefaults()
	dependenciesStatus.SetOptional(getAllFrameworksAvailableDependencies())
	displayedStatus.SetDependencies(*dependenciesStatus)

	clusterConfig := openapi.NewClusterReportedConfig()
	clusterConfig.SetSubdomainEnabled(subdomainEnabled)
	displayedStatus.SetConfig(*clusterConfig)

	result.SetStatus(*displayedStatus)

	return result
}

func BuildDisplayedClusterWithLWS(clusterId string, version string, lwsAvailable bool) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)

	displayedStatus := openapi.NewClusterDisplayedStatus()

	dependenciesStatus := openapi.NewClusterDependenciesStatusWithDefaults()
	dependenciesStatus.SetOptional(map[string]openapi.ClusterDependencyStatus{
		LWSDependency: {
			Available: lwsAvailable,
		},
	})
	displayedStatus.SetDependencies(*dependenciesStatus)

	result.SetStatus(*displayedStatus)

	return result
}

func GetCustomFrameworkDependencies(
	mpiAvailable, mpiSuspendable,
	pytorchAvailable, pytorchSuspendable,
	jaxAvailable, jaxSuspendable,
	tfAvailable, tfSuspendable,
	xgboostAvailable, xgboostSuspendable bool,
) map[string]openapi.ClusterDependencyStatus {
	return map[string]openapi.ClusterDependencyStatus{
		LWSDependency: {
			Available: true,
		},
		KnativeDependency: {
			Available: true,
		},
		MPIDependency: {
			Available: mpiAvailable,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: mpiSuspendable,
				},
			},
		},
		PyTorchDependency: {
			Available: pytorchAvailable,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: pytorchSuspendable,
				},
			},
		},
		JaxDependency: {
			Available: jaxAvailable,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: jaxSuspendable,
				},
			},
		},
		TensorflowDependency: {
			Available: tfAvailable,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: tfSuspendable,
				},
			},
		},
		XGBoostDependency: {
			Available: xgboostAvailable,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: xgboostSuspendable,
				},
			},
		},
	}
}

func getAllFrameworksAvailableDependencies() map[string]openapi.ClusterDependencyStatus {
	return GetCustomFrameworkDependencies(true, true, true, true, true, true, true, true, true, true)
}
