/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, beforeEach, afterEach, it, expect, vi, type Mock } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

import WorkloadYamlTab from "./workload-yaml-tab.vue";
import { EYamlTabView } from "./workload-yaml-tab.model";
import { clusterApi } from "@/services/infra/client-apis/cluster-api/cluster-api";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { createDownloadLink } from "@/utils/common.util";
import type * as CommonUtil from "@/utils/common.util";
import type { Pod, WorkloadElement, Workload } from "@/swagger-models/workloads-client";

vi.mock("@/services/infra/client-apis/cluster-api/cluster-api", () => ({
  clusterApi: {
    k8sObjectApi: {
      getWorkloadObject: vi.fn(),
      getPodObject: vi.fn(),
      getElementObject: vi.fn(),
    },
  },
}));

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getWorkloadPods: vi.fn(),
    getWorkloadElements: vi.fn(),
  },
}));

vi.mock("@/utils/common.util", async (importOriginal) => ({
  ...(await importOriginal<typeof CommonUtil>()),
  createDownloadLink: vi.fn(),
}));

// Monaco can't render under jsdom — stub the editor to surface its `code` prop
// in the DOM so existing assertions on the rendered YAML keep working.
vi.mock("@/components/common/runai-editor", () => ({
  RunaiEditor: {
    name: "runai-editor",
    props: ["code", "readOnly", "height", "language"],
    template: '<pre class="yaml-content">{{ code }}</pre>',
  },
}));

const mockWorkload = {
  id: "workload-1",
  name: "my-workload",
} as Workload;

const mockPod = (id: string, name: string): Pod => ({ id, name, workloadId: "workload-1" }) as Pod;

const mockElement = (id: string, name: string, refName?: string, kind = "ReplicaSet"): WorkloadElement =>
  ({
    id,
    name,
    kubernetesRef: refName ? { name: refName, gvk: { group: "apps", version: "v1", kind } } : undefined,
  }) as WorkloadElement;

const mountTab = (workload: Workload | null = mockWorkload): VueWrapper =>
  mount(WorkloadYamlTab, {
    global: {
      plugins: [Quasar],
    },
    props: { workload },
  });

