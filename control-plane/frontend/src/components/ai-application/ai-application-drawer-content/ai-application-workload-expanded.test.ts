import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import AiApplicationWorkloadExpanded from "./ai-application-workload-expanded.vue";
import type { Workload } from "@/swagger-models/workloads-client";

const createWorkload = (overrides: Partial<Workload> = {}): Workload =>
  ({
    id: "workload-1",
    name: "test-workload",
    type: "NIMService",
    clusterId: "cluster-1",
    workloadRequestedResources: null,
    podsRequestedResources: null,
    allocatedResources: null,
    ...overrides,
  }) as Workload;

const createWrapper = (workload: Workload) =>
  mount(AiApplicationWorkloadExpanded, {
    props: { workload },
  });

const getDetailValue = (wrapper: ReturnType<typeof createWrapper>, label: string): string => {
  const details = wrapper.findAll(".ai-application-workload-details > div");
  const detail = details.find((d) => d.text().startsWith(label));
  return detail ? detail.text().replace(`${label}: `, "") : "";
};

describe("AiApplicationWorkloadExpanded", () => {
  describe("allocation metrics", () => {
    it("should display all allocated resource metrics", () => {
      const workload = createWorkload({
        allocatedResources: {
          gpu: 0.12,
          gpuMemory: "10240M",
          cpu: 0.1,
          cpuMemory: "20Gi",
        },
      });

      const wrapper = createWrapper(workload);

      expect(getDetailValue(wrapper, "GPU memory allocation")).toBe("10.24 GB");
      expect(getDetailValue(wrapper, "GPU compute allocation")).toBe("0.12");
      expect(getDetailValue(wrapper, "CPU memory allocation")).toBe("21.47 GB");
      expect(getDetailValue(wrapper, "CPU compute allocation")).toBe("0.10");
    });

    it("should show dash when allocatedResources is null", () => {
      const workload = createWorkload({ allocatedResources: null });
      const wrapper = createWrapper(workload);

      expect(getDetailValue(wrapper, "GPU memory allocation")).toBe("-");
      expect(getDetailValue(wrapper, "GPU compute allocation")).toBe("-");
      expect(getDetailValue(wrapper, "CPU memory allocation")).toBe("-");
      expect(getDetailValue(wrapper, "CPU compute allocation")).toBe("-");
    });

    it("should show dash for empty allocatedResources", () => {
      const workload = createWorkload({ allocatedResources: {} });
      const wrapper = createWrapper(workload);

      expect(getDetailValue(wrapper, "GPU memory allocation")).toBe("-");
      expect(getDetailValue(wrapper, "GPU compute allocation")).toBe("-");
      expect(getDetailValue(wrapper, "CPU memory allocation")).toBe("-");
      expect(getDetailValue(wrapper, "CPU compute allocation")).toBe("-");
    });
  });

  describe("go-to-workload link", () => {
    it("should emit go-to-workload event when link is clicked", async () => {
      const workload = createWorkload();
      const wrapper = createWrapper(workload);

      await wrapper.find(".go-to-workload-link").trigger("click");

      expect(wrapper.emitted("go-to-workload")).toHaveLength(1);
    });
  });
});
