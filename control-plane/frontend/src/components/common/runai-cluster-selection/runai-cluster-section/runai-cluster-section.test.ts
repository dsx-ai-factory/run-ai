import { describe, vi, it, expect } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { useClusterStore } from "@/stores/cluster.store";
import { RunaiClusterDropdown } from "@/components/common/runai-cluster-selection/runai-cluster-dropdown";
import { unitTestUtil } from "@/utils/unit-test.util";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { ERunaiClusterSectionEvents } from "./runai-cluster-section.vue";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
let wrapper: VueWrapper<InstanceType<typeof RunaiClusterSection>>;
interface IClusterSelectionSectionProps {
  clusterId: string;
  entityName: string;
  disabled?: boolean;
  knativeRequired?: boolean;
  clusterConnectivityRequired?: boolean;
}
describe("cluster-selection-section", () => {
  const factory = (props: IClusterSelectionSectionProps) => {
    wrapper = mount(RunaiClusterSection, {
      global: {
        plugins: [Quasar, pinia],
      },
      props,
    });
    const clusterStore = useClusterStore(pinia);
    clusterStore.$patch({
      clusters: [
        { uuid: "1", name: "Cluster One" },
        { uuid: "2", name: "Cluster Two" },
      ],
    });
  };

  it("renders correctly with props", () => {
    factory({ clusterId: "1", entityName: "data-source", disabled: false });
    expect(wrapper.text()).toContain("Set under which cluster to create this data-source");
    expect(wrapper.find('[aid="cluster-selection-section"]').exists()).toBe(true);
  });

  it("displays the correct cluster name in the subheader", async () => {
    factory({ clusterId: "1", entityName: "data-source", disabled: false });

    await unitTestUtil.closeExpansionItem(wrapper as VueWrapper);

    expect(wrapper.text()).toContain("Cluster One");
  });

  it("displays tooltip when disabled and tooltip text is provided", () => {
    factory({ clusterId: "1", entityName: "data-source", disabled: true });
    expect(wrapper.findComponent(RunaiClusterDropdown).vm.displayTooltip).toBe(true);
  });

  it("does not display tooltip when not disabled", () => {
    factory({ clusterId: "1", entityName: "data-source", disabled: false });
    expect(wrapper.findComponent(RunaiClusterDropdown).vm.displayTooltip).toBe(false);
  });

  it("emits update:cluster-id event when the selected cluster is changed", async () => {
    factory({ clusterId: "", entityName: "data-source", disabled: false });
    wrapper.vm.onClusterChanged("2", { disabled: false });
    expect(wrapper.emitted(ERunaiClusterSectionEvents.UpdateClusterId)).toBeTruthy();
    expect(wrapper.emitted(ERunaiClusterSectionEvents.UpdateClusterId)?.[0]).toEqual(["2", { disabled: false }]);
  });
});
