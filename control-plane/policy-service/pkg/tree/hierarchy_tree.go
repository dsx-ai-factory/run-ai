package tree

import (
	"fmt"
)

type TreeId struct {
	TenantId     int32
	WorkloadType WorkloadType
}

func MakeTreeId(tenantId int32, workloadType WorkloadType) TreeId {
	return TreeId{
		TenantId:     tenantId,
		WorkloadType: workloadType,
	}
}

func (ti *TreeId) ToString() string {
	return fmt.Sprintf("%d:%s", ti.TenantId, ti.WorkloadType)
}

// HierarchyTree maps OrgUnit id (as a string, e.g. dep:1) to its node. The tree manager holds
// many trees like this, one for each [ tenant x workload type ]
type HierarchyTree struct {
	treeId TreeId
}

func NewHierarchyTree(tenantId int32, workloadType WorkloadType) *HierarchyTree {
	result := HierarchyTree{
		treeId: TreeId{TenantId: tenantId, WorkloadType: workloadType},
	}
	return &result
}
