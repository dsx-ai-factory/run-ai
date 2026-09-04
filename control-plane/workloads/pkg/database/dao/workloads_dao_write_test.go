/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
)

const (
	wlWriteDelete                  = "UPDATE workloads SET deleted_at=? WHERE tenant_id = ? AND workloads.id = ? AND workloads.deleted_at IS NULL"
	wlWriteSave                    = "UPDATE workloads SET updated_at=?,deleted_at=?,name=?,external_workload_name=?,source=?,priority_class_name=?,type=?,tenant_id=?,cluster_id=?,category_id=?,namespace=?,requested_gpu_limit=?,requested_gpu_request=?,requested_gpu_memory_limit=?,requested_gpu_memory_request=?,requested_cpu_limit=?,requested_cpu_request=?,requested_cpu_memory_limit=?,requested_cpu_memory_request=?,requested_mig_device=?,support_delete=?,phase=?,phase_message=?,phase_updated_at=?,k8s_phase=?,k8s_phase_updated_at=?,requested_pods_number=?,requested_pods_min=?,requested_pods_max=?,requested_pods_parallelism=?,requested_pods_completions=?,requested_nodepools=?,images=?,distributed_framework=?,urls=?,network_status=?,conditions=?,priority=?,additional_fields=?,fully_saas_workload=?,command=?,arguments=?,environment_variables=?,scheduling_conditions=?,phase_reason=?,guaranteed_runtime_ends_at=?,application_id=?,application_name=?,topology=?,idle_since=?,structure_elements_total_count=?,endpoints=?,node_affinities=?,k8s_object_group=?,k8s_object_version=?,resource_refs=? WHERE workloads.deleted_at IS NULL AND id = ?"
	wlWriteSavepointChildren       = "SAVEPOINT workload_children"
	wlWriteSavepointElements       = "SAVEPOINT workload_elements"
	wlWriteRollbackChildren        = "ROLLBACK TO SAVEPOINT workload_children"
	wlWriteRollbackElements        = "ROLLBACK TO SAVEPOINT workload_elements"
	wlWriteSavepointSp             = "SAVEPOINT sp"
	wlWriteRollbackSp              = "ROLLBACK TO SAVEPOINT sp"
	wlWriteChildInsert             = "INSERT INTO workloads_children (workload_id,child_id,child_type) VALUES (?,?,?) ON CONFLICT DO NOTHING"
	wlWriteElementsDelete          = "DELETE FROM workload_elements WHERE tenant_id = ? AND workload_id = ? AND cluster_id = ? AND id NOT IN (?)"
	wlWriteElementsDeleteNoCluster = "DELETE FROM workload_elements WHERE tenant_id = ? AND workload_id = ? AND id NOT IN (?)"
	wlWriteElementsDeleteAll       = "DELETE FROM workload_elements WHERE tenant_id = ? AND workload_id = ?"
	wlWriteElementsInsert          = "INSERT INTO workload_elements (id,workload_id,tenant_id,cluster_id,name,level,parent_id,instance_key,replica_key,k8s_ref_uid,k8s_ref_name,gvk_group,gvk_version,gvk_kind,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING"
	wlWritePhaseUpdate             = "UPDATE workloads SET deleted_at=?,phase=?,phase_message=? WHERE tenant_id = ? AND id = ? AND phase_updated_at < ?"
	wlWriteInternalInsert          = "INSERT INTO workloads (id,created_at,updated_at,deleted_at,submitted_by,name,external_workload_name,source,priority_class_name,preemptibility,type,tenant_id,cluster_id,project_name,project_id,department_name,department_id,category_id,namespace,requested_gpu_limit,requested_gpu_request,requested_gpu_memory_limit,requested_gpu_memory_request,requested_cpu_limit,requested_cpu_request,requested_cpu_memory_limit,requested_cpu_memory_request,requested_mig_device,support_delete,support_suspend,phase,phase_message,phase_updated_at,k8s_phase,k8s_phase_updated_at,requested_pods_number,requested_pods_min,requested_pods_max,requested_pods_parallelism,requested_pods_completions,requested_nodepools,images,distributed_framework,urls,network_status,conditions,priority,additional_fields,fully_saas_workload,command,arguments,environment_variables,scheduling_conditions,phase_reason,guaranteed_runtime_ends_at,application_id,application_name,topology,idle_since,external_connections,structure_elements_total_count,endpoints,node_affinities,k8s_object_group,k8s_object_version,resource_refs) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING"
	wlWriteInternalUpdate          = "UPDATE workloads SET datasources=?,environments=?,external_connections=?,preemptibility=?,source_api=?,support_delete=?,support_suspend=? WHERE id = ? AND tenant_id = ?"
	wlWriteInternalUpdateMin       = "UPDATE workloads SET preemptibility=?,source_api=? WHERE id = ? AND tenant_id = ?"
	wlWriteConnMetaUpdate          = "UPDATE workloads SET environments=?,external_connections=? WHERE id = ? AND tenant_id = ?"
	wlWriteSyncDelete              = "UPDATE workloads SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND workloads.deleted_at IS NULL RETURNING id"
	wlWriteSyncDeleteExcl          = "UPDATE workloads SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND NOT id IN (?) AND workloads.deleted_at IS NULL RETURNING id"
	wlWriteSyncDeleteFilter        = "UPDATE workloads SET deleted_at=? WHERE tenant_id = ? AND project_id = ? AND id >= ? AND id <= ? AND NOT id IN (?) AND phase = ? AND workloads.deleted_at IS NULL RETURNING id"
	wlWriteClusterDelete           = "DELETE FROM workloads WHERE tenant_id = ? AND cluster_id = ?"
	wlWriteTenantDelete            = "DELETE FROM workloads WHERE tenant_id = ?"
)

