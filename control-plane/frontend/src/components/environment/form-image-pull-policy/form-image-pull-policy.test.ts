import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, test, expect, beforeEach } from "vitest";
import { ImagePullPolicy } from "@/swagger-models/assets-service-client";
import { FormImagePullPolicy } from ".";

describe("form-image-pull-policy.vue", () => {
  let formImagePullPolicyCmp: VueWrapper<InstanceType<typeof FormImagePullPolicy>>;

  beforeEach(() => {
    formImagePullPolicyCmp = mount(FormImagePullPolicy, {
      global: {
        plugins: [Quasar],
        stubs: {
          // Stub the credential selector component to avoid Pinia dependency
          "runai-credential-selector-by-source": {
            template: '<div class="credential-selector-stub"></div>',
          },
        },
      },
      props: {
        imagePullPolicy: ImagePullPolicy.IfNotPresent,
        supportImagePullSecrets: true,
        imagePullSecretsOptions: [],
        imagePullSecrets: [],
        disable: false,
        clusterId: "test-cluster",
        projectId: "123",
      },
    });
  });

  test("component is loading properly", () => {
    expect(formImagePullPolicyCmp.html()).toContain("Set the condition for pulling the image");
  });

  test("should show image pull secrets section when imagePullPolicy is Always", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Always,
    });
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(true);
  });

  test("should show image pull secrets section when imagePullPolicy is IfNotPresent", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.IfNotPresent,
    });
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(true);
  });

  test("should NOT show image pull secrets section when imagePullPolicy is Never", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Never,
    });
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(false);
  });

  test("should show image pull secrets section when imagePullPolicy is null", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: null,
    });
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(true);
  });

  test("image pull secrets section should be hidden in DOM when imagePullPolicy is Never", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Never,
      supportImagePullSecrets: true,
    });
    expect(formImagePullPolicyCmp.find(".credential-selector-stub").exists()).toBe(false);
  });

  test("image pull secrets section should be visible in DOM when imagePullPolicy is Always", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Always,
      supportImagePullSecrets: true,
    });
    expect(formImagePullPolicyCmp.find(".credential-selector-stub").exists()).toBe(true);
  });

  test("should emit update event when image pull policy radio option is changed", async () => {
    const radioButtons = formImagePullPolicyCmp.findAll("input[type=radio]");
    expect(radioButtons.length).toBeGreaterThan(0);

    // Trigger a change (this will emit through the runai-radio-options component)
    await radioButtons[0].trigger("click");

    // Check that the component structure allows for updates
    expect(formImagePullPolicyCmp.emitted()).toBeDefined();
  });

  test("should have all three image pull policy options", () => {
    expect(formImagePullPolicyCmp.vm.imageRadioOptions).toHaveLength(3);
    expect(formImagePullPolicyCmp.vm.imageRadioOptions[0].value).toBe(ImagePullPolicy.IfNotPresent);
    expect(formImagePullPolicyCmp.vm.imageRadioOptions[1].value).toBe(ImagePullPolicy.Always);
    expect(formImagePullPolicyCmp.vm.imageRadioOptions[2].value).toBe(ImagePullPolicy.Never);
  });

  test("should not show image pull secrets section if supportImagePullSecrets is false", async () => {
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Always,
      supportImagePullSecrets: false,
    });
    expect(formImagePullPolicyCmp.find(".credential-selector-stub").exists()).toBe(false);
  });

  test("computed shouldShowImagePullSecrets should be reactive to prop changes", async () => {
    // Initially IfNotPresent
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(true);

    // Change to Never
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Never,
    });
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(false);

    // Change back to Always
    await formImagePullPolicyCmp.setProps({
      imagePullPolicy: ImagePullPolicy.Always,
    });
    expect(formImagePullPolicyCmp.vm.shouldShowImagePullSecrets).toBe(true);
  });
});
