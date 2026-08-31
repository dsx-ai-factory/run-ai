/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	TrainingName  = "wl-training01"
	TrainingName2 = "wl-training02"
	TrainingName3 = "wl-training03"
	WorkloadKind  = "just-kind-value"
)

var (
	training = &v2alpha1.TrainingWorkload{ObjectMeta: metav1.ObjectMeta{
		Name: TrainingName,
	},
		TypeMeta: metav1.TypeMeta{
			Kind: WorkloadKind,
		},
	}
	training2 = &v2alpha1.TrainingWorkload{ObjectMeta: metav1.ObjectMeta{
		Name: TrainingName2,
	},
		TypeMeta: metav1.TypeMeta{
			Kind: WorkloadKind,
		},
	}
	training3 = &v2alpha1.TrainingWorkload{ObjectMeta: metav1.ObjectMeta{
		Name: TrainingName3,
	},
		TypeMeta: metav1.TypeMeta{
			Kind: WorkloadKind,
		},
	}
)

func TestIssue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aggregator Suite")
}

var _ = Describe("Aggregator Test Suite", func() {
	var (
		aggr *resource.Aggregator
	)

	BeforeEach(func() {
		aggr = resource.NewAggregator(nil)
	})

	Context("Aggregator", func() {
		It("New Aggregator", func() {
			Expect(aggr.HasIssues()).To(Equal(false))
			Expect(aggr.HasCritical()).To(Equal(false))
			Expect(len(aggr.GetIssuesList())).To(Equal(0))
		})

		It("None fatal core issue", func() {
			aggr.AddIssues(issue.New(training, fmt.Errorf("none fatal"), false, "", nil))
			Expect(aggr.HasIssues()).To(Equal(true))
			Expect(aggr.HasCritical()).To(Equal(false))
			Expect(len(aggr.GetIssuesList())).To(Equal(1))
			Expect(aggr.GetIssue(training.Kind, training.Name).GetGroup()).To(Equal(issue.Internal))
			Expect(aggr.GetIssue(training.Kind, training.Name).Message()).To(Equal(getMessage(training.Kind, training.Name, "none fatal")))
		})

		It("Fatal core issue", func() {
			aggr.AddIssues(issue.New(training, fmt.Errorf("is fatal"), true, "", nil))
			Expect(aggr.HasIssues()).To(Equal(true))
			Expect(aggr.HasCritical()).To(Equal(true))
			Expect(len(aggr.GetIssuesList())).To(Equal(1))
			Expect(aggr.GetIssue(training.Kind, training.Name).GetGroup()).To(Equal(issue.Internal))
			Expect(aggr.GetIssue(training.Kind, training.Name).Message()).To(Equal(getMessage(training.Kind, training.Name, "is fatal")))
		})

		It("Fatal resource issue", func() {
			aggr.AddIssues(issue.New(&v1.Ingress{}, fmt.Errorf("ingress error"), true, "", nil))
			Expect(aggr.HasIssues()).To(Equal(true))
			Expect(aggr.HasCritical()).To(Equal(true))
			Expect(len(aggr.GetIssuesList())).To(Equal(1))
			Expect(aggr.GetIssue(constants.KindIngress, "").GetGroup()).To(Equal(issue.Network))
			Expect(aggr.GetIssue(constants.KindIngress, "").Message()).To(Equal("[Ingress:] ingress error"))
		})

		It("Multi issue with none fatal issue", func() {
			aggr.AddIssues(issue.New(training, fmt.Errorf("multi core issue"), false, "", nil))
			aggr.AddIssues(issue.New(&corev1.PersistentVolumeClaim{}, fmt.Errorf("multi pvc issue"), false, "", nil))
			Expect(aggr.HasIssues()).To(Equal(true))
			Expect(aggr.HasCritical()).To(Equal(false))
			Expect(len(aggr.GetIssuesList())).To(Equal(2))
			Expect(aggr.GetIssue(WorkloadKind, training.Name).Message()).To(Equal(getMessage(training.Kind, training.Name, "multi core issue")))
			Expect(aggr.GetIssue(WorkloadKind, training.Name).GetGroup()).To(Equal(issue.Internal))
			Expect(aggr.GetIssue(constants.KindPvc, "").Message()).To(Equal("[PersistentVolumeClaim:] multi pvc issue"))
			Expect(aggr.GetIssue(constants.KindPvc, "").GetGroup()).To(Equal(issue.Storage))
		})

		It("Multi issue with fatal issue", func() {
			aggr.AddIssues(issue.New(training, fmt.Errorf("core issue"), false, "", nil))
			aggr.AddIssues(issue.New(&corev1.PersistentVolumeClaim{}, fmt.Errorf("pvc issue"), true, "", nil))
			Expect(aggr.HasIssues()).To(Equal(true))
			Expect(aggr.HasCritical()).To(Equal(true))
			Expect(len(aggr.GetIssuesList())).To(Equal(2))
			Expect(aggr.GetIssue(WorkloadKind, training.Name).Message()).To(Equal(getMessage(training.Kind, training.Name, "core issue")))
			Expect(aggr.GetIssue(WorkloadKind, training.Name).GetGroup()).To(Equal(issue.Internal))
			Expect(aggr.GetIssue(constants.KindPvc, "").Message()).To(Equal(getMessage(constants.KindPvc, "", "pvc issue")))
			Expect(aggr.GetIssue(constants.KindPvc, "").GetGroup()).To(Equal(issue.Storage))
		})

		It("GetIssueGroups", func() {
			aggr.AddIssues(issue.New(training, fmt.Errorf("core issue"), false, "", nil))
			aggr.AddIssues(issue.New(&v1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ingress1"}}, fmt.Errorf("ingress error"), true, "", nil))
			aggr.AddIssues(issue.New(&corev1.PersistentVolumeClaim{}, fmt.Errorf("pvc issue"), true, "", nil))
			aggr.AddIssues(issue.New(&v1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ingress2"}}, fmt.Errorf("ingress error 2"), false, "", nil))

			Expect(aggr.GetIssuesList()).To(HaveLen(4))
			groupsWithIssues := aggr.GetIssueGroups()
			Expect(groupsWithIssues).To(HaveLen(3))
			Expect(groupsWithIssues).To(ContainElements(issue.Internal, issue.Network, issue.Storage))
		})
	})

	Context("ExternalNetwork and HTTPRoute", func() {
		It("GetExternalNetwork returns empty when no network resources", func() {
			Expect(aggr.GetExternalNetwork()).To(BeEmpty())
		})

		It("GetExternalNetwork returns only Ingress when only Ingress added", func() {
			ing := &v1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ingress-only"}}
			aggr.AddIngresses(ing)

			result := aggr.GetExternalNetwork()
			Expect(result).To(HaveLen(1))
			Expect(result[0].GetName()).To(Equal("ingress-only"))
		})

		It("GetExternalNetwork returns HTTPRoute alongside Ingress and Route", func() {
			ing := &v1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "net-ingress"}}
			route := &openshiftV1.Route{ObjectMeta: metav1.ObjectMeta{Name: "net-route"}}
			hr := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "net-httproute"}}

			aggr.AddIngresses(ing)
			aggr.AddRoutes(route)
			aggr.AddHTTPRoutes(hr)

			result := aggr.GetExternalNetwork()
			Expect(result).To(HaveLen(3))

			names := make([]string, len(result))
			for i, obj := range result {
				names[i] = obj.GetName()
			}
			Expect(names).To(ContainElements("net-ingress", "net-route", "net-httproute"))
		})

		It("AddHTTPRoutes and GetHTTPRoutes work correctly", func() {
			hr1 := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "hr-1"}}
			hr2 := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "hr-2"}}

			aggr.AddHTTPRoutes(hr1, hr2)

			routes := aggr.GetHTTPRoutes()
			Expect(routes).To(HaveLen(2))
			Expect(routes[0].Name).To(Equal("hr-1"))
			Expect(routes[1].Name).To(Equal("hr-2"))
		})

		It("GetExternalNetwork returns only HTTPRoute when no Ingress or Route", func() {
			hr := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "gw-only"}}
			aggr.AddHTTPRoutes(hr)

			result := aggr.GetExternalNetwork()
			Expect(result).To(HaveLen(1))
			Expect(result[0].GetName()).To(Equal("gw-only"))
		})
	})

	Context("Result Calculation", func() {
		Context("Single Issue", func() {
			It("Aggregator No Issues - no requeue no fatal", func() {
				Expect(aggr.HasIssues()).To(BeFalse())
				Expect(aggr.HasCritical()).To(BeFalse())
			})

			It("Aggregator with fatal issue - fatal true", func() {
				aggr.AddIssues(issue.New(training, fmt.Errorf(""), true, "", nil))

				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeTrue())
			})
		})

		Context("Multi Issues Result with single resource", func() {
			It("Aggregator with recoverable --> fatal, result: no requeue, fatal", func() {
				// add issues
				aggr.AddIssues(issue.New(training, fmt.Errorf(""), false, "recoverable", nil))
				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeFalse())
				Expect(len(aggr.GetIssuesList())).To(Equal(1))
				Expect(aggr.GetIssuesList()[0].Message()).To(ContainSubstring("recoverable"))

				// add critical issue
				aggr.AddIssues(issue.New(training, fmt.Errorf(""), true, "fatal", nil))
				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeTrue())
				Expect(len(aggr.GetIssuesList())).To(Equal(1))
				Expect(aggr.GetIssuesList()[0].Message()).To(ContainSubstring("fatal"))

				// add issues - keep requeue and fail
				aggr.AddIssues(issue.New(training, fmt.Errorf(""), false, "recoverable", nil))
				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeTrue())
				Expect(len(aggr.GetIssuesList())).To(Equal(1))
				Expect(aggr.GetIssuesList()[0].Message()).To(ContainSubstring("fatal"))
			})
		})

		Context("Multi Issues Result with multi resources", func() {
			It("Aggregator with recoverable --> fatal, result: no requeue, fatal", func() {
				// add issues
				aggr.AddIssues(issue.New(training, fmt.Errorf(""), false, "", nil))
				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeFalse())
				Expect(len(aggr.GetIssuesList())).To(Equal(1))

				// add critical issue
				aggr.AddIssues(issue.New(training2, fmt.Errorf(""), true, "", nil))
				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeTrue())
				Expect(len(aggr.GetIssuesList())).To(Equal(2))

				// add issues - keep requeue and fail
				aggr.AddIssues(issue.New(training3, fmt.Errorf(""), false, "", nil))
				Expect(aggr.HasIssues()).To(BeTrue())
				Expect(aggr.HasCritical()).To(BeTrue())
				Expect(len(aggr.GetIssuesList())).To(Equal(3))
			})
		})
	})
})

func getMessage(kind, name, msg string) string {
	return fmt.Sprintf("[%s:%s] %s", kind, name, msg)
}