type wlWriteIdRow struct {
	ID string `gorm:"column:id"`
}

var wlWriteId = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var wlWriteChildId = uuid.MustParse("22222222-2222-2222-2222-222222222222")
var wlWriteClusterId = uuid.MustParse("33333333-3333-3333-3333-333333333333")

func wlWriteWorkload() models.Workload {
	return models.Workload{
		ID:       wlWriteId,
		Name:     "wl",
		TenantID: 1,
	}
}

func wlWriteFullWorkload() models.Workload {
	preemptibility := "Preemptible"
	return models.Workload{
		ID:                  wlWriteId,
		Name:                "wl",
		TenantID:            1,
		ClusterID:           wlWriteClusterId,
		ProjectID:           "p1",
		ProjectName:         "project",
		DepartmentID:        "d1",
		DepartmentName:      "dep",
		SubmittedBy:         "user",
		Preemptibility:      &preemptibility,
		ExternalConnections: models.ExternalConnectionArray{},
		SupportSuspend:      true,
	}
}

func wlWriteChildren() []models.WorkloadChild {
	return []models.WorkloadChild{{WorkloadID: wlWriteId, ChildID: wlWriteChildId, ChildType: "Pod"}}
}

func wlWriteElements() []models.WorkloadElement {
	return []models.WorkloadElement{{ID: "el-1", WorkloadID: wlWriteId, TenantID: 1, ClusterID: wlWriteClusterId, Name: "el"}}
}

func wlWriteInternalFields() models.InternalCreatedWorkload {
	sourceApi := api.SOURCEAPI_WORKLOADS_V2
	preemptibility := "Preemptible"
	return models.InternalCreatedWorkload{
		Datasources:         models.DatasourceArray{},
		Environments:        models.EnvironmentArray{},
		ExternalConnections: models.ExternalConnectionArray{},
		SourceApi:           &sourceApi,
		Preemptibility:      &preemptibility,
	}
}

func wlWriteConnMetaArgs() (models.EnvironmentArray, models.ExternalConnectionArray) {
	return models.EnvironmentArray{{Name: "e"}}, models.ExternalConnectionArray{{Name: "c"}}
}

