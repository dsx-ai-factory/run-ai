import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import WorkloadNetworkSection from "./workload-network-section.vue";
import type { INetworkSectionData } from "@/models/workload-details.model";

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

function mountSection(data: INetworkSectionData, searchTerm = "") {
  return mount(WorkloadNetworkSection, {
    props: { data, searchTerm },
    global: {
      plugins: [[Quasar, { plugins: {} }]],
      stubs: {
        RunaiDetailsPanelRow: {
          template: "<div class='row-stub'><span class='label'>{{ label }}</span><slot /></div>",
          props: ["label"],
        },
        RunaiDetailsPanelCopyableText: { template: "<span class='url-stub'>{{ text }}</span>", props: ["text"] },
        WorkloadTopologySection: {
          template: "<div class='topology-stub'>{{ topology?.name }}</div>",
          props: ["topology", "levels"],
        },
        QSeparator: true,
      },
    },
  });
}

const data: INetworkSectionData = {
  connections: [
    {
      name: "url-master",
      tool: "Jupyter",
      connectionType: "ExternalUrl",
      access: "All",
      port: 8888,
      url: "https://jupyter.example",
    },
    { name: "url-master-1", tool: "VSCode", connectionType: "ExternalUrl", access: "All", port: 8080 },
    { name: "comet-link", tool: "Comet", connectionType: "ExternalUrl", access: "Anyone", url: "https://comet.example" },
  ],
  topology: { name: "rack-aware", requiredPlacement: "rack" },
  topologyLevels: ["region", "zone", "rack"],
};

describe("workload-network-section search filtering", () => {
  it("renders all connections and topology when search is empty", () => {
    const wrapper = mountSection(data);
    expect(wrapper.text()).toContain("url-master");
    expect(wrapper.text()).toContain("url-master-1");
    expect(wrapper.text()).toContain("comet-link");
    expect(wrapper.find(".topology-stub").text()).toContain("rack-aware");
  });

  it("keeps a whole connection when any of its fields matches", () => {
    const wrapper = mountSection(data, "jupyter");
    expect(wrapper.text()).toContain("url-master");
    expect(wrapper.text()).toContain("Jupyter");
    expect(wrapper.text()).not.toContain("url-master-1");
    expect(wrapper.text()).not.toContain("comet-link");
  });

  it("matches against connection url", () => {
    const wrapper = mountSection(data, "comet.example");
    expect(wrapper.text()).toContain("comet-link");
    expect(wrapper.text()).not.toContain("url-master");
  });

  it("hides topology when nothing in topology matches", () => {
    const wrapper = mountSection(data, "jupyter");
    expect(wrapper.find(".topology-stub").exists()).toBe(false);
  });

  it("shows topology when the topology name matches", () => {
    const wrapper = mountSection(data, "rack-aware");
    expect(wrapper.find(".topology-stub").exists()).toBe(true);
  });

  it("shows topology when a topology level matches", () => {
    const wrapper = mountSection(data, "zone");
    expect(wrapper.find(".topology-stub").exists()).toBe(true);
  });

  it("renders nothing matching when the term has no hits anywhere", () => {
    const wrapper = mountSection(data, "zzz-nope");
    expect(wrapper.text()).not.toContain("url-master");
    expect(wrapper.text()).not.toContain("comet");
    expect(wrapper.find(".topology-stub").exists()).toBe(false);
  });
});