describe("workload-yaml-tab.vue", () => {
  beforeEach(() => {
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockReset();
    (clusterApi.k8sObjectApi.getPodObject as Mock).mockReset();
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockReset();
    (workloadService.getWorkloadPods as Mock).mockReset().mockResolvedValue([]);
    (workloadService.getWorkloadElements as Mock).mockReset().mockResolvedValue([]);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the workload object as YAML on successful fetch", async () => {
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: {
        object: {
          apiVersion: "run.ai/v2alpha1",
          kind: "TrainingWorkload",
          metadata: { name: "my-workload", namespace: "runai-test" },
          spec: { image: "alpine" },
        },
        retrievedAt: "2026-06-02T00:00:00Z",
      },
    });

    const wrapper = mountTab();
    await flushPromises();

    const pre = wrapper.find("pre.yaml-content");
    expect(pre.exists()).toBe(true);
    expect(pre.text()).toContain("apiVersion: run.ai/v2alpha1");
    expect(pre.text()).toContain("kind: TrainingWorkload");
    expect(pre.text()).toContain("name: my-workload");
    expect(clusterApi.k8sObjectApi.getWorkloadObject).toHaveBeenCalledWith("workload-1");
  });

  it("shows the not-available message when the upstream returns 404", async () => {
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    const err: { response: { status: number } } & Error = Object.assign(new Error("not found"), {
      response: { status: 404 },
    });
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockRejectedValue(err);

    const wrapper = mountTab();
    await flushPromises();

    expect(wrapper.text()).toContain("There is no YAML yet.");
    warnSpy.mockRestore();
  });

  it("shows the generic error message on non-404 failures", async () => {
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    const err: { response: { status: number } } & Error = Object.assign(new Error("boom"), {
      response: { status: 500 },
    });
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockRejectedValue(err);

    const wrapper = mountTab();
    await flushPromises();

    expect(wrapper.text()).toContain("YAML isn't available right now.");
    warnSpy.mockRestore();
  });

  it("disables the download action while YAML is unavailable", async () => {
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({ data: { object: null } });

    const wrapper = mountTab();
    await flushPromises();

    const downloadBtn = wrapper.find('[aid="workload-yaml-tab-download"]');
    expect(downloadBtn.exists()).toBe(true);
    expect(downloadBtn.attributes("disabled")).toBeDefined();
  });

  it("ignores a stale response from a previous workload after switching", async () => {
    let resolveSlow!: (value: unknown) => void;
    const slowResponse = new Promise((resolve) => {
      resolveSlow = resolve;
    });
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock)
      .mockReturnValueOnce(slowResponse) // workload-1: never resolves until we say so
      .mockResolvedValueOnce({ data: { object: { kind: "SecondWorkload" } } }); // workload-2: lands first

    const wrapper = mount(WorkloadYamlTab, { global: { plugins: [Quasar] }, props: { workload: mockWorkload } });
    await wrapper.setProps({ workload: { id: "workload-2", name: "second" } as Workload });
    await flushPromises();

    // The stale workload-1 response resolves last — it must NOT overwrite workload-2.
    resolveSlow({ data: { object: { kind: "FirstWorkload" } } });
    await flushPromises();

    expect(wrapper.text()).toContain("kind: SecondWorkload");
    expect(wrapper.text()).not.toContain("FirstWorkload");
  });

  it("downloads the YAML using the selected target's filename", async () => {
    (createDownloadLink as Mock).mockClear();
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" } } },
    });

    const wrapper = mountTab();
    await flushPromises();
    await wrapper.find('[aid="workload-yaml-tab-download"]').trigger("click");

    expect(createDownloadLink).toHaveBeenCalledWith(expect.any(Blob), "my-workload.yaml");
  });

  it("does not fetch when workload is missing", async () => {
    mountTab(null);
    await flushPromises();
    expect(clusterApi.k8sObjectApi.getWorkloadObject).not.toHaveBeenCalled();
    expect(workloadService.getWorkloadPods).not.toHaveBeenCalled();
    expect(workloadService.getWorkloadElements).not.toHaveBeenCalled();
  });

  it("renders the pod object when a pod is selected via the dropdown", async () => {
    (workloadService.getWorkloadPods as Mock).mockResolvedValue([mockPod("pod-1", "my-workload-0-0")]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" }, spec: {} } },
    });
    (clusterApi.k8sObjectApi.getPodObject as Mock).mockResolvedValue({
      data: { object: { kind: "Pod", apiVersion: "v1", metadata: { name: "my-workload-0-0" }, spec: {} } },
    });

    const wrapper = mountTab();
    await flushPromises();

    (wrapper.vm as unknown as { selectedTargetKey: string }).selectedTargetKey = "pod:pod-1";
    await flushPromises();

    expect(clusterApi.k8sObjectApi.getPodObject).toHaveBeenCalledWith("workload-1", "pod-1");
    expect(wrapper.find("pre.yaml-content").text()).toContain("kind: Pod");
  });

  it("renders the element object when an element is selected via the dropdown", async () => {
    (workloadService.getWorkloadElements as Mock).mockResolvedValue([
      mockElement("el-1", "replicaset", "my-workload-5c68d7"),
    ]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" }, spec: {} } },
    });
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockResolvedValue({
      data: {
        object: { kind: "ReplicaSet", apiVersion: "apps/v1", metadata: { name: "my-workload-5c68d7" }, spec: {} },
      },
    });

    const wrapper = mountTab();
    await flushPromises();

    (wrapper.vm as unknown as { selectedTargetKey: string }).selectedTargetKey = "element:el-1";
    await flushPromises();

    expect(clusterApi.k8sObjectApi.getElementObject).toHaveBeenCalledWith("workload-1", "el-1");
    expect(wrapper.find("pre.yaml-content").text()).toContain("kind: ReplicaSet");
  });

  it("collapsed view (default): dropdown lists Workload + Pods only", async () => {
    (workloadService.getWorkloadPods as Mock).mockResolvedValue([mockPod("pod-1", "my-workload-0-0")]);
    (workloadService.getWorkloadElements as Mock).mockResolvedValue([
      mockElement("el-1", "replicaset", "my-workload-5c68d7"),
    ]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" }, spec: {} } },
    });

    const wrapper = mountTab();
    await flushPromises();

    const options = (wrapper.vm as unknown as { targetOptions: Array<{ value: string }> }).targetOptions;
    expect(options.map((o) => o.value)).toEqual(["workload", "pod:pod-1"]);
  });

  it("expanded view at workload root: dropdown lists Workload + all concrete elements + all pods", async () => {
    (workloadService.getWorkloadPods as Mock).mockResolvedValue([mockPod("pod-1", "my-workload-0-0")]);
    (workloadService.getWorkloadElements as Mock).mockResolvedValue([
      mockElement("el-1", "replicaset", "my-workload-5c68d7"),
    ]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" }, spec: {} } },
    });

    const wrapper = mount(WorkloadYamlTab, {
      global: { plugins: [Quasar] },
      props: { workload: mockWorkload, view: EYamlTabView.Expanded },
    });
    await flushPromises();

    const options = (wrapper.vm as unknown as { targetOptions: Array<{ value: string; label: string }> }).targetOptions;
    expect(options.map((o) => o.value)).toEqual(["workload", "element:el-1", "pod:pod-1"]);
    // Label uses the GVK kind as prefix (e.g. "ReplicaSet: <name>"), not the
    // generic "Element" word.
    expect(options.find((o) => o.value === "element:el-1")?.label).toBe("ReplicaSet: my-workload-5c68d7");
  });

  it("expanded view drilled into an element: dropdown narrows to that element + direct children", async () => {
    (workloadService.getWorkloadPods as Mock).mockResolvedValue([
      { ...mockPod("pod-a", "worker-0"), elements: [{ elementId: "el-1", hierarchyLevel: 1 }] } as Pod,
      { ...mockPod("pod-b", "worker-1"), elements: [{ elementId: "el-1", hierarchyLevel: 1 }] } as Pod,
      { ...mockPod("pod-c", "leader-0"), elements: [{ elementId: "el-2", hierarchyLevel: 1 }] } as Pod,
    ]);
    (workloadService.getWorkloadElements as Mock).mockResolvedValue([
      mockElement("el-1", "worker", "yaml-lws-1-0"),
      mockElement("el-2", "leader", "yaml-lws-1-leader-0"),
    ]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "LeaderWorkerSet", metadata: { name: "yaml-lws-1" }, spec: {} } },
    });
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockResolvedValue({
      data: { object: { kind: "StatefulSet", apiVersion: "apps/v1", metadata: { name: "yaml-lws-1-0" }, spec: {} } },
    });

    const wrapper = mount(WorkloadYamlTab, {
      global: { plugins: [Quasar] },
      // Parent (the structure map) tells the YAML tab which element is in scope.
      props: { workload: mockWorkload, view: EYamlTabView.Expanded, scopeElementId: "el-1" },
    });
    await flushPromises();

    const options = (wrapper.vm as unknown as { targetOptions: Array<{ value: string }> }).targetOptions;
    // The dropdown should be scoped to el-1: el-1 itself + its two child pods
    // (pod-c belongs to el-2, must not appear).
    expect(options.map((o) => o.value)).toEqual(["element:el-1", "pod:pod-a", "pod:pod-b"]);
  });

  // Shared LWS fixture used across the scoping tests:
  //   group-0 (abstract, root)
  //     ├── worker-0 (StatefulSet, concrete)   → pod-w1, pod-w2
  //     └── leader-0 (abstract)                → pod-leader
  const lwsElements = (): WorkloadElement[] => [
    { id: "group-0", name: "group-0", kubernetesRef: undefined } as unknown as WorkloadElement,
    {
      ...mockElement("worker-0", "worker", "yaml-lws-1-0", "StatefulSet"),
      parentId: "group-0",
    } as unknown as WorkloadElement,
    {
      id: "leader-0",
      name: "leader-0",
      parentId: "group-0",
      kubernetesRef: undefined,
    } as unknown as WorkloadElement,
  ];
  const lwsPods = (): Pod[] => [
    {
      ...mockPod("pod-w1", "yaml-lws-1-0-1"),
      elements: [
        { elementId: "group-0", hierarchyLevel: 1 },
        { elementId: "worker-0", hierarchyLevel: 2 },
      ],
    } as Pod,
    {
      ...mockPod("pod-w2", "yaml-lws-1-0-2"),
      elements: [
        { elementId: "group-0", hierarchyLevel: 1 },
        { elementId: "worker-0", hierarchyLevel: 2 },
      ],
    } as Pod,
    {
      ...mockPod("pod-leader", "yaml-lws-1-leader"),
      elements: [
        { elementId: "group-0", hierarchyLevel: 1 },
        { elementId: "leader-0", hierarchyLevel: 2 },
      ],
    } as Pod,
  ];

  it("expanded view scoped to an abstract element: surfaces ALL descendant pods (incl. those under concrete sub-elements)", async () => {
    (workloadService.getWorkloadElements as Mock).mockResolvedValue(lwsElements());
    (workloadService.getWorkloadPods as Mock).mockResolvedValue(lwsPods());
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "LeaderWorkerSet", metadata: { name: "yaml-lws-1" }, spec: {} } },
    });
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockResolvedValue({
      data: { object: { kind: "StatefulSet", apiVersion: "apps/v1", metadata: { name: "yaml-lws-1-0" }, spec: {} } },
    });

    const wrapper = mount(WorkloadYamlTab, {
      global: { plugins: [Quasar] },
      props: { workload: mockWorkload, view: EYamlTabView.Expanded, scopeElementId: "group-0" },
    });
    await flushPromises();

    const options = (wrapper.vm as unknown as { targetOptions: Array<{ value: string }> }).targetOptions;
    // group-0 is abstract → skipped in options. Worker-0 (concrete descendant)
    // appears. ALL 3 pods whose chain reaches group-0 surface here — including
    // the two worker pods that live under the concrete worker-0 StatefulSet
    // (regression: a previous filter dropped them when a concrete intermediate
    // existed in the chain).
    expect(options.map((o) => o.value)).toEqual(["element:worker-0", "pod:pod-w1", "pod:pod-w2", "pod:pod-leader"]);
    // Abstract scope auto-selects the first concrete entry — chip shouldn't be
    // blank or fall back to "Workload".
    expect((wrapper.vm as unknown as { selectedTargetKey: string }).selectedTargetKey).toBe("element:worker-0");
  });

  it("expanded view scoped to a concrete element: narrows to that element + only its descendant pods", async () => {
    (workloadService.getWorkloadElements as Mock).mockResolvedValue(lwsElements());
    (workloadService.getWorkloadPods as Mock).mockResolvedValue(lwsPods());
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "LeaderWorkerSet", metadata: { name: "yaml-lws-1" }, spec: {} } },
    });
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockResolvedValue({
      data: { object: { kind: "StatefulSet", apiVersion: "apps/v1", metadata: { name: "yaml-lws-1-0" }, spec: {} } },
    });

    const wrapper = mount(WorkloadYamlTab, {
      global: { plugins: [Quasar] },
      props: { workload: mockWorkload, view: EYamlTabView.Expanded, scopeElementId: "worker-0" },
    });
    await flushPromises();

    const options = (wrapper.vm as unknown as { targetOptions: Array<{ value: string }> }).targetOptions;
    // Drilling into worker-0 narrows: leader pod (chain doesn't reach worker-0)
    // is gone, only the two worker pods + the StatefulSet itself remain.
    expect(options.map((o) => o.value)).toEqual(["element:worker-0", "pod:pod-w1", "pod:pod-w2"]);
  });

  it("expanded view: changing scopeElementId resets the active YAML target to that element", async () => {
    (workloadService.getWorkloadElements as Mock).mockResolvedValue([
      mockElement("el-1", "replicaset", "my-workload-5c68d7"),
    ]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" }, spec: {} } },
    });
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockResolvedValue({
      data: {
        object: { kind: "ReplicaSet", apiVersion: "apps/v1", metadata: { name: "my-workload-5c68d7" }, spec: {} },
      },
    });

    const wrapper = mount(WorkloadYamlTab, {
      global: { plugins: [Quasar] },
      props: { workload: mockWorkload, view: EYamlTabView.Expanded, scopeElementId: null },
    });
    await flushPromises();
    expect(wrapper.find("pre.yaml-content").text()).toContain("kind: TrainingWorkload");

    // Map navigates to el-1.
    await wrapper.setProps({ scopeElementId: "el-1" });
    await flushPromises();

    expect(clusterApi.k8sObjectApi.getElementObject).toHaveBeenCalledWith("workload-1", "el-1");
    expect(wrapper.find("pre.yaml-content").text()).toContain("kind: ReplicaSet");
  });

  it("returns a clear message on a 404 element fetch", async () => {
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    (workloadService.getWorkloadElements as Mock).mockResolvedValue([
      mockElement("el-1", "replicaset", "my-workload-5c68d7"),
    ]);
    (clusterApi.k8sObjectApi.getWorkloadObject as Mock).mockResolvedValue({
      data: { object: { kind: "TrainingWorkload", metadata: { name: "my-workload" }, spec: {} } },
    });
    const err: { response: { status: number } } & Error = Object.assign(new Error("not found"), {
      response: { status: 404 },
    });
    (clusterApi.k8sObjectApi.getElementObject as Mock).mockRejectedValue(err);

    const wrapper = mountTab();
    await flushPromises();

    (wrapper.vm as unknown as { selectedTargetKey: string }).selectedTargetKey = "element:el-1";
    await flushPromises();

    expect(wrapper.text()).toContain("There is no YAML yet.");
    warnSpy.mockRestore();
  });
});
