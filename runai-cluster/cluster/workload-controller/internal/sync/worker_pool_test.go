package sync_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/sync"
)

var _ = Describe("WorkerPool", func() {
	Context("Execute", func() {
		It("should execute all tasks successfully", func() {
			tasks := []int{1, 2, 3, 4, 5}
			handler := func(ctx context.Context, task int) error {
				return nil
			}
			workerPool := sync.NewWorkerPool(2, handler, tasks)

			err := workerPool.Execute(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should stop executing tasks when context is cancelled", func() {
			tasks := []int{1, 2, 3, 4, 5}
			handler := func(ctx context.Context, task int) error {
				return nil
			}
			workerPool := sync.NewWorkerPool(2, handler, tasks)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := workerPool.Execute(ctx)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error when a task fails", func() {
			tasks := []int{1, 2, 3, 4, 5}
			handler := func(ctx context.Context, task int) error {
				if task == 3 {
					return errors.New("task failed")
				}
				return nil
			}
			workerPool := sync.NewWorkerPool(2, handler, tasks)

			err := workerPool.Execute(context.Background())
			Expect(err).To(HaveOccurred())
		})
	})
})
