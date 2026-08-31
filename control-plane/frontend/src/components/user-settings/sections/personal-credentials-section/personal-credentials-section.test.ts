/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { useClusterStore } from "@/stores/cluster.store";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import PersonalCredentialsSection from "./personal-credentials-section.vue";

vi.mock("@/services/control-plane/credential.service/credential.service", () => ({
  credentialService: {
    listCredentials: vi.fn(),
    remove: vi.fn(),
    createDockerRegistry: vi.fn(),
    createGeneralCredential: vi.fn(),
    createNgcApiKey: vi.fn(),
    prepareGenericSecretForServerRequest: vi.fn((req: unknown) => req),
  },
}));

const mockDockerRegistryCred: CredentialsListResponseEntry = {
  meta: { id: "1", name: "docker-cred", kind: AssetKind.DockerRegistry },
} as CredentialsListResponseEntry;

const mockGenericSecretCred: CredentialsListResponseEntry = {
  meta: { id: "2", name: "generic-secret-cred", kind: AssetKind.GenericSecret },
} as CredentialsListResponseEntry;

const mockNgcApiKeyCred: CredentialsListResponseEntry = {
  meta: { id: "3", name: "ngc-api-key-cred", kind: AssetKind.NgcApiKey },
} as CredentialsListResponseEntry;

