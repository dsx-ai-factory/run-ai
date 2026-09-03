/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test, vi, afterEach } from "vitest";
import DockerRegistrySelector from "./docker-registry-selector.vue";
import type { IDockerRegistrySelectorModel } from "./docker-registry-selector.model";
import { EDockerRegistrySelectorEvents } from "./docker-registry-selector.model";
import type { IScopeModel } from "@/models/global.model";
import { AssetSyncStatusEnum, type DockerRegistry } from "@/swagger-models/assets-service-client";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { createTestingPinia } from "@pinia/testing";

vi.mock("@/services/control-plane/credential.service/credential.service", () => ({
  credentialService: {
    listDockerRegistries: vi.fn(),
  },
}));

describe("docker-registry-selector.vue", () => {
  let dockerRegistrySelectorCmp: VueWrapper<InstanceType<typeof DockerRegistrySelector>>;

  const mockScopeModel: IScopeModel = {
    clusterId: "cluster-123",
    projectId: 1,
    departmentId: "department-123",
  } as IScopeModel;

  const mockModelValue: IDockerRegistrySelectorModel = {
    dockerRegistryName: "",
  };

  const mockDockerRegistries: DockerRegistry[] = [
    {
      meta: {
        name: "ngc-secret-1",
        uid: "uid-1",
      },
      spec: {},
      status: {
        status: AssetSyncStatusEnum.NoIssuesFound,
      },
    } as unknown as DockerRegistry,
    {
      meta: {
        name: "ngc-secret-2",
        uid: "uid-2",
      },
      spec: {},
      status: {
        status: AssetSyncStatusEnum.NoIssuesFound,
      },
    } as unknown as DockerRegistry,
    {
      meta: {
        name: "docker-hub-secret",
        uid: "uid-3",
      },
      spec: {},
      status: {
        status: AssetSyncStatusEnum.NoIssuesFound,
      },
    } as unknown as DockerRegistry,
    {
      meta: {
        name: "creating-secret",
        uid: "uid-4",
      },
      spec: {},
      status: {
        status: AssetSyncStatusEnum.Creating,
      },
    } as unknown as DockerRegistry,
    {
      meta: {
        name: "no-status-secret",
        uid: "uid-5",
      },
      spec: {},
    } as unknown as DockerRegistry,
  ];

  function mountComponent(props = {}) {
    const pinia = createTestingPinia({ createSpy: vi.fn });

    dockerRegistrySelectorCmp = mount(DockerRegistrySelector, {
      global: {
        plugins: [Quasar, pinia],
        stubs: {
          ImagePullCredentialLayout: {
            template:
              '<div class="image-pull-credential-layout"><div class="section-title">{{ sectionTitle }}</div><slot></slot><div v-if="showInfoBanner" class="nim-info-banner">NVIDIA NIM requires an image that is pulled from the NGC catalog</div></div>',
            props: ["sectionTitle", "showInfoBanner"],
          },
          RunaiSection: {
            template: '<div class="runai-section"><div class="title">{{ title }}</div><slot></slot></div>',
            props: ["title"],
          },
          RunaiSelect: {
            template: "<select @change=\"$emit('update:model-value', $event.target.value)\"></select>",
            props: ["modelValue", "options", "label", "noOptionText", "stackLabel", "placeholder"],
          },
        },
      },
      props: {
        modelValue: mockModelValue,
        clusterId: "cluster-123",
        isRequired: true,
        scopeModel: mockScopeModel,
        ...props,
      },
    });
  }

  beforeEach(() => {
    vi.mocked(credentialService.listDockerRegistries).mockResolvedValue(mockDockerRegistries);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  test("component is loading properly", async () => {
    mountComponent();
    await flushPromises();

    expect(dockerRegistrySelectorCmp.exists()).toBe(true);
    expect(dockerRegistrySelectorCmp.html()).toContain("Set the credentials for pulling the image");
  });

  test("should display NGC catalog information banner", async () => {
    mountComponent();
    await flushPromises();

    expect(dockerRegistrySelectorCmp.html()).toContain("NVIDIA NIM requires an image that is pulled from the");
    expect(dockerRegistrySelectorCmp.html()).toContain("NGC catalog");
  });

  describe("Lifecycle hooks", () => {
    test("should load credentials on component creation", async () => {
      mountComponent();
      await flushPromises();

      expect(credentialService.listDockerRegistries).toHaveBeenCalledTimes(1);
      expect(credentialService.listDockerRegistries).toHaveBeenCalledWith({
        projectId: mockScopeModel.projectId,
        clusterUuid: mockScopeModel.clusterId,
        departmentId: mockScopeModel.departmentId,
        privateOnly: true,
        statusInfo: true,
      });
    });
  });

  describe("Methods", () => {
    describe("getFilterBy", () => {
      test("should return correct filter object with all scope properties", async () => {
        mountComponent();
        await flushPromises();

        const filter = dockerRegistrySelectorCmp.vm.getFilterBy();

        expect(filter).toEqual({
          projectId: mockScopeModel.projectId,
          clusterUuid: mockScopeModel.clusterId,
          departmentId: mockScopeModel.departmentId,
          privateOnly: true,
          statusInfo: true,
        });
      });

      test("should return filter with undefined clusterUuid when clusterId is not provided", async () => {
        const scopeWithoutCluster = { ...mockScopeModel, clusterId: undefined };
        mountComponent({ scopeModel: scopeWithoutCluster });
        await flushPromises();

        const filter = dockerRegistrySelectorCmp.vm.getFilterBy();

        expect(filter.clusterUuid).toBeUndefined();
      });
    });

    describe("getDockerRegistries", () => {
      test("should fetch and transform docker registries into select options", async () => {
        mountComponent();
        await flushPromises();

        const filter = dockerRegistrySelectorCmp.vm.getFilterBy();
        const options = await dockerRegistrySelectorCmp.vm.getDockerRegistries(filter);

        expect(options).toHaveLength(5);
        expect(options[0]).toEqual({
          value: "ngc-secret-1",
          label: "ngc-secret-1",
          disable: false,
          disabledTooltip: undefined,
        });
        expect(options[1]).toEqual({
          value: "ngc-secret-2",
          label: "ngc-secret-2",
          disable: false,
          disabledTooltip: undefined,
        });
        expect(options[2]).toEqual({
          value: "docker-hub-secret",
          label: "docker-hub-secret",
          disable: false,
          disabledTooltip: undefined,
        });
      });

      test("should return empty array when no docker registries exist", async () => {
        vi.mocked(credentialService.listDockerRegistries).mockResolvedValue([]);
        mountComponent();
        await flushPromises();

        const filter = dockerRegistrySelectorCmp.vm.getFilterBy();
        const options = await dockerRegistrySelectorCmp.vm.getDockerRegistries(filter);

        expect(options).toHaveLength(0);
      });

      test("should disable registries with Creating status or no status", async () => {
        mountComponent();
        await flushPromises();

        const filter = dockerRegistrySelectorCmp.vm.getFilterBy();
        const options = await dockerRegistrySelectorCmp.vm.getDockerRegistries(filter);

        // All 5 registries should be returned
        expect(options).toHaveLength(5);

        // Find the creating and no-status registries
        const creatingRegistry = options.find((opt) => opt.value === "creating-secret");
        const noStatusRegistry = options.find((opt) => opt.value === "no-status-secret");

        // Verify they are disabled with tooltip
        expect(creatingRegistry?.disable).toBe(true);
        expect(creatingRegistry?.disabledTooltip).toBe(
          "This setup is being created. If it isn't created after a few seconds, contact your administrator.",
        );

        expect(noStatusRegistry?.disable).toBe(true);
        expect(noStatusRegistry?.disabledTooltip).toBe(
          "This setup is being created. If it isn't created after a few seconds, contact your administrator.",
        );

        // Verify valid registries are not disabled
        const validRegistry = options.find((opt) => opt.value === "ngc-secret-1");
        expect(validRegistry?.disable).toBe(false);
        expect(validRegistry?.disabledTooltip).toBeUndefined();
      });
    });

    describe("loadCredentials", () => {
      test("should populate credentials data property with fetched registries", async () => {
        mountComponent();
        await flushPromises();

        expect(dockerRegistrySelectorCmp.vm.credentials).toHaveLength(5);
        expect(dockerRegistrySelectorCmp.vm.credentials[0].label).toBe("ngc-secret-1");
      });

      test("should set loading state correctly during fetch", async () => {
        let resolvePromise: ((value: DockerRegistry[]) => void) | undefined;
        const promise = new Promise<DockerRegistry[]>((resolve) => {
          resolvePromise = resolve;
        });

        vi.mocked(credentialService.listDockerRegistries).mockReturnValue(promise);

        mountComponent();

        // Should be loading initially
        expect(dockerRegistrySelectorCmp.vm.loading).toBe(true);

        if (resolvePromise) {
          resolvePromise(mockDockerRegistries);
        }
        await flushPromises();

        // Should not be loading after completion
        expect(dockerRegistrySelectorCmp.vm.loading).toBe(false);
      });

      test("should handle errors gracefully and show notification", async () => {
        const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
        const mockError = new Error("Failed to load docker registries");
        vi.mocked(credentialService.listDockerRegistries).mockRejectedValue(mockError);

        const notifySpy = vi.fn();
        mountComponent();
        dockerRegistrySelectorCmp.vm.$q.notify = notifySpy;

        await dockerRegistrySelectorCmp.vm.loadCredentials();
        await flushPromises();

        expect(notifySpy).toHaveBeenCalled();
        expect(dockerRegistrySelectorCmp.vm.loading).toBe(false);
        errorSpy.mockRestore();
      });
    });

    describe("emitChanges", () => {
      test("should emit UpdateModelValue event with correct payload", async () => {
        mountComponent();
        await flushPromises();

        const newModelValue: IDockerRegistrySelectorModel = {
          dockerRegistryName: "ngc-secret-1",
        };

        dockerRegistrySelectorCmp.vm.emitChanges(newModelValue);

        expect(dockerRegistrySelectorCmp.emitted()).toHaveProperty(EDockerRegistrySelectorEvents.UpdateModelValue);
        const emittedEvents = dockerRegistrySelectorCmp.emitted(EDockerRegistrySelectorEvents.UpdateModelValue);
        if (emittedEvents) {
          expect(emittedEvents[0]).toEqual([newModelValue]);
        }
      });
    });

    describe("onCredentialChanged", () => {
      test("should emit changes when credential is selected", async () => {
        mountComponent();
        await flushPromises();

        const selectedOption = {
          value: "ngc-secret-1",
          label: "ngc-secret-1",
        };

        dockerRegistrySelectorCmp.vm.onCredentialChanged(selectedOption);

        expect(dockerRegistrySelectorCmp.emitted()).toHaveProperty(EDockerRegistrySelectorEvents.UpdateModelValue);
        const emittedEvents = dockerRegistrySelectorCmp.emitted(EDockerRegistrySelectorEvents.UpdateModelValue);
        expect(emittedEvents).toBeDefined();
        if (emittedEvents) {
          expect(emittedEvents[0]).toEqual([
            {
              dockerRegistryName: "ngc-secret-1",
            },
          ]);
        }
      });
    });
  });

  describe("Props", () => {
    test("should accept and use modelValue prop", async () => {
      const modelValue: IDockerRegistrySelectorModel = {
        dockerRegistryName: "ngc-secret-1",
      };

      mountComponent({ modelValue });
      await flushPromises();

      expect(dockerRegistrySelectorCmp.props("modelValue")).toEqual(modelValue);
    });

    test("should accept and use clusterId prop", async () => {
      mountComponent({ clusterId: "cluster-456" });
      await flushPromises();

      expect(dockerRegistrySelectorCmp.props("clusterId")).toBe("cluster-456");
    });

    test("should use default value for isRequired prop when not provided", async () => {
      mountComponent({ isRequired: undefined });
      await flushPromises();

      expect(dockerRegistrySelectorCmp.props("isRequired")).toBe(true);
    });

    test("should accept false for isRequired prop", async () => {
      mountComponent({ isRequired: false });
      await flushPromises();

      expect(dockerRegistrySelectorCmp.props("isRequired")).toBe(false);
    });

    test("should accept and use scopeModel prop", async () => {
      mountComponent();
      await flushPromises();

      expect(dockerRegistrySelectorCmp.props("scopeModel")).toEqual(mockScopeModel);
    });
  });
});
