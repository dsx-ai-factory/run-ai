import { mount } from "@vue/test-utils";
import GitDataMount from "./git-data-mount-section.vue";
import { beforeEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { GitAssetSpec, Password } from "@/swagger-models/assets-service-client";

describe("git-data-mount.vue", () => {
  let gitDataMountCmp: VueWrapper<InstanceType<typeof GitDataMount>>;
  const props = {
    supportGitSyncOneParameter: false,
    modelValue: {
      repository: "https://default.com",
      branch: null,
      revision: null,
      path: "/default",
      passwordAssetId: "default-2",
    } as GitAssetSpec,
    credentials: [
      {
        meta: { name: "test-name-101", id: "test-id-101" },
        spec: { existingSecretName: "test-secret-name", user: "test-user", password: "test-password" },
      },
    ] as Password[],
    loadingCredentials: false as boolean,
  };

  beforeEach(() => {
    gitDataMountCmp = mount(GitDataMount, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  });

  test("the component will emit the updated model-value on a value change", async () => {
    gitDataMountCmp.find("[aid=repository-input]").setValue("https://test.com");
    await gitDataMountCmp.vm.$nextTick();
    const expectedEmit: GitAssetSpec[][] = [[{ ...props.modelValue, repository: "https://test.com" }]];
    expect(gitDataMountCmp.emitted("update:model-value")).toEqual(expectedEmit);
  });
});
