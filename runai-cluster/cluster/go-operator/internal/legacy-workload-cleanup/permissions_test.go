/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permission preflight", func() {
	It("checks every required permission before mutation", func() {
		reviewer := &fakePermissionReviewer{}
		permissions := []requiredPermission{
			{Verb: "list", Group: "run.ai", Resource: "trainingworkloads"},
			{Verb: "patch", Group: "run.ai", Resource: "trainingworkloads"},
		}
		Expect(checkPermissions(context.Background(), reviewer, permissions)).To(Succeed())
		Expect(reviewer.reviewed).To(Equal(permissions))
	})

	It("reports the exact denied verb, resource, and scope", func() {
		reviewer := &fakePermissionReviewer{deny: requiredPermission{
			Verb: "patch", Group: "run.ai", Resource: "trainingworkloads", Namespace: "project-a",
		}}
		err := checkPermissions(context.Background(), reviewer, []requiredPermission{reviewer.deny})
		Expect(err).To(MatchError(And(
			ContainSubstring("patch"), ContainSubstring("run.ai/trainingworkloads"), ContainSubstring("project-a"),
		)))
	})

	It("fails on a SelfSubjectAccessReview error", func() {
		reviewer := &fakePermissionReviewer{reviewError: fmt.Errorf("review unavailable")}
		err := checkPermissions(context.Background(), reviewer, []requiredPermission{{Verb: "get", Resource: "secrets"}})
		Expect(err).To(MatchError(ContainSubstring("review unavailable")))
	})

	It("does not require the pre-upgrade backup ConfigMap", func() {
		management := &fakePermissionReviewer{}
		workload := &fakePermissionReviewer{}
		runner := &Runner{ManagementReviewer: management, WorkloadReviewer: workload}

		Expect(runner.checkAllPermissions(context.Background(), nil)).To(Succeed())
		Expect(management.reviewed).NotTo(ContainElement(requiredPermission{
			Verb: "get", Resource: "configmaps", Namespace: "runai", Name: "runai",
		}))
	})

	It("requests only permissions used by the cleanup transaction", func() {
		management := &fakePermissionReviewer{}
		workload := &fakePermissionReviewer{}
		runner := &Runner{ManagementReviewer: management, WorkloadReviewer: workload}

		Expect(runner.checkAllPermissions(context.Background(), nil)).To(Succeed())
		Expect(management.reviewed).To(ConsistOf(
			requiredPermission{Verb: "create", Group: "authorization.k8s.io", Resource: "selfsubjectaccessreviews"},
			requiredPermission{Verb: "get", Group: "run.ai", Resource: "runaiconfigs", Namespace: "runai", Name: "runai"},
		))
		Expect(workload.reviewed).To(ContainElements(
			requiredPermission{Verb: "create", Group: "authorization.k8s.io", Resource: "selfsubjectaccessreviews"},
			requiredPermission{Verb: "list", Group: "admissionregistration.k8s.io", Resource: "validatingwebhookconfigurations"},
			requiredPermission{Verb: "list", Group: "admissionregistration.k8s.io", Resource: "mutatingwebhookconfigurations"},
			requiredPermission{Verb: "delete", Group: "apps", Resource: "deployments", Namespace: "runai", Name: "workload-controller"},
			requiredPermission{Verb: "list", Resource: "pods", Namespace: "runai"},
			requiredPermission{Verb: "delete", Resource: "services", Namespace: "runai", Name: "workload-controller"},
			requiredPermission{Verb: "list", Group: "rbac.authorization.k8s.io", Resource: "rolebindings"},
			requiredPermission{Verb: "delete", Group: "rbac.authorization.k8s.io", Resource: "rolebindings"},
		))
		for _, descriptor := range legacyWorkloadResources {
			Expect(workload.reviewed).To(ContainElement(requiredPermission{
				Verb: "get", Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions", Name: descriptor.CRDName,
			}))
		}
	})
})

type fakePermissionReviewer struct {
	reviewed    []requiredPermission
	deny        requiredPermission
	reviewError error
}

func (f *fakePermissionReviewer) Review(_ context.Context, permission requiredPermission) (bool, string, error) {
	f.reviewed = append(f.reviewed, permission)
	if f.reviewError != nil {
		return false, "", f.reviewError
	}
	if permission == f.deny && f.deny.Verb != "" {
		return false, "denied by policy", nil
	}
	return true, "", nil
}
