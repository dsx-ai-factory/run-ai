import { mount } from "@vue/test-utils";
import AcessKeySecretSection from "./access-key-secret-section.vue";
import { beforeEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { AccessKeyCreationSpec } from "@/swagger-models/assets-service-client";

describe("access-key-secret-section.vue", () => {
  let accessKeySecretCmp: VueWrapper<InstanceType<typeof AcessKeySecretSection>>;
  const props = {
    secretDetails: {
      existingSecretName: null,
      accessKeyId: "some-ac-id",
      secretAccessKey: "some-ac-secret",
    } as AccessKeyCreationSpec,
    existingSecretNames: ["access-key-1", "access-key-2"] as string[],
    loading: false as boolean,
    projectId: null,
  };
  beforeEach(() => {
    accessKeySecretCmp = mount(AcessKeySecretSection, {
      global: {
        plugins: [Quasar],
      },
      props,
      data() {
        return {
          isExistingSecret: false as boolean,
          isPassHidden: true as boolean,
        };
      },
    });
  });

  test("the component will emit the updated secretDetails on a value change", async () => {
    accessKeySecretCmp.find("[aid=access-key-id-input]").setValue("new-ac-id");
    await accessKeySecretCmp.vm.$nextTick();
    const expectedEmit: AccessKeyCreationSpec[][] = [[{ ...props.secretDetails, accessKeyId: "new-ac-id" }]];
    expect(accessKeySecretCmp.emitted("update-secret-details")).toEqual(expectedEmit);
  });
});
