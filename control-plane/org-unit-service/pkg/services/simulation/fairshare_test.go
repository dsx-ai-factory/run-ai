/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package simulation_test

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/simulation"
)

type kaiCase struct {
	Capacity float64 `json:"capacity"`
	Queues   []struct {
		ID              string  `json:"id"`
		Deserved        float64 `json:"deserved"`
		MaxAllowed      float64 `json:"maxAllowed"`
		OverQuotaWeight float64 `json:"overQuotaWeight"`
		Priority        int     `json:"priority"`
		Request         float64 `json:"request"`
		CreatedAt       int64   `json:"createdAt"`
	} `json:"queues"`
	Expected map[string]float64 `json:"expected"`
}

var _ = Describe("DivideResource", func() {
	It("matches the results of KAI-scheduler's own resource division", func() {
		raw, err := os.ReadFile("testdata/kai_division_cases.json")
		Expect(err).ToNot(HaveOccurred())
		var cases []kaiCase
		Expect(json.Unmarshal(raw, &cases)).To(Succeed())
		Expect(cases).ToNot(BeEmpty())

		for i, c := range cases {
			queues := make([]simulation.Queue, 0, len(c.Queues))
			for _, q := range c.Queues {
				queues = append(queues, simulation.Queue{ID: q.ID, Deserved: q.Deserved, MaxAllowed: q.MaxAllowed,
					OverQuotaWeight: q.OverQuotaWeight, Priority: q.Priority, Request: q.Request, CreatedAt: q.CreatedAt})
			}
			got := simulation.DivideResource(c.Capacity, queues)
			for id, want := range c.Expected {
				Expect(got[id]).To(BeNumerically("~", want, 1e-9), "case %d queue %s", i, id)
			}
		}
	})

	It("gives the over quota to the highest rank first", func() {
		got := simulation.DivideResource(12, []simulation.Queue{
			{ID: "a", Priority: 210, OverQuotaWeight: 1, Request: 10, MaxAllowed: simulation.Unlimited},
			{ID: "b", Priority: 210, OverQuotaWeight: 3, Request: 10, MaxAllowed: simulation.Unlimited},
			{ID: "c", Priority: 150, Deserved: 2, OverQuotaWeight: 1, Request: 10, MaxAllowed: simulation.Unlimited},
			{ID: "d", Priority: 50, Deserved: 4, OverQuotaWeight: 1, Request: 10, MaxAllowed: simulation.Unlimited},
		})
		Expect(got).To(Equal(map[string]float64{"a": 2, "b": 4, "c": 2, "d": 4}))
	})

	It("keeps rounding leftovers with a higher rank", func() {
		got := simulation.DivideResource(6, []simulation.Queue{
			{ID: "high-1", Priority: 2, Deserved: 1, OverQuotaWeight: 1, Request: 5, MaxAllowed: simulation.Unlimited, CreatedAt: 1},
			{ID: "high-2", Priority: 2, Deserved: 1, OverQuotaWeight: 1, Request: 5, MaxAllowed: simulation.Unlimited, CreatedAt: 2},
			{ID: "low", Priority: 1, Deserved: 1, OverQuotaWeight: 1, Request: 5, MaxAllowed: simulation.Unlimited, CreatedAt: 3},
		})
		Expect(got).To(Equal(map[string]float64{"high-1": 3, "high-2": 2, "low": 1}))
	})
})
