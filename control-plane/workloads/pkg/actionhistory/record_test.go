/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actionhistory_test

import (
	"context"
	"io"
	"net/http"
	"sync"
	"testing/iotest"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/actionhistory"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func authenticatedContext(tenant, subject, subjectType string) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenant)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subject)
	return context.WithValue(ctx, runaictx.CfSubjectType, subjectType)
}

func acceptedWorkload() actionhistory.WorkloadToLog {
	return actionhistory.WorkloadToLog{ID: uuid.New(), ClusterID: uuid.New(), Name: "workload"}
}

type authenticatedContextCase struct {
	ctx context.Context
}

var _ = Describe("NewWorkloadActionRecord", func() {
	It("returns 503 when action ID generation fails", func() {
		ctx := authenticatedContext("42", "user@example.test", "user")
		workload := acceptedWorkload()
		uuid.SetRand(iotest.ErrReader(io.ErrUnexpectedEOF))
		DeferCleanup(func() { uuid.SetRand(nil) })

		record, apiErr := actionhistory.NewWorkloadActionRecord(ctx, uuid.Nil, api.WORKLOADACTIONTYPE_CREATE, workload, time.Now())

		Expect(record).To(BeNil())
		Expect(apiErr).To(HaveOccurred())
		Expect(apiErr.Code).To(Equal(int32(http.StatusServiceUnavailable)))
	})

	DescribeTable("rejects invalid authenticated context", func(tc authenticatedContextCase) {
		record, apiErr := actionhistory.NewWorkloadActionRecord(tc.ctx, uuid.Nil, api.WORKLOADACTIONTYPE_CREATE, acceptedWorkload(), time.Now())
		Expect(record).To(BeNil())
		Expect(apiErr).To(HaveOccurred())
	},
		Entry("nil context", authenticatedContextCase{}),
		Entry("missing identity", authenticatedContextCase{context.Background()}),
		Entry("negative tenant", authenticatedContextCase{authenticatedContext("-1", "user@example.test", "user")}),
		Entry("non-numeric tenant", authenticatedContextCase{authenticatedContext("not-a-tenant", "user@example.test", "user")}),
		Entry("bigint overflow", authenticatedContextCase{authenticatedContext("9223372036854775808", "user@example.test", "user")}),
		Entry("empty subject type", authenticatedContextCase{authenticatedContext("42", "user@example.test", "")}),
		Entry("blank subject type", authenticatedContextCase{authenticatedContext("42", "user@example.test", " \t")}),
		Entry("wrong subject type value", authenticatedContextCase{context.WithValue(authenticatedContext("42", "user@example.test", "user"), runaictx.CfSubjectType, 1)}),
	)

	DescribeTable("snapshots authenticated identity", func(subject, subjectType string) {
		ctx := authenticatedContext("42", subject, subjectType)
		workload := acceptedWorkload()
		acceptedAt := time.Date(2026, 10, 11, 14, 15, 16, 123456789, time.FixedZone("offset", 3*60*60))
		first, apiErr := actionhistory.NewWorkloadActionRecord(ctx, uuid.Nil, api.WORKLOADACTIONTYPE_CREATE, workload, acceptedAt)
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(first).NotTo(BeNil())
		Expect(first.ID).NotTo(Equal(uuid.Nil))
		Expect(first.ID.Version()).To(Equal(uuid.Version(4)))
		Expect(first.TenantID).To(Equal(uint(42)))
		Expect(first.Subject).To(Equal(subject))
		Expect(first.SubjectType).To(Equal(subjectType))
		Expect(first.WorkloadID).To(Equal(workload.ID))
		Expect(first.ClusterID).To(Equal(workload.ClusterID))
		Expect(first.WorkloadName).To(Equal(workload.Name))
		Expect(first.Action).To(Equal(string(api.WORKLOADACTIONTYPE_CREATE)))
		Expect(first.AcceptedAt).To(Equal(acceptedAt.UTC().Truncate(time.Microsecond)))

		workload.Name = "changed"
		ctx = context.WithValue(ctx, runaictx.CfSubjectID, "another-user@example.test")
		second, apiErr := actionhistory.NewWorkloadActionRecord(ctx, uuid.Nil, api.WORKLOADACTIONTYPE_CREATE, workload, acceptedAt)
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(second.ID).NotTo(Equal(first.ID))
		Expect(first.Subject).To(Equal(subject))
		Expect(first.WorkloadName).To(Equal("workload"))
		Expect(second.Subject).To(Equal("another-user@example.test"))
	},
		Entry("gateway user", "user@example.test", "User"),
		Entry("gateway service account", "service-account-id", "ServiceAccount"),
		Entry("user", "user@example.test", "user"),
		Entry("service account", "service-account-id", "service-account"),
		Entry("legacy app", "app-id", "App"),
		Entry("legacy lowercase app", "app-id", "app"),
		Entry("future identity type", "identity-id", "FutureIdentity"),
		Entry("value preserved without trimming", "identity-id", " User "),
	)

	DescribeTable("accepts tenant boundaries", func(tenant string, expected uint64) {
		ctx := authenticatedContext(tenant, "user@example.test", "user")
		record, apiErr := actionhistory.NewWorkloadActionRecord(ctx, uuid.Nil, api.WORKLOADACTIONTYPE_CREATE, acceptedWorkload(), time.Now())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(uint64(record.TenantID)).To(Equal(expected))
	},
		Entry("zero", "0", uint64(0)),
		Entry("maximum PostgreSQL bigint", "9223372036854775807", uint64(9223372036854775807)),
	)

	It("assigns unique IDs to concurrent batch items", func() {
		ctx := authenticatedContext("42", "user@example.test", "user")
		const count = 32
		records := make(chan *models.WorkloadActionHistory, count)
		errors := make(chan error, count)
		var wg sync.WaitGroup
		for range count {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				record, apiErr := actionhistory.NewWorkloadActionRecord(ctx, uuid.Nil, api.WORKLOADACTIONTYPE_DELETE, acceptedWorkload(), time.Now())
				if apiErr != nil {
					errors <- apiErr
					return
				}
				records <- record
			}()
		}
		wg.Wait()
		close(records)
		close(errors)
		for err := range errors {
			Expect(err).NotTo(HaveOccurred())
		}
		ids := make(map[uuid.UUID]struct{}, count)
		for record := range records {
			Expect(record).NotTo(BeNil())
			Expect(record.Subject).To(Equal("user@example.test"))
			Expect(record.TenantID).To(Equal(uint(42)))
			Expect(ids).NotTo(HaveKey(record.ID))
			ids[record.ID] = struct{}{}
		}
		Expect(ids).To(HaveLen(count))
	})
})
