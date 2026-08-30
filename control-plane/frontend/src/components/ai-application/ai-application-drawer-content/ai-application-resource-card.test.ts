import { describe, expect, it } from "vitest";
import { shallowMount } from "@vue/test-utils";
import AiApplicationResourceCard from "./ai-application-resource-card.vue";
import {
  EResourceCardType,
  RESOURCE_TOP_HANDLE_ID,
  WORKLOAD_BOTTOM_HANDLE_ID,
  WORKLOAD_TOP_HANDLE_ID,
} from "./ai-application-diagram.util";

const createWrapper = (type: EResourceCardType, expanded = false, searchTerm = "") =>
  shallowMount(AiApplicationResourceCard, {
    props: {
      name: "resource",
      type,
      phase: type === EResourceCardType.Workload ? "Running" : null,
      expanded,
      searchTerm,
    },
    global: {
      stubs: {
        Handle: {
          props: ["id", "type"],
          template: '<span class="handle-stub" :data-id="id" :data-type="type" />',
        },
        RunaiHighlightedText: false,
        RunaiSvgIcon: true,
        WorkloadStatusDot: true,
      },
    },
  });

describe("AiApplicationResourceCard", () => {
  it("renders top and bottom source handles for workload connections", () => {
    const wrapper = createWrapper(EResourceCardType.Workload);
    const handles = wrapper.findAll(".handle-stub");

    expect(handles.map((handle) => handle.attributes("data-id"))).toEqual([
      WORKLOAD_TOP_HANDLE_ID,
      WORKLOAD_BOTTOM_HANDLE_ID,
    ]);
    expect(handles.every((handle) => handle.attributes("data-type") === "source")).toBe(true);
  });

  it("renders a top target handle for connected storage resources", () => {
    const wrapper = createWrapper(EResourceCardType.Pvc);
    const handles = wrapper.findAll(".handle-stub");

    expect(handles.map((handle) => handle.attributes("data-id"))).toEqual([RESOURCE_TOP_HANDLE_ID]);
    expect(handles.every((handle) => handle.attributes("data-type") === "target")).toBe(true);
  });

  it("preserves expanded workload card state", () => {
    const wrapper = createWrapper(EResourceCardType.Workload, true);

    expect(wrapper.find(".resource-card").classes()).toContain("expanded");
    expect(wrapper.find(".expand-chevron").classes()).toContain("expanded");
  });

  it("highlights matching text in the resource name", () => {
    const wrapper = createWrapper(EResourceCardType.Workload, false, "sour");

    expect(wrapper.find("mark.search-highlight").text()).toBe("sour");
    expect(wrapper.find(".card-name").text()).toBe("resource");
  });
});
