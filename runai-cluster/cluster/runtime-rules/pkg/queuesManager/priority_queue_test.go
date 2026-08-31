/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"sync"
	"time"
)

var _ = Describe("Priority queue test", Ordered, func() {
	var (
		pq       *queuesManager.PriorityQueue
		wl1, wl2 queuesManager.WorkloadTask
	)

	BeforeEach(func() {
		pq = queuesManager.NewPriorityQueue()
		wl1 = queuesManager.WorkloadTask{WorkloadName: "wl01", NextEventAt: time.Now()}
		wl2 = queuesManager.WorkloadTask{WorkloadName: "wl02", NextEventAt: time.Now().Add(time.Hour)}

		Expect(pq.Len()).To(Equal(0))
		pq.Push(&wl1)
		Expect(pq.Len()).To(Equal(1))
		pq.Push(&wl2)
		Expect(pq.Len()).To(Equal(2))
	})

	Context("Queue basics", func() {
		It("push non workload object", func() {
			pq = queuesManager.NewPriorityQueue()
			Expect(pq.Len()).To(Equal(0))
			pq.Push("this should not be pushed")
			Expect(pq.Len()).To(Equal(0))
		})

		It("pop workload object", func() {
			pop := pq.Pop()
			wtTmp, ok := pop.(*queuesManager.WorkloadTask)
			Expect(ok).To(BeTrue())
			Expect(pq.Len()).To(Equal(1))
			Expect(wtTmp.WorkloadName).To(Equal(wl2.WorkloadName))
			Expect(wtTmp.NextEventAt).To(Equal(wl2.NextEventAt))

			pop = pq.Pop()
			Expect(pq.Len()).To(Equal(0))
			wtTmp, ok = pop.(*queuesManager.WorkloadTask)
			Expect(ok).To(BeTrue())
			Expect(wtTmp.WorkloadName).To(Equal(wl1.WorkloadName))
			Expect(wtTmp.NextEventAt).To(Equal(wl1.NextEventAt))

			// empty queue will return nil
			pop = pq.Pop()
			Expect(pop).To(BeNil())
		})

		It("switch workload object", func() {
			pq.Swap(0, 1)
			pop := pq.Pop()
			wtTmp, ok := pop.(*queuesManager.WorkloadTask)
			Expect(ok).To(BeTrue())
			Expect(pq.Len()).To(Equal(1))
			Expect(wtTmp.WorkloadName).To(Equal(wl1.WorkloadName))
			Expect(wtTmp.NextEventAt).To(Equal(wl1.NextEventAt))

			pop = pq.Pop()
			Expect(pq.Len()).To(Equal(0))
			wtTmp, ok = pop.(*queuesManager.WorkloadTask)
			Expect(ok).To(BeTrue())
			Expect(wtTmp.WorkloadName).To(Equal(wl2.WorkloadName))
			Expect(wtTmp.NextEventAt).To(Equal(wl2.NextEventAt))
		})

		It("less workload object", func() {
			Expect(pq.Less(0, 1)).To(BeTrue())
			pq.Swap(0, 1)
			Expect(pq.Less(0, 1)).To(BeFalse())
		})

		It("Clean queue and mutual exclusion check", func() {
			wg := sync.WaitGroup{}
			wg.Add(4)
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					wl := queuesManager.WorkloadTask{
						WorkloadName: fmt.Sprintf("queue-%d-%d", 1, i),
						NextEventAt:  time.Now()}
					pq.Push(&wl)
				}
			}()
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					wl := queuesManager.WorkloadTask{
						WorkloadName: fmt.Sprintf("queue-%d-%d", 2, i),
						NextEventAt:  time.Now()}
					pq.Push(&wl)
				}
			}()
			go func() {
				defer wg.Done()
				for i := 0; i < 25; i++ {
					for {
						pop := pq.Pop()
						if pop != nil {
							break
						}
					}
				}
			}()
			go func() {
				defer wg.Done()
				for i := 0; i < 25; i++ {
					for {
						pop := pq.Pop()
						if pop != nil {
							break
						}
					}
				}
			}()
			wg.Wait()

			Expect(pq.Len()).To(Equal(152))
			pq.Clean()
			Expect(pq.Len()).To(Equal(0))
		})
	})

	Context("Queue errors", func() {
		It("switch out of scope index", func() {
			Expect(func() {
				pq.Swap(0, 2)
			}).To(Panic())
		})

		It("less out of scope index", func() {
			Expect(func() {
				pq.Less(0, 2)
			}).To(Panic())
		})
	})
})