var _ = Describe("DbDeleteWorkload", func() {
	It("soft deletes the workload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteDelete).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbDeleteWorkload(testCtx(1), 1, wlWriteId)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := (&WorkloadsDao{}).DbDeleteWorkload(testCtx(1), 1, wlWriteId)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertWorkload", func() {
	It("saves the workload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkload(testCtx(1), wlWriteWorkload())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the save fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkload(testCtx(1), wlWriteWorkload())

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertWorkloadFull", func() {
	It("saves the workload together with its children and elements", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).Returns(1)
		m.ExpectExec(wlWriteSavepointChildren).Returns(0)
		m.ExpectExec(wlWriteChildInsert).Returns(1)
		m.ExpectExec(wlWriteSavepointElements).Returns(0)
		m.ExpectExec(wlWriteElementsDelete).Returns(0)
		m.ExpectExec(wlWriteElementsInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadFull(testCtx(1), 1, wlWriteClusterId.String(),
			wlWriteWorkload(), wlWriteChildren(), wlWriteElements())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadFull(testCtx(1), 1, "", wlWriteWorkload(), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the workload save fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadFull(testCtx(1), 1, "", wlWriteWorkload(), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the children insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).Returns(1)
		m.ExpectExec(wlWriteSavepointChildren).Returns(0)
		m.ExpectExec(wlWriteChildInsert).ReturnsError(errors.New("boom"))
		m.ExpectExec(wlWriteRollbackChildren).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadFull(testCtx(1), 1, "", wlWriteWorkload(), wlWriteChildren(), wlWriteElements())

		Expect(apiErr).ToNot(BeNil())
	})

	It("succeeds when only the elements update fails and is rolled back to its savepoint", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).Returns(1)
		m.ExpectExec(wlWriteSavepointChildren).Returns(0)
		m.ExpectExec(wlWriteSavepointElements).Returns(0)
		m.ExpectExec(wlWriteElementsDeleteNoCluster).ReturnsError(errors.New("boom"))
		m.ExpectExec(wlWriteRollbackElements).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadFull(testCtx(1), 1, "", wlWriteWorkload(), nil, wlWriteElements())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSave).Returns(1)
		m.ExpectExec(wlWriteSavepointChildren).Returns(0)
		m.ExpectExec(wlWriteSavepointElements).Returns(0)
		m.ExpectExec(wlWriteElementsDeleteAll).Returns(0)

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadFull(testCtx(1), 1, "", wlWriteWorkload(), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpdateWorkloadPhase", func() {
	It("updates the phase", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWritePhaseUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpdateWorkloadPhase(testCtx(1), 1, wlWriteId.String(),
			api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, time.Now())

		Expect(apiErr).To(BeNil())
	})
})

var _ = Describe("DbUpdateWorkloadPhaseWithResult", func() {
	It("reports the workload as updated", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWritePhaseUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		updated, apiErr := DbUpdateWorkloadPhaseWithResult(testCtx(1), 1, wlWriteId.String(),
			api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, time.Now())

		Expect(apiErr).To(BeNil())
		Expect(updated).To(BeTrue())
	})

	It("reports the workload as updated when the patch also soft deletes it", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWritePhaseUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		updated, apiErr := DbUpdateWorkloadPhaseWithResult(testCtx(1), 1, wlWriteId.String(),
			api.WorkloadPatchFields{Phase: api.PHASE_DELETING, SoftDelete: true}, time.Now())

		Expect(apiErr).To(BeNil())
		Expect(updated).To(BeTrue())
	})

	It("reports the workload as not updated when no rows are affected", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWritePhaseUpdate).Returns(0)
		m.ExpectTransaction().Commit()

		updated, apiErr := DbUpdateWorkloadPhaseWithResult(testCtx(1), 1, wlWriteId.String(),
			api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, time.Now())

		Expect(apiErr).To(BeNil())
		Expect(updated).To(BeFalse())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWritePhaseUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		updated, apiErr := DbUpdateWorkloadPhaseWithResult(testCtx(1), 1, wlWriteId.String(),
			api.WorkloadPatchFields{Phase: api.PHASE_RUNNING}, time.Now())

		Expect(apiErr).ToNot(BeNil())
		Expect(updated).To(BeFalse())
	})
})

