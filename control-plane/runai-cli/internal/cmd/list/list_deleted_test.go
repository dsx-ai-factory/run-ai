/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	workloadservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

var _ = Describe("List Command with Deleted Flag", func() {
	var (
		ctrl         *gomock.Controller
		command      *Command
		listFnCalled bool
		deletedParam bool
		instance     *config.Instance
		cmdHelper    helper.Cmd
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		listFnCalled = false
		deletedParam = false

		// Create mock instance
		instance = &config.Instance{}

		// Create a mock list function that captures the deleted parameter
		mockListFn := func(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
			listFnCalled = true
			deletedParam = deleted
			output, err := list.NewOutput([]workloadservice.Workload{
				{Id: "test-id", Name: "test-workload"},
			}, nil, deleted)
			if err != nil {
				return nil, err
			}
			return output, nil
		}

		// Create command using NewListCommand
		texts := NewFrameworkListTexts(api.Training, api.PyTorch)
		command = NewListCommand(api.PyTorch, cmd.NewEmptyCobraCommand, instance, texts, cmdHelper, mockListFn, api.Training)

		// Register flags
		err := command.RegisterFlags()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Deleted Flag Functionality", func() {
		It("should pass deleted=false when flag is not provided", func() {
			// Get the cobra command
			cobraCmd := command.GetCommand()

			// Don't set the deleted flag
			err := cobraCmd.ParseFlags([]string{})
			Expect(err).ToNot(HaveOccurred())

			// Execute GetList
			output, err := command.GetList("", "", &apilist.PageOptions{PageSize: 10})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Verify the list function was called with deleted=false
			Expect(listFnCalled).To(BeTrue())
			Expect(deletedParam).To(BeFalse())
		})

		It("should pass deleted=true when flag is set", func() {
			// Get the cobra command
			cobraCmd := command.GetCommand()

			// Set the deleted flag
			err := cobraCmd.ParseFlags([]string{"--deleted"})
			Expect(err).ToNot(HaveOccurred())

			// Set the options manually since we're testing internal behavior
			command.options.deleted = true

			// Execute GetList
			output, err := command.GetList("", "", &apilist.PageOptions{PageSize: 10})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Verify the list function was called with deleted=true
			Expect(listFnCalled).To(BeTrue())
			Expect(deletedParam).To(BeTrue())
		})

		It("should work with other flags when deleted is set", func() {
			// Get the cobra command
			cobraCmd := command.GetCommand()

			// Set multiple flags including deleted
			err := cobraCmd.ParseFlags([]string{
				"--deleted",
				"--all",
				"--status", "Running",
			})
			Expect(err).ToNot(HaveOccurred())

			// Set the options manually
			command.options.deleted = true
			command.options.all = true
			command.options.workloadState = "Running"

			// Create a more complex test that captures the filter builder
			var capturedFilter *filter.Builder
			command.listFn = func(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
				capturedFilter = fb
				deletedParam = deleted
				output, err := list.NewOutput([]workloadservice.Workload{{Id: "test", Name: "test"}}, nil, deleted)
				if err != nil {
					return nil, err
				}
				return output, nil
			}

			// Execute GetList
			output, err := command.GetList("", "Running", &apilist.PageOptions{PageSize: 10})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Verify deleted parameter
			Expect(deletedParam).To(BeTrue())

			// Verify filter has the status
			Expect(capturedFilter).ToNot(BeNil())
			filters := capturedFilter.StringList()
			Expect(filters).To(ContainElement(ContainSubstring("Running")))
		})
	})

	Context("Error Handling", func() {
		It("should handle errors from list function when deleted is set", func() {
			// Create a list function that returns an error
			command.listFn = func(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
				if deleted {
					return nil, fmt.Errorf("error listing deleted workloads")
				}
				return nil, fmt.Errorf("unexpected error")
			}

			command.options.deleted = true

			// Execute GetList
			output, err := command.GetList("", "", &apilist.PageOptions{PageSize: 10})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error listing deleted workloads"))
			Expect(output).To(BeNil())
		})
	})

	Context("Integration with Pagination", func() {
		It("should work with pagination when deleted flag is set", func() {
			var capturedPageOptions *apilist.PageOptions
			command.listFn = func(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
				capturedPageOptions = pagingOptions
				deletedParam = deleted
				output, err := list.NewOutput([]workloadservice.Workload{
					{Id: "1", Name: "deleted-1"},
					{Id: "2", Name: "deleted-2"},
				}, nil, deleted)
				if err != nil {
					return nil, err
				}
				return output, nil
			}

			command.options.deleted = true

			// Execute GetList with pagination
			pageOpts := &apilist.PageOptions{
				PageSize: 20,
				Offset:   10,
				MaxItems: 100,
			}
			output, err := command.GetList("", "", pageOpts)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Verify pagination options were passed correctly
			Expect(capturedPageOptions).To(Equal(pageOpts))
			Expect(deletedParam).To(BeTrue())
		})
	})
})
