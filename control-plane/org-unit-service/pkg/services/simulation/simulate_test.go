/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package simulation_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/simulation"
)

var maxDemand = math.Inf(1)

func department(id string, deserved, weight float64, priority int) simulation.OrgUnit {
	return simulation.OrgUnit{ID: id, Deserved: deserved, Limit: simulation.Unlimited, Weight: weight, Priority: priority}
}

func project(id, parentID string, deserved, weight, demand float64) simulation.OrgUnit {
	return simulation.OrgUnit{ID: id, ParentID: parentID, Deserved: deserved, Limit: simulation.Unlimited, Weight: weight,
		Priority: 100, Demand: demand}
}

var _ = Describe("Simulate", func() {
	It("gives each department its quota, then splits the rest by rank and weight", func() {
		departments := []simulation.OrgUnit{department("a", 10, 1, 100), department("b", 20, 3, 100)}
		projects := []simulation.OrgUnit{project("pa", "a", 0, 1, maxDemand), project("pb", "b", 0, 1, maxDemand)}

		shares, _ := simulation.Simulate(70, departments, projects)

		Expect(shares["a"]).To(Equal(simulation.Share{FairShare: 20, Guaranteed: 10, OverQuota: 10, Unmet: 50}))
		Expect(shares["b"]).To(Equal(simulation.Share{FairShare: 50, Guaranteed: 20, OverQuota: 30, Unmet: 20}))
	})

	It("splits a department's fair share between its projects", func() {
		departments := []simulation.OrgUnit{department("a", 0, 1, 100)}
		projects := []simulation.OrgUnit{project("p1", "a", 4, 1, maxDemand), project("p2", "a", 0, 3, maxDemand)}

		_, shares := simulation.Simulate(20, departments, projects)

		Expect(shares["p1"]).To(Equal(simulation.Share{FairShare: 8, Guaranteed: 4, OverQuota: 4, Unmet: 12}))
		Expect(shares["p2"]).To(Equal(simulation.Share{FairShare: 12, Guaranteed: 0, OverQuota: 12, Unmet: 8}))
	})

	It("gives what a department's projects don't ask for to its siblings", func() {
		departments := []simulation.OrgUnit{department("a", 0, 1, 100), department("b", 0, 1, 100)}
		projects := []simulation.OrgUnit{project("pa", "a", 0, 1, 10), project("pb", "b", 0, 1, maxDemand)}

		departmentShares, projectShares := simulation.Simulate(100, departments, projects)

		Expect(departmentShares["a"].FairShare).To(Equal(10.0))
		Expect(departmentShares["b"].FairShare).To(Equal(90.0))
		Expect(projectShares["pa"]).To(Equal(simulation.Share{FairShare: 10, OverQuota: 10}))
	})

	It("gives nothing to a department without projects", func() {
		departments := []simulation.OrgUnit{department("a", 10, 1, 100), department("b", 0, 1, 100)}
		projects := []simulation.OrgUnit{project("pb", "b", 0, 1, maxDemand)}

		shares, _ := simulation.Simulate(40, departments, projects)

		Expect(shares["a"].FairShare).To(BeZero())
		Expect(shares["b"].FairShare).To(Equal(40.0))
	})

	It("serves higher ranks first", func() {
		departments := []simulation.OrgUnit{department("high", 0, 1, 210), department("low", 0, 9, 50)}
		projects := []simulation.OrgUnit{project("ph", "high", 0, 1, 30), project("pl", "low", 0, 1, maxDemand)}

		shares, _ := simulation.Simulate(40, departments, projects)

		Expect(shares["high"].FairShare).To(Equal(30.0))
		Expect(shares["low"].FairShare).To(Equal(10.0))
	})

	It("stops an org unit at its limit and reports unmet demand against it", func() {
		limited := project("p1", "a", 0, 1, maxDemand)
		limited.Limit = 5
		departments := []simulation.OrgUnit{department("a", 0, 1, 100)}
		projects := []simulation.OrgUnit{limited, project("p2", "a", 0, 1, maxDemand)}

		_, shares := simulation.Simulate(20, departments, projects)

		Expect(shares["p1"]).To(Equal(simulation.Share{FairShare: 5, OverQuota: 5}))
		Expect(shares["p2"].FairShare).To(Equal(15.0))
	})
})
