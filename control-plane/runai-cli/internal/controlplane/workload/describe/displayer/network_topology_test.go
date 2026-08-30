package displayer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("NetworkTopologyDisplayer", func() {
	It("should return correct subject", func() {
		section := displayer.NewNetworkTopologySection(&displayer.NetworkTopology{})
		Expect(section.Displayer.Subject()).To(Equal("Network Topology"))
	})

	It("should return correct columns", func() {
		section := displayer.NewNetworkTopologySection(&displayer.NetworkTopology{})
		columns := section.Displayer.Columns()
		Expect(columns).To(HaveLen(4))
		Expect(columns[0].Title).To(Equal("Network topology name"))
		Expect(columns[1].Title).To(Equal("Required Placement"))
		Expect(columns[2].Title).To(Equal("Preferred Placement"))
		Expect(columns[3].Title).To(Equal("Actual Placement"))
	})

	It("should return nil data when network topology is nil", func() {
		section := displayer.NewNetworkTopologySection(nil)
		Expect(section.Displayer.Data()).To(BeNil())
	})

	It("should return one row with all fields populated", func() {
		nt := &displayer.NetworkTopology{
			Name:               ptr.To("mnnvl"),
			RequiredPlacement:  ptr.To("rack"),
			PreferredPlacement: ptr.To("block"),
			ActualPlacement:    ptr.To("network.topology.nvidia.com/rack (rack-24)"),
		}
		section := displayer.NewNetworkTopologySection(nt)
		data := section.Displayer.Data()
		Expect(data).To(HaveLen(1))
		Expect(data[0][0]).To(Equal("mnnvl"))
		Expect(data[0][1]).To(Equal("rack"))
		Expect(data[0][2]).To(Equal("block"))
		Expect(data[0][3]).To(Equal("network.topology.nvidia.com/rack (rack-24)"))
	})

	It("should handle partial fields", func() {
		nt := &displayer.NetworkTopology{
			Name:              ptr.To("mnnvl"),
			RequiredPlacement: ptr.To("rack"),
		}
		section := displayer.NewNetworkTopologySection(nt)
		data := section.Displayer.Data()
		Expect(data).To(HaveLen(1))
		Expect(data[0][0]).To(Equal("mnnvl"))
		Expect(data[0][1]).To(Equal("rack"))
		Expect(data[0][2]).To(Equal(""))
		Expect(data[0][3]).To(Equal(""))
	})
})

var _ = Describe("FormatActualPlacement", func() {
	It("should format level and value", func() {
		result := displayer.FormatActualPlacement(ptr.To("network.topology.nvidia.com/rack"), ptr.To("rack-24"))
		Expect(result).ToNot(BeNil())
		Expect(*result).To(Equal("network.topology.nvidia.com/rack (rack-24)"))
	})

	It("should return level only when value is empty", func() {
		result := displayer.FormatActualPlacement(ptr.To("network.topology.nvidia.com/rack"), ptr.To(""))
		Expect(result).ToNot(BeNil())
		Expect(*result).To(Equal("network.topology.nvidia.com/rack"))
	})

	It("should return level only when value is nil", func() {
		result := displayer.FormatActualPlacement(ptr.To("network.topology.nvidia.com/rack"), nil)
		Expect(result).ToNot(BeNil())
		Expect(*result).To(Equal("network.topology.nvidia.com/rack"))
	})

	It("should return nil when level is nil", func() {
		result := displayer.FormatActualPlacement(nil, ptr.To("rack-24"))
		Expect(result).To(BeNil())
	})

	It("should return nil when level is empty", func() {
		result := displayer.FormatActualPlacement(ptr.To(""), ptr.To("rack-24"))
		Expect(result).To(BeNil())
	})
})
