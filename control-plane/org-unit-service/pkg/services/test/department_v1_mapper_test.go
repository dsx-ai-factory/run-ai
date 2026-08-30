package test

import (
	"context"

	clusteropenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	v1 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/v1"
)

var _ = Describe("Departments 1 mapper test", Ordered, func() {
	Context("test department mapper, cpu ft off", func() {
		cpuFt := false
		overrideDeserved := false

		It("map model to response", func() {
			department := createDepartment(false, cpuFt, true)
			api, err := v1.MapDepartmentModelToV1Response(context.Background(), *department, nil, false, map[string]clusteropenapi.NodepoolMinimal{"1": {Name: "default", Id: "1"}, nodePool2Id: {Name: nodePool2Name, Id: nodePool2Id}})
			Expect(err).To(BeNil())
			assertV1ApiDepartmentDefaults(api)
			AssertV1ApiResources(api.NodePoolsResources, false, cpuFt, overrideDeserved)
		})
	})

	Context("test department mapper, cpu ft on", func() {
		cpuFt := true
		overrideDeserved := false

		It("map model to response", func() {
			department := createDepartment(false, cpuFt, true)
			api, err := v1.MapDepartmentModelToV1Response(context.Background(), *department, nil, false, map[string]clusteropenapi.NodepoolMinimal{"1": {Name: "default", Id: "1", SchedulingCapabilities: &clusteropenapi.SchedulingCapabilities{CpuQuota: true}}, nodePool2Id: {Name: nodePool2Name, Id: nodePool2Id, SchedulingCapabilities: &clusteropenapi.SchedulingCapabilities{CpuQuota: true}}})
			Expect(err).To(BeNil())
			assertV1ApiDepartmentDefaults(api)
			AssertV1ApiResources(api.NodePoolsResources, false, cpuFt, overrideDeserved)
		})
	})
})

func assertV1ApiDepartmentDefaults(requestToApi api.DepartmentForGetV1) {
	Expect(requestToApi.Name).To(Equal(departmentName))
	Expect(requestToApi.ProjectsDeservedGpus).To(Equal(""))
}
