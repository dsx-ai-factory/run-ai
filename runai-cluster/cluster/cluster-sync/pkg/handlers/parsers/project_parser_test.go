/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	org_unit_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	runaikaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ParseKaiProject", func() {
	const projectId = "project-id-123"

	buildKaiProject := func() *runaikaiv1alpha1.Project {
		return &runaikaiv1alpha1.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "my-project",
				Labels: map[string]string{parsers.ProjectIdLabel: projectId},
			},
			Status: runaikaiv1alpha1.ProjectStatus{
				Namespace: "runai-my-project",
				Phase:     runaikaiv1alpha1.Ready,
				Message:   "all good",
			},
		}
	}

	Describe("ParseKaiProjectToUpdateStatusRequest", func() {
		It("maps the id from the project-id label", func() {
			req := parsers.ParseKaiProjectToUpdateStatusRequest(buildKaiProject())
			Expect(req.Id).To(Equal(projectId))
		})

		It("maps namespace, message and phase from status", func() {
			req := parsers.ParseKaiProjectToUpdateStatusRequest(buildKaiProject())
			Expect(req.Status.Namespace.Get()).To(HaveValue(Equal("runai-my-project")))
			Expect(req.Status.PhaseMessage.Get()).To(HaveValue(Equal("all good")))
			Expect(req.Status.Phase).To(HaveValue(Equal(org_unit_openapi.PROJECTPHASE_READY)))
		})

		DescribeTable("phase conversion",
			func(phase runaikaiv1alpha1.ProjectPhase, expected org_unit_openapi.ProjectPhase) {
				project := buildKaiProject()
				project.Status.Phase = phase
				req := parsers.ParseKaiProjectToUpdateStatusRequest(project)
				Expect(req.Status.Phase).To(HaveValue(Equal(expected)))
			},
			Entry("ready", runaikaiv1alpha1.Ready, org_unit_openapi.PROJECTPHASE_READY),
			Entry("not ready", runaikaiv1alpha1.NotReady, org_unit_openapi.PROJECTPHASE_NOT_READY),
			Entry("empty -> initializing", runaikaiv1alpha1.ProjectPhase(""), org_unit_openapi.PROJECTPHASE_INITIALIZING),
			Entry("unknown -> unknown", runaikaiv1alpha1.ProjectPhase("Bogus"), org_unit_openapi.PROJECTPHASE_UNKNOWN),
		)

		It("reads the last sync time from the annotation", func() {
			project := buildKaiProject()
			syncTime := time.Date(2026, 7, 12, 10, 30, 0, 0, time.UTC)
			project.Annotations = map[string]string{
				parsers.LastSyncTimestampAnnotation: syncTime.Format(time.RFC3339),
			}
			req := parsers.ParseKaiProjectToUpdateStatusRequest(project)
			Expect(req.Status.LastUpdatedTime.Get()).To(HaveValue(BeTemporally("==", syncTime)))
		})

		It("falls back to the zero time when the annotation is missing", func() {
			req := parsers.ParseKaiProjectToUpdateStatusRequest(buildKaiProject())
			Expect(req.Status.LastUpdatedTime.Get()).To(HaveValue(Equal(time.Time{})))
		})
	})

	Describe("ParseKaiProjectToDeleteRequest", func() {
		It("sets the phase to deleted and keeps the id", func() {
			req := parsers.ParseKaiProjectToDeleteRequest(buildKaiProject())
			Expect(req.Id).To(Equal(projectId))
			Expect(req.Status.Phase).To(HaveValue(Equal(org_unit_openapi.PROJECTPHASE_DELETED)))
		})
	})
})
