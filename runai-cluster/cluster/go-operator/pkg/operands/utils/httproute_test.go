/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("HTTPRouteAccepted", func() {
	It("should return true when Accepted condition is True", func() {
		status := gwv1.HTTPRouteStatus{
			RouteStatus: gwv1.RouteStatus{
				Parents: []gwv1.RouteParentStatus{{
					Conditions: []metav1.Condition{{
						Type:   string(gwv1.RouteConditionAccepted),
						Status: metav1.ConditionTrue,
					}},
				}},
			},
		}
		accepted, msg := HTTPRouteAccepted(status)
		Expect(accepted).To(BeTrue())
		Expect(msg).To(BeEmpty())
	})

	It("should return false when Accepted condition is False", func() {
		status := gwv1.HTTPRouteStatus{
			RouteStatus: gwv1.RouteStatus{
				Parents: []gwv1.RouteParentStatus{{
					Conditions: []metav1.Condition{{
						Type:    string(gwv1.RouteConditionAccepted),
						Status:  metav1.ConditionFalse,
						Message: "gateway not found",
					}},
				}},
			},
		}
		accepted, msg := HTTPRouteAccepted(status)
		Expect(accepted).To(BeFalse())
		Expect(msg).To(Equal("gateway not found"))
	})

	It("should return false when no parents exist", func() {
		status := gwv1.HTTPRouteStatus{}
		accepted, msg := HTTPRouteAccepted(status)
		Expect(accepted).To(BeFalse())
		Expect(msg).To(ContainSubstring("no gateway"))
	})

	It("should return false when parents exist but no Accepted condition", func() {
		status := gwv1.HTTPRouteStatus{
			RouteStatus: gwv1.RouteStatus{
				Parents: []gwv1.RouteParentStatus{{
					Conditions: []metav1.Condition{{
						Type:   string(gwv1.RouteConditionResolvedRefs),
						Status: metav1.ConditionTrue,
					}},
				}},
			},
		}
		accepted, msg := HTTPRouteAccepted(status)
		Expect(accepted).To(BeFalse())
		Expect(msg).To(ContainSubstring("accepted"))
	})
})
