import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { NodePoolNetworkSection } from "@/components/node-pools/node-pool-edit-form/node-pool-network-section/";
import { NodepoolGPUNetworkAccelerationDetection } from "@/swagger-models/cluster-service-client";
import { Quasar } from "quasar";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { NetworkTopologySelection } from "@/components/node-pools/node-pool-edit-form/node-pool-network-section/network-topology-selection/";
import { mockNetworkTopologies } from "./node-pool-network-section.mock";

let wrapper: VueWrapper<InstanceType<typeof NodePoolNetworkSection>>;

describe("NodePoolNetworkSection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(clusterService, "getNetworkTopologies").mockResolvedValue(mockNetworkTopologies);
  });

  const mountComponent = (props = {}) => {
    return mount(NodePoolNetworkSection, {
      global: {
        plugins: [Quasar],
        mocks: {
          $route: {
            query: "",
          },
        },
        directives: {
          clickOutside: vi.fn(),
          permission: vi.fn(),
        },
      },
      props: {
        ...props,
        clusterId: "123",
      },
    });
  };

  it("renders correctly with default props", () => {
    wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
  });

  describe("props", () => {
    it("initializes with the correct detection method default state", async () => {
      wrapper = mountComponent();
      await wrapper.vm.$nextTick();
      const radioOptions = wrapper.findComponent(RunaiRadioOptions);
      expect(radioOptions.props("modelValue")).toBe(NodepoolGPUNetworkAccelerationDetection.Auto);
      expect(wrapper.vm.discoveryMethod).toBe(NodepoolGPUNetworkAccelerationDetection.Auto);
      expect(wrapper.emitted("update:gpu-network-acceleration-detection")).toBeTruthy();
      expect(wrapper.emitted("update:gpu-network-acceleration-detection")?.[0]).toEqual([
        NodepoolGPUNetworkAccelerationDetection.Auto,
      ]);
    });

    it("initializes with the correct detection method", async () => {
      wrapper = mountComponent({ gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Use });
      await wrapper.vm.$nextTick();
      const radioOptions = wrapper.findComponent(RunaiRadioOptions);
      expect(radioOptions.props("modelValue")).toBe(NodepoolGPUNetworkAccelerationDetection.Use);
      expect(wrapper.vm.discoveryMethod).toBe(NodepoolGPUNetworkAccelerationDetection.Use);
    });
  });

  describe("emits", () => {
    it("emits update:gpu-network-acceleration-detection on mount if not set", async () => {
      wrapper = mountComponent();
      await flushPromises();
      expect(wrapper.emitted("update:gpu-network-acceleration-detection")).toBeTruthy();
      expect(wrapper.emitted("update:gpu-network-acceleration-detection")?.[0]).toEqual([
        NodepoolGPUNetworkAccelerationDetection.Auto,
      ]);
    });

    it("emits update:gpu-network-acceleration-detection when detection method changes", async () => {
      wrapper = mountComponent({ gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Use });
      const radioOptions = wrapper.findComponent(RunaiRadioOptions);
      await radioOptions.vm.$emit("update:model-value", NodepoolGPUNetworkAccelerationDetection.DontUse);
      expect(wrapper.emitted("update:gpu-network-acceleration-detection")).toBeTruthy();
    });
  });

  describe("template rendering", () => {
    it("should render the label when the discovery method is auto detect", async () => {
      wrapper = mountComponent({ gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Auto });
      await wrapper.vm.$nextTick();
      const labelInput = wrapper.find('[aid="node-pool-network-label-input"]');
      expect(labelInput.exists()).toBe(true);
    });

    it("should render the label when the discovery method is mnnvl is present (use)", async () => {
      wrapper = mountComponent({ gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Use });
      await wrapper.vm.$nextTick();
      const labelInput = wrapper.find('[aid="node-pool-network-label-input"]');
      expect(labelInput.exists()).toBe(true);
    });

    it("should not render the label when the discovery method is mnnvl is not present (dont use)", async () => {
      wrapper = mountComponent({ gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.DontUse });
      await wrapper.vm.$nextTick();
      const labelInput = wrapper.find('[aid="node-pool-network-label-input"]');
      expect(labelInput.exists()).toBe(false);
    });
  });

  describe("network topology selection", () => {
    it("renders network topology selection component", async () => {
      wrapper = mountComponent();
      await flushPromises();
      const topologySelection = wrapper.findComponent(NetworkTopologySelection);
      expect(topologySelection.exists()).toBe(true);
    });

    it("passes networkTopologyId to network topology selection", async () => {
      wrapper = mountComponent({ networkTopologyId: "topology-1" });
      await flushPromises();
      const topologySelection = wrapper.findComponent(NetworkTopologySelection);
      expect(topologySelection.props("networkTopologyId")).toBe("topology-1");
    });

    it("emits update:network-topology-id when topology is selected", async () => {
      wrapper = mountComponent();
      await flushPromises();
      const topologySelection = wrapper.findComponent(NetworkTopologySelection);

      await topologySelection.vm.$emit("select", "topology-1");

      expect(wrapper.emitted("update:network-topology-id")).toBeTruthy();
      expect(wrapper.emitted("update:network-topology-id")?.[0]).toEqual(["topology-1"]);
    });

    it("emits update:network-topology-id with null when topology is cleared", async () => {
      wrapper = mountComponent({ networkTopologyId: "topology-1" });
      await flushPromises();
      const topologySelection = wrapper.findComponent(NetworkTopologySelection);

      await topologySelection.vm.$emit("select", null);

      expect(wrapper.emitted("update:network-topology-id")).toBeTruthy();
      expect(wrapper.emitted("update:network-topology-id")?.[0]).toEqual([null]);
    });

    it("loads network topologies on mount", async () => {
      wrapper = mountComponent();
      await flushPromises();
      expect(clusterService.getNetworkTopologies).toHaveBeenCalledWith(["clusterId==123"]);
    });

    it("reloads network topologies when clusterId changes", async () => {
      wrapper = mountComponent({ clusterId: "cluster-1" });
      await flushPromises();

      await wrapper.setProps({ clusterId: "cluster-2" });
      await flushPromises();

      expect(clusterService.getNetworkTopologies).toHaveBeenCalledTimes(2);
      expect(clusterService.getNetworkTopologies).toHaveBeenLastCalledWith(["clusterId==cluster-2"]);
    });
  });

  describe("network topology selection clearable functionality", () => {
    it("handles clear event from network topology selection", async () => {
      wrapper = mountComponent({ networkTopologyId: "topology-1" });
      await flushPromises();
      const topologySelection = wrapper.findComponent(NetworkTopologySelection);
      const runaiSelect = topologySelection.findComponent({ name: "runai-select" });

      await runaiSelect.vm.$emit("update:model-value", null);

      expect(wrapper.emitted("update:network-topology-id")).toBeTruthy();
      expect(wrapper.emitted("update:network-topology-id")?.[0]).toEqual([null]);
    });

    it("shows clear button when topology is selected", async () => {
      wrapper = mountComponent({ networkTopologyId: "topology-1" });
      await flushPromises();
      const topologySelection = wrapper.findComponent(NetworkTopologySelection);
      const runaiSelect = topologySelection.findComponent({ name: "runai-select" });

      expect(runaiSelect.props("modelValue")).toBe("topology-1");
    });
  });
});
