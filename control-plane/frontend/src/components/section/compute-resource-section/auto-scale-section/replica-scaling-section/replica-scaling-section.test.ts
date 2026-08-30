import { describe, it, expect } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import ReplicaScalingSection from "@/components/section/compute-resource-section/auto-scale-section/replica-scaling-section/replica-scaling-section.vue";
import {
  EReplicaScalingSectionEvents,
  type IReplicaScalingSection,
} from "@/components/section/compute-resource-section/auto-scale-section/replica-scaling-section/replica-scaling-section.model";
import { Quasar } from "quasar";
import { type AutoScalingRules } from "@/swagger-models/assets-service-client";
import { errorMessages } from "@/common/error-message.constant";

interface IReplicaScalingProps {
  modelValue: IReplicaScalingSection;
  disabled: boolean;
  policyRules: AutoScalingRules | null;
  isAutoScaleEnabled: boolean;
}
describe("replica-scaling-section.vue", () => {
  const baseProps: IReplicaScalingProps = {
    modelValue: { minReplicas: 1, maxReplicas: 5 },
    disabled: false,
    policyRules: null,
    isAutoScaleEnabled: true,
  };

  function getMountedComponent(props: IReplicaScalingProps): VueWrapper<InstanceType<typeof ReplicaScalingSection>> {
    return mount(ReplicaScalingSection, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  it("renders with correct initial values", () => {
    const wrapper = getMountedComponent(baseProps);
    expect(wrapper.props().modelValue.minReplicas).toBe(1);
    expect(wrapper.props().modelValue.maxReplicas).toBe(5);
  });

  describe("Computes", () => {
    it("computed minReplicasAllowed returns 0 if auto scale enabled", () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: true });
      expect(wrapper.vm.minReplicasAllowed).toBe(0);
    });

    it("computed minReplicasAllowed returns 1 if auto scale disabled", () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: false });
      expect(wrapper.vm.minReplicasAllowed).toBe(1);
    });

    it("computed minValueForMaxReplicas returns 1 if minReplicas is undefined or 0", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: undefined, maxReplicas: 5 } });
      expect(wrapper.vm.minValueForMaxReplicas).toBe(1);

      const wrapper2 = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 0, maxReplicas: 5 } });
      expect(wrapper2.vm.minValueForMaxReplicas).toBe(1);
    });

    it("computed minValueForMaxReplicas returns minReplicas if defined and not 0", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 3, maxReplicas: 5 } });
      expect(wrapper.vm.minValueForMaxReplicas).toBe(3);
    });
  });

  describe("Methods", () => {
    it("isMinReplicasValid returns true for valid value", () => {
      const wrapper = getMountedComponent(baseProps);
      expect(wrapper.vm.isMinReplicasValid(1)).toBe(true);
    });

    it("isMinReplicasValid returns error message for invalid value", () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: false });
      expect(wrapper.vm.isMinReplicasValid(0)).toContain("Enter a value equal or above 1");
    });

    it("isMaxMorThanMin returns true if max >= min", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 3 } });
      expect(wrapper.vm.isMaxMorThanMin()).toBe(true);
    });

    it("isMaxMorThanMin returns error message if max < min", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 5, maxReplicas: 2 } });
      expect(wrapper.vm.isMaxMorThanMin()).toContain(errorMessages.EQUAL_TO_OR_HIGHER_THAN_THE_MINIMUM);
    });

    it("isMaxMoreThanMaxReplicas returns true if maxReplicas <= 10000", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 1, maxReplicas: 10000 } });
      expect(wrapper.vm.isMaxMoreThanMaxReplicas()).toBe(true);
    });

    it("isMaxMoreThanMaxReplicas returns error message if maxReplicas > 10000", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 1, maxReplicas: 10001 } });
      expect(wrapper.vm.isMaxMoreThanMaxReplicas()).toContain("Enter a number equal to or lower than");
    });

    it("isMinReplicasAllowedZero returns true if auto scale enabled", () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: true });
      expect(wrapper.vm.isMinReplicasAllowedZero(0)).toBe(true);
    });

    it("isMinReplicasAllowedZero returns error message if value is 0 and auto scale disabled", () => {
      const wrapper = getMountedComponent({ ...baseProps, isAutoScaleEnabled: false });
      expect(wrapper.vm.isMinReplicasAllowedZero(0)).toContain(errorMessages.MIN_REPLICAS_ALLOWED_ZERO);
    });
  });

  describe("Emits", () => {
    it("updateMax emits update event with correct values", async () => {
      const wrapper = getMountedComponent(baseProps);
      wrapper.vm.updateMax("7");
      const expectedEmit = [
        [
          {
            minReplicas: 1,
            maxReplicas: 7,
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin emits correct values when minReplicas is increased above maxReplicas", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 2 } });
      wrapper.vm.updateMin("10");
      const expectedEmit = [
        [
          {
            minReplicas: 10,
            maxReplicas: 10,
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin emits correct values when minReplicas is set to 0 (does not force maxReplicas to 1)", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 2 } });
      wrapper.vm.updateMin("0");
      const expectedEmit = [
        [
          {
            minReplicas: 0,
            maxReplicas: 2, // maxReplicas remains unchanged
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin emits correct values when minReplicas is set to a value less than maxReplicas", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 5, maxReplicas: 10 } });
      wrapper.vm.updateMin("3");
      const expectedEmit = [
        [
          {
            minReplicas: 3,
            maxReplicas: 10, // maxReplicas remains unchanged
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin emits correct values when minReplicas is set to undefined (empty string)", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 5, maxReplicas: 10 } });
      wrapper.vm.updateMin("");
      const expectedEmit = [
        [
          {
            minReplicas: undefined,
            maxReplicas: 10, // maxReplicas remains unchanged
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin emits correct values when minReplicas is set to the same value as maxReplicas", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 3 } });
      wrapper.vm.updateMin("3");
      const expectedEmit = [
        [
          {
            minReplicas: 3,
            maxReplicas: 3, // maxReplicas remains unchanged
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin emits correct values when minReplicas is set to a value and maxReplicas is undefined", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: undefined } });
      wrapper.vm.updateMin("5");
      const expectedEmit = [
        [
          {
            minReplicas: 5,
            maxReplicas: 5,
          },
        ],
      ];
      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin does not change maxReplicas when new minReplicas is less than current maxReplicas", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 10 } });
      wrapper.vm.updateMin("5");
      const expectedEmit = [
        [
          {
            minReplicas: 5,
            maxReplicas: 10, // maxReplicas remains unchanged
          },
        ],
      ];

      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin does not change maxReplicas when new minReplicas is equal to current maxReplicas", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 4, maxReplicas: 5 } });
      wrapper.vm.updateMin("5");
      const expectedEmit = [
        [
          {
            minReplicas: 5,
            maxReplicas: 5, // maxReplicas remains unchanged
          },
        ],
      ];

      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin does not change maxReplicas when new minReplicas is 0", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 7 } });
      wrapper.vm.updateMin("0");
      const expectedEmit = [
        [
          {
            minReplicas: 0,
            maxReplicas: 7, // maxReplicas remains unchanged
          },
        ],
      ];

      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });

    it("updateMin does not change maxReplicas when new minReplicas is undefined (empty string)", () => {
      const wrapper = getMountedComponent({ ...baseProps, modelValue: { minReplicas: 2, maxReplicas: 7 } });
      wrapper.vm.updateMin("");
      const expectedEmit = [
        [
          {
            minReplicas: undefined,
            maxReplicas: 7, // maxReplicas remains unchanged
          },
        ],
      ];

      expect(wrapper.emitted()[EReplicaScalingSectionEvents.UpdateModelValue]).toEqual(expectedEmit);
    });
  });

  describe("Policy Rules", () => {
    const policyRules = {
      minReplicas: {
        canEdit: false,
        min: 2,
        max: 5,
        step: 2,
        required: true,
      },
      maxReplicas: {
        canEdit: true,
        min: 3,
        max: 10,
        step: 1,
        required: false,
      },
    };

    function getMinInput(wrapper: VueWrapper<InstanceType<typeof ReplicaScalingSection>>) {
      return wrapper.findAllComponents({ name: "policy-number-input" })[0];
    }

    function getMaxInput(wrapper: VueWrapper<InstanceType<typeof ReplicaScalingSection>>) {
      return wrapper.findAllComponents({ name: "policy-number-input" })[1];
    }

    it("disables minReplicas input when canEdit is false", () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules,
      });
      const minInput = getMinInput(wrapper);
      expect(minInput.vm.disableInput).toBe(true);
    });

    it("disables maxReplicas input when canEdit is false", () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules: {
          ...policyRules,
          maxReplicas: {
            ...policyRules.maxReplicas,
            canEdit: false,
          },
        },
      });
      const maxInput = getMaxInput(wrapper);
      expect(maxInput.vm.disableInput).toBe(true);
    });

    it("applies min, max, and step from policyRules to minReplicas input", () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules,
      });
      const minInput = getMinInput(wrapper);
      expect(minInput.vm.min).toBe(2);
      expect(minInput.vm.max).toBe(5);
      expect(minInput.vm.step).toBe(2);
    });

    it("applies min, max, and step from policyRules to maxReplicas input", () => {
      const wrapper = getMountedComponent({
        ...baseProps,
        policyRules,
      });
      const maxInput = getMaxInput(wrapper);
      expect(maxInput.vm.min).toBe(3);
      expect(maxInput.vm.max).toBe(10);
      expect(maxInput.vm.step).toBe(1);
    });
  });
});