const mockAccessKeyCred: CredentialsListResponseEntry = {
  meta: { id: "4", name: "access-key-cred", kind: AssetKind.AccessKey },
} as CredentialsListResponseEntry;

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("personal-credentials-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof PersonalCredentialsSection>>;
  let clusterStore: ReturnType<typeof useClusterStore>;

  const listCredentialsMock = credentialService.listCredentials as ReturnType<typeof vi.fn>;

  beforeEach(() => {
    clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ clusters: [{ uuid: "1", name: "test-cluster", tenantId: 1, version: "1.0.0" }] });
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);

    listCredentialsMock.mockResolvedValue([
      mockDockerRegistryCred,
      mockGenericSecretCred,
      mockNgcApiKeyCred,
      mockAccessKeyCred,
    ]);

    wrapper = mount(PersonalCredentialsSection, {
      global: {
        plugins: [Quasar, pinia],
        stubs: {
          RunaiExpansionItem: {
            template: "<div><slot /></div>",
            props: ["label", "subheader"],
          },
          RunaiSection: { template: "<div><slot /></div>" },
          NotificationAlertBar: { template: "<div />" },
          PersonalCredentialItem: { template: '<div class="personal-credential-item" />', props: ["credential"] },
          RunaiBtn: { template: "<div><slot /></div>" },
          CredentialDropdown: { template: "<div />", props: ["visibleCreateOptions"] },
          RunaiDeleteModal: { template: "<div />" },
          CredentialCreationModal: { template: "<div />" },
        },
      },
    });
    wrapper.vm.$q.notify = vi.fn();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("getPersonalCredentials", () => {
    test("fetches credentials with tenant scope and private-only filter", async () => {
      await flushPromises();

      expect(listCredentialsMock).toHaveBeenCalledWith({
        scope: Scope.Tenant,
        privateOnly: true,
      });
    });

    test("includes NgcApiKey credentials in the filtered results", async () => {
      await flushPromises();

      const credentialKinds = wrapper.vm.credentials.map((c: CredentialsListResponseEntry) => c.meta.kind);
      expect(credentialKinds).toContain(AssetKind.NgcApiKey);
    });

    test("includes DockerRegistry and GenericSecret credentials", async () => {
      await flushPromises();

      const credentialKinds = wrapper.vm.credentials.map((c: CredentialsListResponseEntry) => c.meta.kind);
      expect(credentialKinds).toContain(AssetKind.DockerRegistry);
      expect(credentialKinds).toContain(AssetKind.GenericSecret);
    });

    test("filters out unsupported credential kinds like AccessKey", async () => {
      await flushPromises();

      const credentialKinds = wrapper.vm.credentials.map((c: CredentialsListResponseEntry) => c.meta.kind);
      expect(credentialKinds).not.toContain(AssetKind.AccessKey);
    });

    test("returns exactly the three supported credential types", async () => {
      await flushPromises();

      expect(wrapper.vm.credentials).toHaveLength(3);
    });
  });

  describe("visibleCreateOptions", () => {
    test("includes NGC API key in the create dropdown options", () => {
      const options = wrapper.vm.visibleCreateOptions as Set<CREDENTIAL_ROUTE_NAMES>;
      expect(options.has(CREDENTIAL_ROUTE_NAMES.CREDENTIAL_NGC_API_KEY_NEW)).toBe(true);
    });

    test("includes Docker Registry in the create dropdown options", () => {
      const options = wrapper.vm.visibleCreateOptions as Set<CREDENTIAL_ROUTE_NAMES>;
      expect(options.has(CREDENTIAL_ROUTE_NAMES.CREDENTIAL_DOCKER_REGISTRY_NEW)).toBe(true);
    });

    test("includes Generic Secret in the create dropdown options", () => {
      const options = wrapper.vm.visibleCreateOptions as Set<CREDENTIAL_ROUTE_NAMES>;
      expect(options.has(CREDENTIAL_ROUTE_NAMES.CREDENTIAL_GENERIC_SECRET_NEW)).toBe(true);
    });
  });

  describe("createCredential", () => {
    test("calls createNgcApiKey when credential type is NgcApiKey", async () => {
      await flushPromises();

      const createNgcApiKeyMock = credentialService.createNgcApiKey as ReturnType<typeof vi.fn>;
      const mockModel = { meta: { name: "test-ngc-key", scope: Scope.Tenant }, spec: { ngc_api_key: "key123" } };

      wrapper.vm.credentialToCreate = AssetKind.NgcApiKey;
      await wrapper.vm.createCredential(mockModel);

      expect(createNgcApiKeyMock).toHaveBeenCalledWith(mockModel);
    });

    test("calls createDockerRegistry when credential type is DockerRegistry", async () => {
      await flushPromises();

      const createDockerRegistryMock = credentialService.createDockerRegistry as ReturnType<typeof vi.fn>;
      const mockModel = {
        meta: { name: "test-docker", scope: Scope.Tenant },
        spec: { url: "https://docker.io", username: "u", password: "p" },
      };

      wrapper.vm.credentialToCreate = AssetKind.DockerRegistry;
      await wrapper.vm.createCredential(mockModel);

      expect(createDockerRegistryMock).toHaveBeenCalledWith(mockModel);
    });

    test("calls createGeneralCredential when credential type is GenericSecret", async () => {
      await flushPromises();

      const createGeneralCredentialMock = credentialService.createGeneralCredential as ReturnType<typeof vi.fn>;
      const mockModel = {
        meta: { name: "test-secret", scope: Scope.Tenant },
        spec: { keyValuePairs: [{ key: "k", value: "v" }] },
      };

      wrapper.vm.credentialToCreate = AssetKind.GenericSecret;
      await wrapper.vm.createCredential(mockModel);

      expect(createGeneralCredentialMock).toHaveBeenCalled();
    });

    test("refreshes credentials list after successful creation", async () => {
      await flushPromises();
      listCredentialsMock.mockClear();

      const mockModel = { meta: { name: "test-ngc-key", scope: Scope.Tenant }, spec: { ngc_api_key: "key123" } };

      wrapper.vm.credentialToCreate = AssetKind.NgcApiKey;
      await wrapper.vm.createCredential(mockModel);
      await flushPromises();

      expect(listCredentialsMock).toHaveBeenCalledTimes(1);
    });
  });

  describe("deleteCredential", () => {
    test("calls remove with the correct id and kind", async () => {
      await flushPromises();

      const removeMock = credentialService.remove as ReturnType<typeof vi.fn>;
      wrapper.vm.credentialToDelete = { id: "3", name: "ngc-api-key-cred", kind: AssetKind.NgcApiKey };

      await wrapper.vm.deleteCredential();

      expect(removeMock).toHaveBeenCalledWith("3", AssetKind.NgcApiKey);
    });

    test("removes the credential from the local list on success", async () => {
      await flushPromises();
      expect(wrapper.vm.credentials).toHaveLength(3);

      wrapper.vm.credentialToDelete = { id: "3", name: "ngc-api-key-cred", kind: AssetKind.NgcApiKey };
      await wrapper.vm.deleteCredential();

      expect(wrapper.vm.credentials).toHaveLength(2);
      const ids = wrapper.vm.credentials.map((c: CredentialsListResponseEntry) => c.meta.id);
      expect(ids).not.toContain("3");
    });
  });
});
