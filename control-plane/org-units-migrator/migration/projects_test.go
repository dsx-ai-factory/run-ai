package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orgUnitsServiceDao "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	orgUnitsServiceModels "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
)

var _ = Describe("Projects tests", Ordered, func() {
	It("Should get projects grouped", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		for _, tenant := range tenants {
			projects, err := database.GetTenantProjectsV1(tenant.Id)
			Expect(err).To(BeNil())
			Expect(len(projects)).To(Equal(database.TestProjectsCountPerDepartment * database.TestDepartmentsCountPerTenant))
		}
	})
})

func validateProjectsMigration(tenant models.Tenant, setting map[string]bool) {
	projects, err := orgUnitsServiceDao.GetProjects(context.Background(), fmt.Sprintf("%d", tenant.Id), []validators.ParameterizedFilter{}, "", "", 0, 0, "", nil, false, orgUnitsServiceModels.SlimVerbosity)
	Expect(err).To(BeNil())
	Expect(projects).To(Not(BeNil()))
	Expect(len(*projects)).To(Equal(database.TestProjectsCountPerDepartment * database.TestDepartmentsCountPerTenant))
	limit := float64(1.5)
	deserved := float64(1.5)
	oqw := float64(2)
	if setting[database.ProjectManualOverQuota] {
		oqw = deserved
	}
	for _, project := range *projects {
		Expect(project.ID).To(BeNumerically(">=", database.TestProjectsCountPerDepartment*database.TestDepartmentsCountPerTenant*(tenant.Id-1)))
		Expect(project.ID).To(BeNumerically("<=", database.TestProjectsCountPerDepartment*database.TestDepartmentsCountPerTenant*tenant.Id))
		Expect(project.Name).To(Equal(fmt.Sprintf("proj-%d", project.ID)))
		Expect(project.TenantId).To(Equal(fmt.Sprintf("%d", tenant.Id)))
		Expect(project.ClusterId.String()).To(Equal(database.MapTenantIdToClusterId[tenant.Id]))
		Expect(project.CreatedAt).To(Not(Equal(time.Time{})))
		Expect(*project.ParentID).To(BeNumerically(">=", tenant.Id))
		Expect(*project.ParentID).To(BeNumerically("<=", database.TestDepartmentsCountPerTenant*tenant.Id))
		Expect(project.Phase).ToNot(BeNil())
		Expect(*project.Phase).To(Equal("Ready"))
		Expect(project.RequestedNamespace).To(Equal(fmt.Sprintf("ruani-proj-%d", project.ID)))
		Expect(project.CreatedBy).To(Equal("test@run.ai"))
		Expect(project.UpdatedBy).To(Equal("test-2@run.ai"))
		Expect(*project.OrgUnitData.InteractiveJobTimeLimit).To(Equal(int32(1)))
		Expect(project.OrgUnitData.InteractiveJobMaxIdleDuration).To(BeNil())
		Expect(*project.OrgUnitData.InteractiveJobPreemptIdleDuration).To(Equal(int32(3)))
		Expect(*project.OrgUnitData.TrainingJobMaxIdleDuration).To(Equal(int32(4)))
		Expect(*project.OrgUnitData.TrainingJobTimeLimit).To(Equal(int32(5)))
		Expect(project.OrgUnitData.TrainingNodeTypes).NotTo(BeNil())
		Expect(len(*project.OrgUnitData.TrainingNodeTypes)).To(Equal(3))
		var nodeType orgUnitsServiceModels.NodeType
		for i := range *project.OrgUnitData.TrainingNodeTypes {
			if (*project.OrgUnitData.TrainingNodeTypes)[i].ID == int64(tenant.Id*100+1) {
				nodeType = (*project.OrgUnitData.TrainingNodeTypes)[i]
			}
		}
		Expect(nodeType).To(Not(BeNil()))
		Expect((*project.OrgUnitData.TrainingNodeTypes)[0].Name).To(Equal(fmt.Sprintf("node-type-%d", (*project.OrgUnitData.TrainingNodeTypes)[0].ID)))
		Expect(project.OrgUnitData.WorkspaceNodeTypes).To(BeNil())
		Expect(project.DeletedAt.Time).To(Equal(time.Time{}))
		Expect(len(project.OrgUnitData.Resources)).To(Equal(database.TestNodepoolsCountPerTenant))
		for _, resource := range project.OrgUnitData.Resources {
			Expect(resource.CpuDeserved).To(Not(BeNil()))
			Expect(*resource.CpuDeserved).To(Equal(deserved))
			Expect(resource.CpuLimit).To(Not(BeNil()))
			Expect(*resource.CpuLimit).To(Equal(limit))
			Expect(resource.CpuOverQuotaWeight).To(Not(BeNil()))
			Expect(*resource.CpuOverQuotaWeight).To(Equal(oqw))
			Expect(resource.GpuDeserved).To(Not(BeNil()))
			Expect(*resource.GpuDeserved).To(Equal(deserved))
			Expect(resource.GpuLimit).To(Not(BeNil()))
			Expect(*resource.GpuLimit).To(Equal(limit))
			Expect(resource.GpuOverQuotaWeight).To(Not(BeNil()))
			Expect(*resource.GpuOverQuotaWeight).To(Equal(oqw))
			Expect(resource.MemoryDeserved).To(Not(BeNil()))
			Expect(*resource.MemoryDeserved).To(Equal(deserved))
			Expect(resource.MemoryLimit).To(Not(BeNil()))
			Expect(*resource.MemoryLimit).To(Equal(limit))
			Expect(resource.MemoryOverQuotaWeight).To(Not(BeNil()))
			Expect(*resource.MemoryOverQuotaWeight).To(Equal(oqw))

			Expect(resource.NodePoolName).To(Equal(fmt.Sprintf("np-%s", resource.NodePoolId)))
		}
	}
}
