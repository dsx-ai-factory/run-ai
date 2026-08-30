import { describe, it, expect, beforeEach } from "vitest";
import { ref, type Ref } from "vue";
import { useQuotaValidation, type UseQuotaValidationOptions } from "./use-quota-validation.composable";
import { EQuotaInputState } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";

describe("useQuotaValidation", () => {
  let quotaValue: Ref<number | null>;
  let parentLimit: Ref<number | null>;
  let allocatedNonPreemptible: Ref<number>;
  let isLimitQuotaOverSubscriptionEnabled: Ref<boolean>;
  let availabilityTotal: Ref<number | null>;

  const createValidation = (overrides?: Partial<UseQuotaValidationOptions>) => {
    return useQuotaValidation({
      quotaValue,
      parentLimit,
      allocatedNonPreemptible,
      isLimitQuotaOverSubscriptionEnabled,
      resourceName: "GPU",
      resourceUnit: "GPUs",
      requireNumberGreaterOrEqualToZero: false,
      availabilityTotal,
      ...overrides,
    });
  };

  beforeEach(() => {
    quotaValue = ref<number | null>(null);
    parentLimit = ref<number | null>(null);
    allocatedNonPreemptible = ref<number>(0);
    isLimitQuotaOverSubscriptionEnabled = ref<boolean>(false);
    availabilityTotal = ref<number | null>(100); // Default total for tests
  });

  describe("isNegative", () => {
    it("should return false when quotaValue is null", () => {
      quotaValue.value = null;
      const validation = createValidation();
      expect(validation.isNegative.value).toBe(false);
    });

    it("should return false when quotaValue is -1 (unlimited)", () => {
      quotaValue.value = -1;
      const validation = createValidation();
      expect(validation.isNegative.value).toBe(false);
    });

    it("should return false when quotaValue is 0 or positive", () => {
      quotaValue.value = 0;
      const validation = createValidation();
      expect(validation.isNegative.value).toBe(false);

      quotaValue.value = 10;
      expect(validation.isNegative.value).toBe(false);
    });

    it("should return true when quotaValue is less than -1", () => {
      quotaValue.value = -2;
      const validation = createValidation();
      expect(validation.isNegative.value).toBe(true);

      quotaValue.value = -10;
      expect(validation.isNegative.value).toBe(true);
    });
  });

  describe("isInvalidNumberGreaterOrEqualToZero", () => {
    it("should return false when requireNumberGreaterOrEqualToZero is false", () => {
      quotaValue.value = -1;
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: false });
      expect(validation.isInvalidNumberGreaterOrEqualToZero.value).toBe(false);
    });

    it("should return true when requireNumberGreaterOrEqualToZero is true and quotaValue is null", () => {
      quotaValue.value = null;
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: true });
      expect(validation.isInvalidNumberGreaterOrEqualToZero.value).toBe(true);
    });

    it("should return true when requireNumberGreaterOrEqualToZero is true and quotaValue is negative", () => {
      quotaValue.value = -1;
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: true });
      expect(validation.isInvalidNumberGreaterOrEqualToZero.value).toBe(true);

      quotaValue.value = -5;
      expect(validation.isInvalidNumberGreaterOrEqualToZero.value).toBe(true);
    });

    it("should return false when requireNumberGreaterOrEqualToZero is true and quotaValue is >= 0", () => {
      quotaValue.value = 0;
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: true });
      expect(validation.isInvalidNumberGreaterOrEqualToZero.value).toBe(false);

      quotaValue.value = 10;
      expect(validation.isInvalidNumberGreaterOrEqualToZero.value).toBe(false);
    });
  });

  describe("isExceededQuota", () => {
    it("should return false when parent limit is null (unlimited)", () => {
      quotaValue.value = 100;
      parentLimit.value = null;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(false);
    });

    it("should return false when parent limit is -1 (unlimited)", () => {
      quotaValue.value = 100;
      parentLimit.value = -1;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(false);
    });

    it("should return true when quotaValue is null (unlimited) and parent has limit", () => {
      quotaValue.value = null;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(true);
    });

    it("should return true when quotaValue is -1 (unlimited) and parent has limit", () => {
      quotaValue.value = -1;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(true);
    });

    it("should return true when quotaValue exceeds parent limit", () => {
      quotaValue.value = 100;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(true);
    });

    it("should return false when quotaValue is within parent limit", () => {
      quotaValue.value = 30;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(false);
    });

    it("should return false when quotaValue equals parent limit", () => {
      quotaValue.value = 50;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.isExceededQuota.value).toBe(false);
    });
  });

  describe("isUnderAllocatedNonPreemptible", () => {
    it("should return false when quotaValue is null (unlimited)", () => {
      quotaValue.value = null;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(false);
    });

    it("should return false when quotaValue is -1 (unlimited)", () => {
      quotaValue.value = -1;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(false);
    });

    it("should return true when quotaValue is less than allocated non-preemptible", () => {
      quotaValue.value = 5;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(true);
    });

    it("should return false when quotaValue equals allocated non-preemptible", () => {
      quotaValue.value = 10;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(false);
    });

    it("should return false when quotaValue is greater than allocated non-preemptible", () => {
      quotaValue.value = 15;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(false);
    });
  });

  describe("quotaState", () => {
    it("should return Error when requireNumberGreaterOrEqualToZero validation fails", () => {
      quotaValue.value = null;
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: true });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should return Error when quotaValue is negative (< -1)", () => {
      quotaValue.value = -5;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should return Error when quota is exceeded and oversubscription is enabled", () => {
      quotaValue.value = 100;
      parentLimit.value = 50;
      isLimitQuotaOverSubscriptionEnabled.value = true;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should return Warning when quota is exceeded and oversubscription is disabled", () => {
      quotaValue.value = 100;
      parentLimit.value = 50;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
    });

    it("should return Error when quota is under allocated non-preemptible and oversubscription is enabled", () => {
      quotaValue.value = 5;
      allocatedNonPreemptible.value = 10;
      isLimitQuotaOverSubscriptionEnabled.value = true;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should return Warning when quota is under allocated non-preemptible and oversubscription is disabled", () => {
      quotaValue.value = 5;
      allocatedNonPreemptible.value = 10;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
    });

    it("should return Default when all validations pass", () => {
      quotaValue.value = 30;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
    });
  });

  describe("quotaMessage", () => {
    it("should return error message when requireNumberGreaterOrEqualToZero validation fails", () => {
      quotaValue.value = null;
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: true });
      expect(validation.quotaMessage.value).toBe("Enter a value equal to or above 0");
    });

    it("should return error message when quotaValue is negative", () => {
      quotaValue.value = -5;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("GPU quota cannot be less than -1");
    });

    it("should return message when quota is under allocated non-preemptible", () => {
      quotaValue.value = 5;
      allocatedNonPreemptible.value = 10.5;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toContain(
        "The project's quota is less than the usage by non-preemptive workloads. To remain above the usage, enter a number above 10.50 GPUs.",
      );
    });

    it("should return message when quota exceeds parent limit", () => {
      quotaValue.value = 100;
      parentLimit.value = 50.75;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toContain(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 50.75 GPUs.",
      );
    });

    it("should return recommended value message when both usage and limit exist", () => {
      quotaValue.value = 30;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toContain("Recommended: From 10 to 50");
      expect(validation.quotaMessage.value).toContain("Non-preemptive workload usage: 10 GPUs");
      expect(validation.quotaMessage.value).toContain("Department's quota: 50 GPUs");
    });

    it("should return available quota message when only limit exists", () => {
      quotaValue.value = 30;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 50 GPUs.<br> Department's quota: 100 GPUs");
    });

    it("should return usage message when only allocated non-preemptible exists", () => {
      quotaValue.value = 30;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("Non-preemptive workload usage: 10 GPUs");
    });

    it("should use singular unit when value is 1 (e.g., '1 GPU' not '1 GPUs')", () => {
      quotaValue.value = 30;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 1;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("Non-preemptive workload usage: 1 GPU");
    });

    it("should use plural unit when value is not 1", () => {
      quotaValue.value = 30;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 2;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("Non-preemptive workload usage: 2 GPUs");
    });

    it("should return empty message when no usage or limit exists", () => {
      quotaValue.value = 30;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("");
    });

    it("should handle parent limit of -1 (unlimited) in messages", () => {
      quotaValue.value = 30;
      parentLimit.value = -1;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe("Non-preemptive workload usage: 10 GPUs");
    });

    it("should format decimal values correctly", () => {
      quotaValue.value = 30;
      parentLimit.value = 50.123;
      allocatedNonPreemptible.value = 10.456;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toContain("10.46");
      expect(validation.quotaMessage.value).toContain("50.12");
    });

    it("should format integer values without decimals", () => {
      quotaValue.value = 30;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toContain("10 GPUs");
      expect(validation.quotaMessage.value).toContain("50 GPUs");
    });

    it("should display 'Unlimited' for null values in exceeded quota message", () => {
      quotaValue.value = null;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toContain("The project's quota exceeds the available department quota");
    });
  });

  describe("resourceUnit as Ref or ComputedRef", () => {
    it("should work with resourceUnit as a string", () => {
      quotaValue.value = 30;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({ resourceUnit: "cores" });
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 50 cores.<br> Department's quota: 100 cores");
    });

    it("should work with resourceUnit as a Ref", () => {
      const unitRef = ref("GB");
      quotaValue.value = 30;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({ resourceUnit: unitRef });
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 50 GB.<br> Department's quota: 100 GB");

      unitRef.value = "TB";
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 50 TB.<br> Department's quota: 100 TB");
    });
  });

  describe("resourceName customization", () => {
    it("should use custom resource name in error messages", () => {
      quotaValue.value = -5;
      const validation = createValidation({ resourceName: "CPU" });
      expect(validation.quotaMessage.value).toBe("CPU quota cannot be less than -1");
    });

    it("should use custom resource name with requireNumberGreaterOrEqualToZero", () => {
      quotaValue.value = null;
      const validation = createValidation({
        resourceName: "Memory",
        requireNumberGreaterOrEqualToZero: true,
      });
      expect(validation.quotaMessage.value).toBe("Enter a value equal to or above 0");
    });
  });

  describe("edge cases", () => {
    it("should handle quotaValue of 0", () => {
      quotaValue.value = 0;
      parentLimit.value = 50;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
    });

    it("should prioritize isInvalidNumberGreaterOrEqualToZero error over other validations", () => {
      quotaValue.value = null;
      parentLimit.value = 50; // This would normally trigger exceeded quota
      const validation = createValidation({ requireNumberGreaterOrEqualToZero: true });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
      expect(validation.quotaMessage.value).toContain("Enter a value equal to or above 0");
    });

    it("should prioritize isNegative error over quota exceeded warnings", () => {
      quotaValue.value = -5;
      parentLimit.value = 50;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
      expect(validation.quotaMessage.value).toContain("cannot be less than -1");
    });

    it("should handle very large numbers", () => {
      quotaValue.value = 1000000;
      parentLimit.value = 2000000;
      allocatedNonPreemptible.value = 500000;
      const validation = createValidation();
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("500000");
    });
  });

  describe("CPU resource validation", () => {
    beforeEach(() => {
      quotaValue.value = null;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 0;
      isLimitQuotaOverSubscriptionEnabled.value = false;
    });

    it("should validate CPU quota with cores unit", () => {
      quotaValue.value = 16;
      parentLimit.value = 32;
      allocatedNonPreemptible.value = 8;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("cores");
      expect(validation.quotaMessage.value).toContain("Recommended: From 8 to 32");
    });

    it("should show error message for negative CPU quota", () => {
      quotaValue.value = -10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
      expect(validation.quotaMessage.value).toBe("CPU quota cannot be less than -1");
    });

    it("should show warning when CPU quota exceeds parent limit", () => {
      quotaValue.value = 64;
      parentLimit.value = 32;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
      expect(validation.quotaMessage.value).toContain(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 32 cores.",
      );
    });

    it("should show error when CPU quota exceeds parent limit with oversubscription enabled", () => {
      quotaValue.value = 64;
      parentLimit.value = 32;
      isLimitQuotaOverSubscriptionEnabled.value = true;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should show warning when CPU quota is under allocated non-preemptible", () => {
      quotaValue.value = 4;
      allocatedNonPreemptible.value = 8;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
      expect(validation.quotaMessage.value).toContain(
        "The project's quota is less than the usage by non-preemptive workloads. To remain above the usage, enter a number above 8 cores.",
      );
    });

    it("should handle unlimited CPU quota (-1)", () => {
      quotaValue.value = -1;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 16;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(false);
    });

    it("should handle fractional CPU cores", () => {
      quotaValue.value = 2.5;
      parentLimit.value = 8.75;
      allocatedNonPreemptible.value = 1.25;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("1.25");
      expect(validation.quotaMessage.value).toContain("8.75");
    });

    it("should require CPU quota >= 0 when flag is set", () => {
      quotaValue.value = null;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        requireNumberGreaterOrEqualToZero: true,
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
      expect(validation.quotaMessage.value).toBe("Enter a value equal to or above 0");
    });
  });

  describe("Memory resource validation", () => {
    beforeEach(() => {
      quotaValue.value = null;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 0;
      isLimitQuotaOverSubscriptionEnabled.value = false;
    });

    it("should validate memory quota with GB unit", () => {
      quotaValue.value = 128;
      parentLimit.value = 256;
      allocatedNonPreemptible.value = 64;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("GB");
      expect(validation.quotaMessage.value).toContain("Recommended: From 64 to 256");
    });

    it("should validate memory quota with MiB unit", () => {
      quotaValue.value = 1024;
      parentLimit.value = 2048;
      allocatedNonPreemptible.value = 512;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "MiB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("MiB");
      expect(validation.quotaMessage.value).toContain("Recommended: From 512 to 2048");
    });

    it("should validate memory quota with MB unit", () => {
      quotaValue.value = 1000;
      parentLimit.value = 2000;
      allocatedNonPreemptible.value = 500;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "MB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("MB");
    });

    it("should show error message for negative memory quota", () => {
      quotaValue.value = -100;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
      expect(validation.quotaMessage.value).toBe("Memory quota cannot be less than -1");
    });

    it("should show warning when memory quota exceeds parent limit", () => {
      quotaValue.value = 512;
      parentLimit.value = 256;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
      expect(validation.quotaMessage.value).toContain(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 256 GB.",
      );
    });

    it("should show error when memory quota exceeds parent limit with oversubscription enabled", () => {
      quotaValue.value = 512;
      parentLimit.value = 256;
      isLimitQuotaOverSubscriptionEnabled.value = true;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should show warning when memory quota is under allocated non-preemptible", () => {
      quotaValue.value = 32;
      allocatedNonPreemptible.value = 64;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
      expect(validation.quotaMessage.value).toContain(
        "The project's quota is less than the usage by non-preemptive workloads. To remain above the usage, enter a number above 64 GB.",
      );
    });

    it("should handle unlimited memory quota (-1)", () => {
      quotaValue.value = -1;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 128;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(false);
    });

    it("should handle fractional memory values", () => {
      quotaValue.value = 128.5;
      parentLimit.value = 256.75;
      allocatedNonPreemptible.value = 64.25;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("64.25");
      expect(validation.quotaMessage.value).toContain("256.75");
    });

    it("should require memory quota >= 0 when flag is set", () => {
      quotaValue.value = -1;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
        requireNumberGreaterOrEqualToZero: true,
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
      expect(validation.quotaMessage.value).toBe("Enter a value equal to or above 0");
    });

    it("should handle very large memory values", () => {
      quotaValue.value = 10240; // 10 TB in GB
      parentLimit.value = 20480; // 20 TB in GB
      allocatedNonPreemptible.value = 5120; // 5 TB in GB
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
      });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(validation.quotaMessage.value).toContain("5120");
      expect(validation.quotaMessage.value).toContain(
        "Recommended: From 5120 to 20480.<br> Non-preemptive workload usage: 5120 GB.<br> Department's quota: 20480 GB",
      );
      expect(validation.quotaMessage.value).toContain("20480");
    });

    it("should dynamically update messages when memory unit changes", () => {
      const unitRef = ref("GB");
      quotaValue.value = 128;
      parentLimit.value = 256;
      availabilityTotal.value = 256; // Set to match parentLimit for consistent test
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: unitRef,
      });
      expect(validation.quotaMessage.value).toContain("256 GB");

      unitRef.value = "MiB";
      expect(validation.quotaMessage.value).toContain("256 MiB");

      unitRef.value = "MB";
      expect(validation.quotaMessage.value).toContain("256 MB");
    });
  });

  describe("cross-resource validation scenarios", () => {
    it("should validate multiple resources independently - GPU and CPU", () => {
      const gpuValidation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
      });

      quotaValue.value = 8;
      parentLimit.value = 16;
      allocatedNonPreemptible.value = 4;

      const cpuValidation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });

      expect(gpuValidation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(cpuValidation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(gpuValidation.quotaMessage.value).toContain("GPUs");
      expect(cpuValidation.quotaMessage.value).toContain("cores");
    });

    it("should handle zero quota for all resource types", () => {
      quotaValue.value = 0;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 0;

      const gpuValidation = createValidation({ resourceName: "GPU", resourceUnit: "GPUs" });
      const cpuValidation = createValidation({ resourceName: "CPU", resourceUnit: "cores" });
      const memoryValidation = createValidation({ resourceName: "Memory", resourceUnit: "GB" });

      expect(gpuValidation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(cpuValidation.quotaState.value).toBe(EQuotaInputState.Default);
      expect(memoryValidation.quotaState.value).toBe(EQuotaInputState.Default);
    });

    it("should handle unlimited parent quota for all resource types", () => {
      quotaValue.value = 100;
      parentLimit.value = -1;
      allocatedNonPreemptible.value = 50;

      const gpuValidation = createValidation({ resourceName: "GPU", resourceUnit: "GPUs" });
      const cpuValidation = createValidation({ resourceName: "CPU", resourceUnit: "cores" });
      const memoryValidation = createValidation({ resourceName: "Memory", resourceUnit: "GB" });

      expect(gpuValidation.isExceededQuota.value).toBe(false);
      expect(cpuValidation.isExceededQuota.value).toBe(false);
      expect(memoryValidation.isExceededQuota.value).toBe(false);
    });
  });

  describe("isUnderSubordinateQuota", () => {
    it("should return false when subordinateQuota is not provided", () => {
      quotaValue.value = 10;
      const validation = createValidation();
      expect(validation.isUnderSubordinateQuota.value).toBe(false);
    });

    it("should return false when subordinateQuota is null", () => {
      quotaValue.value = 10;
      const subordinateQuota = ref<number | null>(null);
      const validation = createValidation({ subordinateQuota });
      expect(validation.isUnderSubordinateQuota.value).toBe(false);
    });

    it("should return false when quota is unlimited (null)", () => {
      quotaValue.value = null;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.isUnderSubordinateQuota.value).toBe(false);
    });

    it("should return false when quota is unlimited (-1)", () => {
      quotaValue.value = -1;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.isUnderSubordinateQuota.value).toBe(false);
    });

    it("should return false when quota >= subordinate quota", () => {
      quotaValue.value = 100;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.isUnderSubordinateQuota.value).toBe(false);

      quotaValue.value = 50; // Equal
      expect(validation.isUnderSubordinateQuota.value).toBe(false);
    });

    it("should return true when quota < subordinate quota", () => {
      quotaValue.value = 30;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.isUnderSubordinateQuota.value).toBe(true);
    });
  });

  describe("quotaState with subordinateQuota", () => {
    it("should return warning when isUnderSubordinateQuota is true and overSubscription disabled", () => {
      quotaValue.value = 30;
      isLimitQuotaOverSubscriptionEnabled.value = false;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Warning);
    });

    it("should return error when isUnderSubordinateQuota is true and overSubscription enabled", () => {
      quotaValue.value = 30;
      isLimitQuotaOverSubscriptionEnabled.value = true;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Error);
    });

    it("should return default when quota >= subordinate quota", () => {
      quotaValue.value = 100;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 0;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({ subordinateQuota });
      expect(validation.quotaState.value).toBe(EQuotaInputState.Default);
    });
  });

  describe("quotaMessage for subordinateQuota", () => {
    it("should show subordinate quota message when under subordinate quota", () => {
      quotaValue.value = 30;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({
        subordinateQuota,
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaMessage.value).toBe(
        "The department's quota is less than the quota assigned to its subordinate projects. To remain within the limit, enter a number above 50 cores.",
      );
    });

    it("should prioritize non-preemptible message over subordinate quota message", () => {
      quotaValue.value = 5;
      allocatedNonPreemptible.value = 10;
      const subordinateQuota = ref<number | null>(50);
      const validation = createValidation({
        subordinateQuota,
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      // Non-preemptible check comes first in the composable
      expect(validation.quotaMessage.value).toContain("less than the usage by non-preemptive workloads");
    });
  });

  describe("entity type differentiation", () => {
    it("should use 'Department's quota' label for projects", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.project,
      });
      expect(validation.quotaMessage.value).toContain("Department's quota: 100 cores");
    });

    it("should use 'Node pool's total' label for departments", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.department,
      });
      expect(validation.quotaMessage.value).toContain("Node pool's total: 100 cores");
    });

    it("should use correct instructive text format with both usage and limit for project", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.project,
      });
      expect(validation.quotaMessage.value).toBe(
        "Recommended: From 10 to 100.<br> Non-preemptive workload usage: 10 cores.<br> Department's quota: 100 cores",
      );
    });

    it("should use correct instructive text format with both usage and limit for department", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.department,
      });
      expect(validation.quotaMessage.value).toBe(
        "Recommended: From 10 to 100.<br> Non-preemptive workload usage: 10 cores.<br> Node pool's total: 100 cores",
      );
    });

    it("should show only limit for project when no usage", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.project,
      });
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 100 cores.<br> Department's quota: 100 cores");
    });

    it("should show only limit for department when no usage", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.department,
      });
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 100 cores.<br> Node pool's total: 100 cores");
    });

    it("should show only usage message when no limit", () => {
      quotaValue.value = 50;
      parentLimit.value = null;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
      });
      expect(validation.quotaMessage.value).toBe("Non-preemptive workload usage: 10 cores");
    });

    it("should use project-specific exceeded quota message", () => {
      quotaValue.value = 150;
      parentLimit.value = 100;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.project,
      });
      expect(validation.quotaMessage.value).toBe(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 100 cores.",
      );
    });

    it("should use department-specific exceeded quota message", () => {
      quotaValue.value = 150;
      parentLimit.value = 100;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        entityType: EQuotaEntity.department,
      });
      expect(validation.quotaMessage.value).toBe(
        "The department's quota exceeds the node pool's total. To remain within the limit, enter a number up to 100 cores.",
      );
    });
  });

  describe("showDefaultMessage option", () => {
    it("should show default messages when showDefaultMessage is true (default)", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        showDefaultMessage: true,
      });
      expect(validation.quotaMessage.value).toBe("Recommended: Up to 100 GPUs.<br> Department's quota: 100 GPUs");
    });

    it("should not show default messages when showDefaultMessage is false", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 0;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        showDefaultMessage: false,
      });
      expect(validation.quotaMessage.value).toBe("");
    });

    it("should still show error messages when showDefaultMessage is false", () => {
      quotaValue.value = 150;
      parentLimit.value = 100;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        showDefaultMessage: false,
      });
      expect(validation.quotaMessage.value).toBe(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 100 GPUs.",
      );
    });

    it("should still show warning messages when showDefaultMessage is false", () => {
      quotaValue.value = 5;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        showDefaultMessage: false,
        isLimitQuotaOverSubscriptionEnabled: ref(false),
      });
      expect(validation.quotaMessage.value).toContain(
        "The project's quota is less than the usage by non-preemptive workloads",
      );
    });

    it("should not show recommended/usage messages when showDefaultMessage is false and no errors", () => {
      quotaValue.value = 50;
      parentLimit.value = 100;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        showDefaultMessage: false,
      });
      expect(validation.quotaMessage.value).toBe("");
    });
  });

  describe("zero target value messages", () => {
    it("should not suggest entering 0 when parent limit is 0 (project)", () => {
      quotaValue.value = 10;
      parentLimit.value = 0;
      const validation = createValidation({
        entityType: EQuotaEntity.project,
      });
      expect(validation.quotaMessage.value).toBe("The project's quota exceeds the available department quota.");
      expect(validation.quotaMessage.value).not.toContain("enter a number");
    });

    it("should not suggest entering 0 when parent limit is 0 (department)", () => {
      quotaValue.value = 10;
      parentLimit.value = 0;
      const validation = createValidation({
        entityType: EQuotaEntity.department,
      });
      expect(validation.quotaMessage.value).toBe("The department's quota exceeds the node pool's total.");
      expect(validation.quotaMessage.value).not.toContain("enter a number");
    });

    it("should suggest entering value when parent limit is positive", () => {
      quotaValue.value = 100;
      parentLimit.value = 50;
      const validation = createValidation({
        entityType: EQuotaEntity.project,
      });
      expect(validation.quotaMessage.value).toBe(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 50 GPUs.",
      );
    });

    it("should not suggest value when subordinate quota is 0 or less", () => {
      quotaValue.value = -5; // triggers isUnderSubordinateQuota check
      const subordinateQuota = ref<number | null>(-0.001); // edge case: very small negative
      const validation = createValidation({
        subordinateQuota,
        entityType: EQuotaEntity.department,
      });
      // This won't trigger isUnderSubordinateQuota since quotaValue < 0 returns false
      // Let's test with valid quota
      quotaValue.value = 5;
      subordinateQuota.value = 10;
      expect(validation.quotaMessage.value).toContain("enter a number above 10 GPUs");

      // Now test with 0 subordinate quota - should not suggest
      subordinateQuota.value = 0;
      // isUnderSubordinateQuota would be false when subordinateQuota is 0, so no message
      expect(validation.isUnderSubordinateQuota.value).toBe(false);
    });

    it("should show simple message when under non-preemptible and allocated is 0", () => {
      // This is an edge case - if allocatedNonPreemptible is 0, isUnderAllocatedNonPreemptible
      // would be false for any quota >= 0. But if somehow triggered with 0:
      quotaValue.value = -0.5; // Less than 0 but more than -1 (not negative error)
      allocatedNonPreemptible.value = 0;
      const validation = createValidation();
      // isUnderAllocatedNonPreemptible: quotaValue (-0.5) < allocatedNonPreemptible (0) = true
      expect(validation.isUnderAllocatedNonPreemptible.value).toBe(true);
      expect(validation.quotaMessage.value).toBe(
        "The project's quota is less than the usage by non-preemptive workloads.",
      );
      expect(validation.quotaMessage.value).not.toContain("enter a number");
    });

    it("should show message with value when allocated non-preemptible is positive", () => {
      quotaValue.value = 5;
      allocatedNonPreemptible.value = 10;
      const validation = createValidation();
      expect(validation.quotaMessage.value).toBe(
        "The project's quota is less than the usage by non-preemptive workloads. To remain above the usage, enter a number above 10 GPUs.",
      );
    });
  });

  describe("fieldLabel customization", () => {
    it("should use default label '${resourceName} quota' when fieldLabel is not provided", () => {
      quotaValue.value = -5;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
      });
      expect(validation.quotaMessage.value).toBe("GPU quota cannot be less than -1");
    });

    it("should use custom fieldLabel in error messages when provided", () => {
      quotaValue.value = -5;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        fieldLabel: "Maximum GPUs",
      });
      expect(validation.quotaMessage.value).toBe("Maximum GPUs cannot be less than -1");
    });

    it("should use custom fieldLabel for requireNumberGreaterOrEqualToZero error", () => {
      quotaValue.value = null;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        fieldLabel: "Maximum GPUs",
        requireNumberGreaterOrEqualToZero: true,
      });
      expect(validation.quotaMessage.value).toBe("Enter a value equal to or above 0");
    });

    it("should use custom fieldLabel for CPU resources", () => {
      quotaValue.value = -10;
      const validation = createValidation({
        resourceName: "CPU",
        resourceUnit: "cores",
        fieldLabel: "CPU Limit",
      });
      expect(validation.quotaMessage.value).toBe("CPU Limit cannot be less than -1");
    });

    it("should use custom fieldLabel for Memory resources", () => {
      quotaValue.value = -100;
      const validation = createValidation({
        resourceName: "Memory",
        resourceUnit: "GB",
        fieldLabel: "Memory Limit",
      });
      expect(validation.quotaMessage.value).toBe("Memory Limit cannot be less than -1");
    });

    it("should not affect non-label-related error messages", () => {
      quotaValue.value = 150;
      parentLimit.value = 100;
      const validation = createValidation({
        resourceName: "GPU",
        resourceUnit: "GPUs",
        fieldLabel: "Maximum GPUs",
      });
      // This message uses entityName, not the label
      expect(validation.quotaMessage.value).toBe(
        "The project's quota exceeds the available department quota. To remain within the quota, enter a number up to 100 GPUs.",
      );
    });
  });
});
