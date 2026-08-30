import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import WorkloadEnvironmentSection from "./workload-environment-section.vue";
import type { IEnvironmentSectionData } from "@/models/workload-details.model";

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

function mountSection(data: IEnvironmentSectionData, searchTerm = "") {
  return mount(WorkloadEnvironmentSection, {
    props: { data, searchTerm },
    global: {
      plugins: [[Quasar, { plugins: {} }]],
      stubs: {
        RunaiDetailsPanelRow: {
          template: "<div class='row-stub'><span>{{ label }}</span><slot /></div>",
          props: ["label"],
        },
        RunaiDetailsPanelLink: { template: "<a class='link-stub'>{{ link?.text }}</a>", props: ["link"] },
        RunaiInlineCodeChip: { template: "<code class='code-stub'>{{ text }}</code>", props: ["text"] },
        QSeparator: true,
      },
    },
  });
}

const fullData: IEnvironmentSectionData = {
  items: [
    { label: "Image URL", value: "nvcr.io/jupyter:latest" },
    { label: "Command", value: "python train.py", isCode: true },
  ],
  environmentVariables: ["DEBUG=true", "LOG_LEVEL=info", "NODE_ENV=prod"],
  applicationParameters: [{ label: "epochs", value: "10" }],
  labels: ["team=ml", "owner=alice"],
  annotations: ["run.ai/created-by=ui", "run.ai/category=Train"],
};

describe("workload-environment-section search filtering", () => {
  it("renders every collection when there is no search term", () => {
    const wrapper = mountSection(fullData);
    const codes = wrapper.findAll(".code-stub").map((c) => c.text());
    expect(codes).toContain("DEBUG=true");
    expect(codes).toContain("LOG_LEVEL=info");
    expect(codes).toContain("team=ml");
    expect(codes).toContain("run.ai/created-by=ui");
    expect(wrapper.text()).toContain("Image URL");
    expect(wrapper.text()).toContain("epochs");
  });

  it("filters environment variables by term", () => {
    const wrapper = mountSection(fullData, "log");
    const codes = wrapper.findAll(".code-stub").map((c) => c.text());
    expect(codes).toContain("LOG_LEVEL=info");
    expect(codes).not.toContain("DEBUG=true");
    expect(codes).not.toContain("NODE_ENV=prod");
  });

  it("filters labels by term", () => {
    const wrapper = mountSection(fullData, "owner");
    const codes = wrapper.findAll(".code-stub").map((c) => c.text());
    expect(codes).toContain("owner=alice");
    expect(codes).not.toContain("team=ml");
  });

  it("filters annotations by term", () => {
    const wrapper = mountSection(fullData, "category");
    const codes = wrapper.findAll(".code-stub").map((c) => c.text());
    expect(codes).toContain("run.ai/category=Train");
    expect(codes).not.toContain("run.ai/created-by=ui");
  });

  it("filters items rows by term", () => {
    const wrapper = mountSection(fullData, "image");
    expect(wrapper.text()).toContain("Image URL");
    expect(wrapper.text()).not.toContain("Command");
  });

  it("filters application parameters by term", () => {
    const wrapper = mountSection(fullData, "epochs");
    expect(wrapper.text()).toContain("epochs");
  });

  it("hides a collection's row when none of its entries match", () => {
    const wrapper = mountSection(fullData, "zzz");
    expect(wrapper.text()).not.toContain("Environment variables");
    expect(wrapper.text()).not.toContain("Labels");
    expect(wrapper.text()).not.toContain("Annotations");
  });
});
