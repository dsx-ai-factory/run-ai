package submitdistbc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Command", func() {
	var (
		instance     *config.Instance
		cmdHelper    helper.Cmd
		commandTexts *cmd.BaseCommandTexts
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		commandTexts = NewFrameworkText(api.Distributed, api.MPI)
	})

	Describe("NewCommand", func() {
		It("should create a new Command", func() {
			command := NewCommand(api.MPI, spec.NewDistributedBCSpec(api.MPI), nil, commandTexts, instance, cmdHelper)
			Expect(command).NotTo(BeNil())
			Expect(command.Framework).To(Equal(api.MPI))
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			distributedCmd := NewCommand(api.MPI, spec.NewDistributedBCSpec(api.MPI), nil, commandTexts, instance, cmdHelper)
			err := distributedCmd.RegisterFlags()
			Expect(err).To(BeNil())
		})
	})
})
