/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import YamlWorkloadSection from "./yaml-workload-section.vue";

// services
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { fileUtil } from "@/utils/file.util";

// Mock services
vi.mock("@/services/infra/client-apis/workloads-service-api/workloads-service-api");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/utils/file.util");

const pinia = createTestingPinia({ createSpy: vi.fn });

const mockWorkloadTypes = [
  { id: "1", name: "TFJob" },
  { id: "2", name: "PyTorchJob" },
  { id: "3", name: "JAXJob" },
];

describe("yaml-workload-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof YamlWorkloadSection>>;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();

    vi.mocked(workloadServiceApi.workloadPropertiesApi.listWorkloadTypes).mockResolvedValue({
      data: { types: mockWorkloadTypes },
    } as never);

    vi.mocked(workloadService.createWorkloadFromYaml).mockResolvedValue({
      metadata: { projectId: "1" },
    } as never);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function createWrapper(props = {}) {
    return mount(YamlWorkloadSection, {
      props: {
        clusterId: "cluster-1",
        selectedFormValues: { projectId: null, priority: null, preemptibility: null },
        ...props,
      },
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $q: {
            notify: vi.fn(),
          },
        },
        stubs: {
          RunaiExpansionItem: {
            template: "<div class='expansion-item'><slot name='subheader'></slot><slot></slot></div>",
            props: ["label", "bodyNoPadding", "sectionInvalid"],
          },
          RunaiEditor: {
            template: "<div class='runai-editor'></div>",
            props: ["code", "height", "showExpandButton", "isExpanded", "showPlaceholder", "placeholder"],
            emits: ["changed", "validation-status", "toggle-expand", "pasted"],
            methods: {
              setContent: vi.fn(),
            },
          },
          RunaiBtn: {
            template: "<button @click='$emit(\"click\")'><slot></slot></button>",
            props: ["icon", "label", "color", "textColor"],
          },
          QMenu: { template: "<div><slot></slot></div>" },
          QList: { template: "<div><slot></slot></div>" },
          QItem: { template: "<div><slot></slot></div>" },
          QItemSection: { template: "<div><slot></slot></div>" },
          QSpinner: { template: "<div class='spinner'></div>" },
        },
      },
    });
  }

  describe("Component Initialization", () => {
    it("should render the component", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.exists()).toBe(true);
    });

    it("should load workload types on created and sort them alphabetically via computed", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(workloadServiceApi.workloadPropertiesApi.listWorkloadTypes).toHaveBeenCalledWith(true, undefined, false);
      // Raw data is stored unsorted
      expect(wrapper.vm.workloadTypes).toEqual([
        { id: "1", name: "TFJob" },
        { id: "2", name: "PyTorchJob" },
        { id: "3", name: "JAXJob" },
      ]);
      // Computed property returns sorted list
      expect(wrapper.vm.sortedWorkloadTypes).toEqual([
        { id: "3", name: "JAXJob" },
        { id: "2", name: "PyTorchJob" },
        { id: "1", name: "TFJob" },
      ]);
    });

    it("should initialize with empty YAML content", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.yamlContent).toBe("");
      expect(wrapper.vm.isYamlValid).toBe(false);
    });

    it("should initialize with provided yaml", async () => {
      const yaml = "apiVersion: v1\nkind: Pod";
      wrapper = createWrapper({ yaml });
      await flushPromises();

      expect(wrapper.vm.yamlContent).toBe(yaml);
    });
  });

  describe("Subheader Text", () => {
    it("should display 'Name: None' when YAML is empty", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.subheaderText).toBe("Name: None");
    });

    it("should display workload name when YAML has metadata.name", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: my-workload";

      expect(wrapper.vm.subheaderText).toBe("Name: my-workload");
    });

    it("should display 'Name: None' when YAML has no name", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod";

      expect(wrapper.vm.subheaderText).toBe("Name: None");
    });
  });

  describe("YAML Changed Event", () => {
    it("should emit yaml-input on content change", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const yamlContent = "apiVersion: v1";
      wrapper.vm.onYamlChanged(yamlContent);

      expect(wrapper.emitted("yaml-input")).toBeTruthy();
      expect(wrapper.emitted("yaml-input")?.[0]).toEqual([yamlContent]);
    });

    it("should update internal yamlContent", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const yamlContent = "apiVersion: v1";
      wrapper.vm.onYamlChanged(yamlContent);

      expect(wrapper.vm.yamlContent).toBe(yamlContent);
    });
  });

  describe("YAML Pasted Event", () => {
    it("should emit yaml-loaded on paste", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const yamlContent = "apiVersion: v1\nkind: Pod";
      wrapper.vm.onYamlPasted(yamlContent);

      expect(wrapper.emitted("yaml-loaded")).toBeTruthy();
      expect(wrapper.emitted("yaml-loaded")?.[0]).toEqual([yamlContent]);
    });

    it("should reset previousDryRunFields on paste", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.previousDryRunFields = {
        name: "old-name",
        apiVersion: "v1",
        kind: "Pod",
        namespace: "default",
      };

      wrapper.vm.onYamlPasted("apiVersion: v1\nkind: Pod");

      expect(wrapper.vm.previousDryRunFields).toEqual({
        name: null,
        apiVersion: null,
        kind: null,
        namespace: null,
      });
    });
  });

  describe("File Upload", () => {
    it("should read file content and update editor", async () => {
      const fileContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      vi.mocked(fileUtil.readAsText).mockResolvedValue(fileContent);

      wrapper = createWrapper();
      await flushPromises();

      const mockFile = new File([fileContent], "test.yaml", { type: "text/yaml" });
      const mockEvent = {
        target: {
          files: [mockFile],
          value: "",
        },
      } as unknown as Event;

      await wrapper.vm.onFileSelected(mockEvent);
      await flushPromises();

      expect(fileUtil.readAsText).toHaveBeenCalledWith(mockFile);
      expect(wrapper.vm.yamlContent).toBe(fileContent);
      expect(wrapper.emitted("yaml-input")).toBeTruthy();
      expect(wrapper.emitted("yaml-loaded")).toBeTruthy();
    });

    it("should reject invalid file extensions", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      const mockFile = new File(["content"], "test.txt", { type: "text/plain" });
      const mockEvent = {
        target: {
          files: [mockFile],
          value: "",
        },
      } as unknown as Event;

      await wrapper.vm.onFileSelected(mockEvent);

      expect(fileUtil.readAsText).not.toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith("Invalid file type. Only .yaml and .yml files are allowed.");

      consoleSpy.mockRestore();
    });

    it("should accept .yml extension", async () => {
      const fileContent = "apiVersion: v1";
      vi.mocked(fileUtil.readAsText).mockResolvedValue(fileContent);

      wrapper = createWrapper();
      await flushPromises();

      const mockFile = new File([fileContent], "test.yml", { type: "text/yaml" });
      const mockEvent = {
        target: {
          files: [mockFile],
          value: "",
        },
      } as unknown as Event;

      await wrapper.vm.onFileSelected(mockEvent);
      await flushPromises();

      expect(fileUtil.readAsText).toHaveBeenCalled();
    });
  });

  describe("Validation Status", () => {
    it("should emit validation-changed when status changes", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.onValidationStatusChanged({ isValid: true, error: "" });

      expect(wrapper.emitted("validation-changed")).toBeTruthy();
      expect(wrapper.emitted("validation-changed")?.[0]).toEqual([true]);
    });

    it("should update isYamlValid on validation status change", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.onValidationStatusChanged({ isValid: true, error: "" });
      expect(wrapper.vm.isYamlValid).toBe(true);

      wrapper.vm.onValidationStatusChanged({ isValid: false, error: "Invalid YAML" });
      expect(wrapper.vm.isYamlValid).toBe(false);
    });
  });

  describe("Dry-Run Validation", () => {
    it("should trigger dry-run when apiVersion changes", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      wrapper.vm.previousDryRunFields = {
        name: "test",
        apiVersion: "v1",
        kind: "Pod",
        namespace: null,
      };

      const newYaml = "apiVersion: v2\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.checkDryRunFieldsChanged(newYaml);

      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should trigger dry-run when kind changes", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      wrapper.vm.previousDryRunFields = {
        name: "test",
        apiVersion: "v1",
        kind: "Pod",
        namespace: null,
      };

      const newYaml = "apiVersion: v1\nkind: Deployment\nmetadata:\n  name: test";
      wrapper.vm.checkDryRunFieldsChanged(newYaml);

      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should trigger dry-run when namespace changes", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      wrapper.vm.previousDryRunFields = {
        name: "test",
        apiVersion: "v1",
        kind: "Pod",
        namespace: "default",
      };

      const newYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test\n  namespace: production";
      wrapper.vm.checkDryRunFieldsChanged(newYaml);

      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should trigger dry-run when name changes", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      wrapper.vm.previousDryRunFields = {
        name: "old-name",
        apiVersion: "v1",
        kind: "Pod",
        namespace: null,
      };

      const newYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: new-name";
      wrapper.vm.checkDryRunFieldsChanged(newYaml);

      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should NOT trigger dry-run when fields are unchanged", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      wrapper.vm.previousDryRunFields = {
        name: "test",
        apiVersion: "v1",
        kind: "Pod",
        namespace: null,
      };

      const sameYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.checkDryRunFieldsChanged(sameYaml);

      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).not.toHaveBeenCalled();
    });

    it("should emit dryrun-success on successful validation", async () => {
      const mockWorkload = {
        metadata: { projectId: "1", priority: "high" },
      };
      vi.mocked(workloadService.createWorkloadFromYaml).mockResolvedValue(mockWorkload as never);

      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      const validYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.yamlContent = validYaml;

      await wrapper.vm.runDryRunValidation(validYaml);

      expect(wrapper.emitted("dryrun-success")).toBeTruthy();
      expect(wrapper.emitted("dryrun-success")?.[0]).toEqual([mockWorkload]);
    });

    it("should show notification and set error state on validation failure", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(workloadService.createWorkloadFromYaml).mockRejectedValue(new Error("Invalid configuration"));

      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      const validYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.yamlContent = validYaml;

      await wrapper.vm.runDryRunValidation(validYaml);

      expect(wrapper.vm.isDryRunValid).toBe(false);
      expect(wrapper.vm.dryRunError).toBe("Invalid configuration");
      // Error is handled internally, no dryrun-success emitted
      expect(wrapper.emitted("dryrun-success")).toBeFalsy();
      errorSpy.mockRestore();
    });

    it("should NOT run dry-run if clusterId is missing", async () => {
      wrapper = createWrapper({ clusterId: "" });
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      await wrapper.vm.runDryRunValidation("apiVersion: v1\nkind: Pod\nmetadata:\n  name: test");

      expect(workloadService.createWorkloadFromYaml).not.toHaveBeenCalled();
    });

    it("should NOT run dry-run if YAML is invalid", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = false;
      await wrapper.vm.runDryRunValidation("invalid: yaml: content:");

      expect(workloadService.createWorkloadFromYaml).not.toHaveBeenCalled();
    });

    it("should use projectId instead of clusterId when valid projectId exists", async () => {
      wrapper = createWrapper({
        selectedFormValues: { projectId: 5, priority: "high", preemptibility: "Preemptible" },
      });
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      const validYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.yamlContent = validYaml;

      await wrapper.vm.runDryRunValidation(validYaml);

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalledWith(
        expect.objectContaining({
          metadata: expect.objectContaining({
            projectId: "5",
          }),
        }),
        true,
        true,
      );
    });

    it("should use clusterId when projectId is not valid", async () => {
      wrapper = createWrapper({
        clusterId: "cluster-1",
        selectedFormValues: { projectId: -1, priority: null, preemptibility: null },
      });
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      const validYaml = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.yamlContent = validYaml;

      await wrapper.vm.runDryRunValidation(validYaml);

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalledWith(
        expect.objectContaining({
          metadata: expect.objectContaining({
            clusterId: "cluster-1",
          }),
        }),
        true,
        true,
      );
    });
  });

  describe("Editor Expand/Collapse", () => {
    it("should toggle isExpanded state", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.isExpanded).toBe(false);

      wrapper.vm.toggleExpand();
      expect(wrapper.vm.isExpanded).toBe(true);

      wrapper.vm.toggleExpand();
      expect(wrapper.vm.isExpanded).toBe(false);
    });

    it("should return correct editor height based on expanded state", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.editorHeight).toBe("300px");

      wrapper.vm.isExpanded = true;
      expect(wrapper.vm.editorHeight).toBe("700px");
    });
  });

  describe("Workload Types Loading", () => {
    it("should show loading spinner while fetching types", async () => {
      wrapper = createWrapper();

      expect(wrapper.vm.loadingWorkloadTypes).toBe(true);

      await flushPromises();

      expect(wrapper.vm.loadingWorkloadTypes).toBe(false);
    });

    it("should handle workload types fetch error", async () => {
      vi.mocked(workloadServiceApi.workloadPropertiesApi.listWorkloadTypes).mockRejectedValue(
        new Error("Network error"),
      );

      const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.workloadTypes).toEqual([]);
      expect(consoleSpy).toHaveBeenCalledWith("Failed to load workload types:", expect.anything());

      consoleSpy.mockRestore();
    });
  });

  describe("Clear Dry-Run Validation", () => {
    it("should reset dry-run state", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.dryRunError = "Some error";
      wrapper.vm.isDryRunValid = false;

      wrapper.vm.clearDryRunValidation();

      expect(wrapper.vm.dryRunError).toBe("");
      expect(wrapper.vm.isDryRunValid).toBe(true);
    });
  });

  describe("Watch: yaml", () => {
    it("should update content when yaml prop changes externally", async () => {
      wrapper = createWrapper({ yaml: "" });
      await flushPromises();

      await wrapper.setProps({ yaml: "apiVersion: v1\nkind: Pod" });
      await flushPromises();

      expect(wrapper.vm.yamlContent).toBe("apiVersion: v1\nkind: Pod");
    });

    it("should NOT update if change originated internally", async () => {
      wrapper = createWrapper({ yaml: "" });
      await flushPromises();

      wrapper.vm.isInternalUpdate = true;
      wrapper.vm.yamlContent = "internal content";

      await wrapper.setProps({ yaml: "external content" });
      await flushPromises();

      // Should not have been overwritten
      expect(wrapper.vm.yamlContent).toBe("internal content");
    });
  });

  describe("Watch: clusterId", () => {
    it("should re-run dry-run validation when clusterId changes and YAML is valid", async () => {
      wrapper = createWrapper({ clusterId: "cluster-1" });
      await flushPromises();

      wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.isYamlValid = true;

      vi.clearAllMocks();

      await wrapper.setProps({ clusterId: "cluster-2" });
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });
  });

  describe("Watch: selectedFormValues", () => {
    it("should re-run dry-run validation when project changes and YAML is valid", async () => {
      wrapper = createWrapper({
        clusterId: "cluster-1",
        selectedFormValues: { projectId: 1, priority: null, preemptibility: null },
      });
      await flushPromises();

      wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.isYamlValid = true;

      vi.clearAllMocks();

      await wrapper.setProps({
        selectedFormValues: { projectId: 2, priority: null, preemptibility: null },
      });

      // Wait for debounce
      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should re-run dry-run validation when priority changes and YAML is valid", async () => {
      wrapper = createWrapper({
        clusterId: "cluster-1",
        selectedFormValues: { projectId: 1, priority: "low", preemptibility: null },
      });
      await flushPromises();

      wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.isYamlValid = true;

      vi.clearAllMocks();

      await wrapper.setProps({
        selectedFormValues: { projectId: 1, priority: "high", preemptibility: null },
      });

      // Wait for debounce
      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should re-run dry-run validation when preemptibility changes and YAML is valid", async () => {
      wrapper = createWrapper({
        clusterId: "cluster-1",
        selectedFormValues: { projectId: 1, priority: null, preemptibility: "Preemptible" },
      });
      await flushPromises();

      wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test";
      wrapper.vm.isYamlValid = true;

      vi.clearAllMocks();

      await wrapper.setProps({
        selectedFormValues: { projectId: 1, priority: null, preemptibility: "NonPreemptible" },
      });

      // Wait for debounce
      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).toHaveBeenCalled();
    });

    it("should NOT re-run validation when YAML is empty or invalid", async () => {
      wrapper = createWrapper({
        clusterId: "cluster-1",
        selectedFormValues: { projectId: 1, priority: null, preemptibility: null },
      });
      await flushPromises();

      // No YAML content
      wrapper.vm.yamlContent = "";
      wrapper.vm.isYamlValid = false;

      vi.clearAllMocks();

      await wrapper.setProps({
        selectedFormValues: { projectId: 2, priority: null, preemptibility: null },
      });

      // Wait for debounce
      vi.advanceTimersByTime(500);
      await flushPromises();

      expect(workloadService.createWorkloadFromYaml).not.toHaveBeenCalled();
    });
  });
});