var _ = Describe("DbCreateInternalWorkload", func() {
	It("creates the workload and updates all internal fields", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteInternalInsert).Returns(1)
		m.ExpectExec(wlWriteInternalUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbCreateInternalWorkload(testCtx(1), wlWriteWorkload(), wlWriteInternalFields())

		Expect(apiErr).To(BeNil())
	})

	It("updates only the minimal set of fields when the internal fields are empty", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteInternalInsert).Returns(1)
		m.ExpectExec(wlWriteInternalUpdateMin).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbCreateInternalWorkload(testCtx(1), wlWriteWorkload(), models.InternalCreatedWorkload{})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := (&WorkloadsDao{}).DbCreateInternalWorkload(testCtx(1), wlWriteWorkload(), models.InternalCreatedWorkload{})

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the create fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteInternalInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := (&WorkloadsDao{}).DbCreateInternalWorkload(testCtx(1), wlWriteWorkload(), models.InternalCreatedWorkload{})

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteInternalInsert).Returns(1)
		m.ExpectExec(wlWriteInternalUpdateMin).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := (&WorkloadsDao{}).DbCreateInternalWorkload(testCtx(1), wlWriteWorkload(), models.InternalCreatedWorkload{})

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteInternalInsert).Returns(1)
		m.ExpectExec(wlWriteInternalUpdateMin).Returns(1)

		apiErr := (&WorkloadsDao{}).DbCreateInternalWorkload(testCtx(1), wlWriteWorkload(), models.InternalCreatedWorkload{})

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpdateWorkloadConnectionMetadata", func() {
	It("does nothing when there is nothing to update", func() {
		setupDaoMock(GinkgoT())

		apiErr := (&WorkloadsDao{}).DbUpdateWorkloadConnectionMetadata(testCtx(1), 1, wlWriteId, nil, nil)

		Expect(apiErr).To(BeNil())
	})

	It("updates the connection metadata", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteConnMetaUpdate).Returns(1)
		m.ExpectTransaction().Commit()
		environments, externalConnections := wlWriteConnMetaArgs()

		apiErr := (&WorkloadsDao{}).DbUpdateWorkloadConnectionMetadata(testCtx(1), 1, wlWriteId, environments, externalConnections)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the workload is not found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteConnMetaUpdate).Returns(0)
		m.ExpectTransaction().Commit()
		environments, externalConnections := wlWriteConnMetaArgs()

		apiErr := (&WorkloadsDao{}).DbUpdateWorkloadConnectionMetadata(testCtx(1), 1, wlWriteId, environments, externalConnections)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteConnMetaUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()
		environments, externalConnections := wlWriteConnMetaArgs()

		apiErr := (&WorkloadsDao{}).DbUpdateWorkloadConnectionMetadata(testCtx(1), 1, wlWriteId, environments, externalConnections)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbSyncWorkloads", func() {
	It("syncs the given workloads and reports the deleted ones", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectExec(wlWriteSave).Returns(1)
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectQuery(wlWriteSyncDeleteExcl).Returns([]wlWriteIdRow{{ID: wlWriteChildId.String()}})
		m.ExpectTransaction().Commit()

		synced, failed, deleted, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "",
			wlWriteClusterId.String(), []models.Workload{wlWriteWorkload()}, nil, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(synced).To(Equal([]string{wlWriteId.String()}))
		Expect(failed).To(BeEmpty())
		Expect(deleted).To(Equal([]string{wlWriteChildId.String()}))
	})

	It("reports a workload as failed when its save fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectExec(wlWriteSave).ReturnsError(errors.New("boom"))
		m.ExpectExec(wlWriteRollbackSp).Returns(0)
		m.ExpectQuery(wlWriteSyncDeleteExcl).Returns([]wlWriteIdRow{})
		m.ExpectTransaction().Commit()

		synced, failed, deleted, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "",
			wlWriteClusterId.String(), []models.Workload{wlWriteWorkload()}, nil, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(Equal([]string{wlWriteId.String()}))
		Expect(deleted).To(BeEmpty())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		_, _, _, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "", "", nil, nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns a bad request error when neither a project nor a cluster scope is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectTransaction().Rollback()

		_, _, _, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "", "", nil, nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(400)))
	})

	It("applies the id range and the extra filters", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectQuery(wlWriteSyncDeleteFilter).Returns([]wlWriteIdRow{})
		m.ExpectTransaction().Commit()

		from := wlWriteId
		to := wlWriteChildId
		filters := []queryapi.Filter{{Field: "phase", Operator: "=", Values: []string{"Running"}}}

		_, _, _, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "p1", "", nil,
			&to, &from, []string{wlWriteId.String()}, filters)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the stale delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectQuery(wlWriteSyncDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		_, _, _, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "",
			wlWriteClusterId.String(), nil, nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteSavepointSp).Returns(0)
		m.ExpectQuery(wlWriteSyncDelete).Returns([]wlWriteIdRow{})

		_, _, _, apiErr := (&WorkloadsDao{}).DbSyncWorkloads(testCtx(1), 1, "",
			wlWriteClusterId.String(), nil, nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertWorkloadsChildren", func() {
	It("inserts the children", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteChildInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsChildren(testCtx(1), wlWriteChildren(), nil)

		Expect(apiErr).To(BeNil())
	})

	It("skips children of failed workloads", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsChildren(testCtx(1), wlWriteChildren(), []string{wlWriteId.String()})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsChildren(testCtx(1), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteChildInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsChildren(testCtx(1), wlWriteChildren(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsChildren(testCtx(1), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertWorkloadsPodsChildren", func() {
	It("inserts the pod children", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteChildInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsPodsChildren(testCtx(1), wlWriteChildren(), nil)

		Expect(apiErr).To(BeNil())
	})

	It("filters out the failed pods", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectTransaction().Commit()

		apiErr := (&WorkloadsDao{}).DbUpsertWorkloadsPodsChildren(testCtx(1), wlWriteChildren(), []string{wlWriteChildId.String()})

		Expect(apiErr).To(BeNil())
	})
})

var _ = Describe("DeleteWorkloadsOfClusterId", func() {
	It("deletes the workloads of the cluster", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteClusterDelete).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadsOfClusterId(testCtx(1), "1", wlWriteClusterId.String())

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteClusterDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadsOfClusterId(testCtx(1), "1", wlWriteClusterId.String())

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteWorkloadsOfTenantId", func() {
	It("deletes the workloads of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteTenantDelete).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadsOfTenantId(testCtx(1), "1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wlWriteTenantDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadsOfTenantId(testCtx(1), "1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("omitWorkloadsEmptyStaticFields", func() {
	It("omits every empty static column", func() {
		setupDaoMock(GinkgoT())

		query := omitWorkloadsEmptyStaticFields(runaidb.DB().Table(WorkloadsTableName), models.Workload{})

		Expect(query.Actual().Statement.Omits).To(Equal([]string{
			"project_id", "project_name", "department_id", "department_name",
			"submitted_by", WorkloadsColumnPreemptibility, WorkloadsColumnExternalConnections,
			WorkloadsColumnSupportSuspend,
		}))
	})

	It("omits nothing when every static column is populated", func() {
		setupDaoMock(GinkgoT())

		query := omitWorkloadsEmptyStaticFields(runaidb.DB().Table(WorkloadsTableName), wlWriteFullWorkload())

		Expect(query.Actual().Statement.Omits).To(BeEmpty())
	})
})
