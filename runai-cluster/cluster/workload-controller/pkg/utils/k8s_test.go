/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"testing"

	client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	accessRuleName1 = "accessrule-name1"
	accessRuleName2 = "accessrule-name2"
	accessRuleName3 = "accessrule-name3"
	accessRuleName4 = "accessrule-name4"
)

func init() {
	_ = runaiv2.AddToScheme(scheme.Scheme)
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

func TestJobSuspend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "k8s utils tests")
}

var _ = Describe("k8s utils tests", func() {
	var (
		k8sClient client.WithWatch
	)
	Context("k8s utils", func() {
		Describe("ListAccessRules", func() {
			BeforeEach(func() {
				accessRule1Obj := &v1alpha1.AccessRule{
					ObjectMeta: metav1.ObjectMeta{
						Name:   accessRuleName1,
						Labels: map[string]string{},
					},
					Spec: v1alpha1.AccessRuleSpec{
						ScopeType: v1alpha1.DepartmentScope,
						RoleName:  "department-administrator",
						SubjectId: "subject-id",
					},
				}
				accessRule2Obj := &v1alpha1.AccessRule{
					ObjectMeta: metav1.ObjectMeta{
						Name:   accessRuleName2,
						Labels: map[string]string{},
					},
					Spec: v1alpha1.AccessRuleSpec{
						ScopeType: v1alpha1.ProjectScope,
						RoleName:  "system-administrator",
						SubjectId: "subject-id",
					},
				}
				accessRule3Obj := &v1alpha1.AccessRule{
					ObjectMeta: metav1.ObjectMeta{
						Name:   accessRuleName3,
						Labels: map[string]string{},
					},
					Spec: v1alpha1.AccessRuleSpec{
						ScopeType: v1alpha1.ClusterScope,
						RoleName:  "l1-researcher",
						SubjectId: "subject-id2",
					},
				}
				accessRule4Obj := &v1alpha1.AccessRule{
					ObjectMeta: metav1.ObjectMeta{
						Name:   accessRuleName4,
						Labels: map[string]string{},
					},
					Spec: v1alpha1.AccessRuleSpec{
						ScopeType: v1alpha1.ClusterScope,
						RoleName:  "editor",
						SubjectId: "subject-id2",
					},
				}

				objects := []runtime.Object{accessRule1Obj, accessRule2Obj, accessRule3Obj, accessRule4Obj}
				k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
				k8sClient = &client2.FakeClientWithFieldSelector{WithWatch: k8sClient}
			})
			It("ListAccessRules - validate response with field selector", func() {
				selector := fields.OneTermEqualSelector(
					common.PrivilegedWorkloadCrudPermissionsField, "subject-id")
				accessRules, err := ListAccessRules(context.TODO(), k8sClient, &selector)
				Expect(err).To(BeNil())
				Expect(accessRules).To(Not(BeNil()))
				Expect(len(accessRules)).To(Equal(1))
				Expect(accessRules[0].Name).To(Equal(accessRuleName1))
			})
			It("ListAccessRules - validate response with field selector 2", func() {
				selector := fields.OneTermEqualSelector(
					common.PrivilegedWorkloadCrudPermissionsField, "subject-id2")
				accessRules, err := ListAccessRules(context.TODO(), k8sClient, &selector)
				Expect(err).To(BeNil())
				Expect(accessRules).To(Not(BeNil()))
				Expect(len(accessRules)).To(Equal(1))
				Expect([]string{accessRules[0].Name}).To(ConsistOf([]string{accessRuleName4}))
			})
			It("ListAccessRules3 - without field selector", func() {
				accessRules, err := ListAccessRules(context.TODO(), k8sClient, nil)
				Expect(err).To(BeNil())
				Expect(accessRules).To(Not(BeNil()))
				Expect(len(accessRules)).To(Equal(4))
			})
		})
	})
})
