import { describe, it, expect } from "vitest";
import type { UnifiedPolicyInfo } from "@/swagger-models/workloads-client";
import { distributedInferenceUtil } from "./distributed-inference.util";

const { hasNoWorkerPolicy, areLeaderAndWorkerPoliciesEqual, shouldCopyLeaderToWorker } = distributedInferenceUtil;

/**
 * Unit tests for the distributed inference policy utility functions.
 *
 * The logic being tested:
 * 1. No policy or no worker policy => copy leader data into worker
 * 2. Policy is equal between leader and worker (defaults AND rules) => copy leader data into worker
 * 3. Policy is NOT equal between leader and worker => do not change worker data, show mismatch alert
 *
 * Tests cover:
 * - hasNoWorkerPolicy: Checks if worker policy is missing or empty
 * - areLeaderAndWorkerPoliciesEqual: Deep comparison of leader and worker policies
 * - shouldCopyLeaderToWorker: Integration of the above logic
 */

describe("distributed-inference.util", () => {
  describe("hasNoWorkerPolicy", () => {
    it("should return true when policy is null", () => {
      expect(hasNoWorkerPolicy(null)).toBe(true);
    });

    it("should return true when policy is undefined", () => {
      expect(hasNoWorkerPolicy(undefined)).toBe(true);
    });

    it("should return true when worker is undefined", () => {
      const policy: UnifiedPolicyInfo = {
        master: { defaults: {}, rules: {} },
      };
      expect(hasNoWorkerPolicy(policy)).toBe(true);
    });

    it("should return true when worker is null", () => {
      const policy: UnifiedPolicyInfo = {
        master: { defaults: {}, rules: {} },
        worker: null,
      };
      expect(hasNoWorkerPolicy(policy)).toBe(true);
    });

    it("should return true when worker has no defaults and no rules", () => {
      const policy: UnifiedPolicyInfo = {
        master: { defaults: {}, rules: {} },
        worker: {},
      };
      expect(hasNoWorkerPolicy(policy)).toBe(true);
    });

    it("should return true when worker has null defaults and null rules", () => {
      const policy: UnifiedPolicyInfo = {
        master: { defaults: {}, rules: {} },
        worker: { defaults: null, rules: null },
      };
      expect(hasNoWorkerPolicy(policy)).toBe(true);
    });

    it("should return false when worker has defaults", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: { compute: { cpuCoreRequest: 1 } },
        },
      };
      expect(hasNoWorkerPolicy(policy)).toBe(false);
    });

    it("should return false when worker has rules", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          rules: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
        },
      };
      expect(hasNoWorkerPolicy(policy)).toBe(false);
    });

    it("should return false when worker has both defaults and rules", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: { compute: { cpuCoreRequest: 1 } },
          rules: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
        },
      };
      expect(hasNoWorkerPolicy(policy)).toBe(false);
    });
  });

  describe("areLeaderAndWorkerPoliciesEqual", () => {
    it("should return true when both policies are undefined", () => {
      const policy: UnifiedPolicyInfo = {};
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(true);
    });

    it("should return true when worker is null", () => {
      const policy: UnifiedPolicyInfo = {
        worker: null,
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(true);
    });

    it("should return true when both have empty objects", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: { leader: {}, worker: {} },
          rules: { leader: {}, worker: {} },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(true);
    });

    it("should return true when defaults and rules are identical", () => {
      const sharedDefaults = { compute: { cpuCoreRequest: 2, cpuMemoryRequest: "1G" } };
      const sharedRules = { compute: { cpuCoreRequest: { min: 1, max: 10 } } };

      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: { leader: sharedDefaults, worker: sharedDefaults },
          rules: { leader: sharedRules, worker: sharedRules },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(true);
    });

    it("should return true when defaults and rules are deeply equal (different object references)", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: { compute: { cpuCoreRequest: 2, cpuMemoryRequest: "1G" } },
            worker: { compute: { cpuCoreRequest: 2, cpuMemoryRequest: "1G" } },
          },
          rules: {
            leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
            worker: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
          },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(true);
    });

    it("should return false when defaults are different", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: { compute: { cpuCoreRequest: 2 } },
            worker: { compute: { cpuCoreRequest: 4 } },
          },
          rules: { leader: {}, worker: {} },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(false);
    });

    it("should return false when rules are different", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: { leader: {}, worker: {} },
          rules: {
            leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
            worker: { compute: { cpuCoreRequest: { min: 2, max: 20 } } },
          },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(false);
    });

    it("should return false when only leader has defaults", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: { compute: { cpuCoreRequest: 2 } },
            worker: {},
          },
          rules: { leader: {}, worker: {} },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(false);
    });

    it("should return false when only worker has rules", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: { leader: {}, worker: {} },
          rules: {
            leader: {},
            worker: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
          },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(false);
    });

    it("should return false when defaults are equal but rules are different", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: { compute: { cpuCoreRequest: 2 } },
            worker: { compute: { cpuCoreRequest: 2 } },
          },
          rules: {
            leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
            worker: { compute: { cpuCoreRequest: { min: 5, max: 50 } } },
          },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(false);
    });

    it("should return false when rules are equal but defaults are different", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: { compute: { cpuCoreRequest: 2 } },
            worker: { compute: { cpuCoreRequest: 8 } },
          },
          rules: {
            leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
            worker: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
          },
        },
      };
      expect(areLeaderAndWorkerPoliciesEqual(policy)).toBe(false);
    });
  });

  describe("shouldCopyLeaderToWorker", () => {
    describe("Scenario 1: No policy or no worker policy => should copy", () => {
      it("should return true when policy is null", () => {
        expect(shouldCopyLeaderToWorker(null)).toBe(true);
      });

      it("should return true when policy is undefined", () => {
        expect(shouldCopyLeaderToWorker(undefined)).toBe(true);
      });

      it("should return true when worker policy is missing", () => {
        const policy: UnifiedPolicyInfo = {
          master: { defaults: { compute: { cpuCoreRequest: 2 } } },
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(true);
      });

      it("should return true when worker has no defaults and no rules", () => {
        const policy: UnifiedPolicyInfo = {
          master: { defaults: { compute: { cpuCoreRequest: 2 } } },
          worker: {},
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(true);
      });
    });

    describe("Scenario 2: Policies are equal => should copy", () => {
      it("should return true when both policies are empty", () => {
        const policy: UnifiedPolicyInfo = {
          master: {},
          worker: {},
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(true);
      });

      it("should return true when policies are deeply equal", () => {
        const policy: UnifiedPolicyInfo = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 2 } },
            },
            rules: {
              leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
              worker: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
            },
          },
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(true);
      });
    });

    describe("Scenario 3: Policies are NOT equal => should NOT copy", () => {
      it("should return false when defaults differ", () => {
        const policy: UnifiedPolicyInfo = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 4 } },
            },
            rules: { leader: {}, worker: {} },
          },
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(false);
      });

      it("should return false when rules differ", () => {
        const policy: UnifiedPolicyInfo = {
          worker: {
            defaults: { leader: {}, worker: {} },
            rules: {
              leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
              worker: { compute: { cpuCoreRequest: { min: 2, max: 20 } } },
            },
          },
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(false);
      });

      it("should return false when both defaults and rules differ", () => {
        const policy: UnifiedPolicyInfo = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
            rules: {
              leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
              worker: { compute: { cpuCoreRequest: { min: 5, max: 50 } } },
            },
          },
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(false);
      });

      it("should return false when leader has policy but worker has different policy", () => {
        const policy: UnifiedPolicyInfo = {
          worker: {
            defaults: {
              leader: { compute: { gpuDevicesRequest: 1 } },
              worker: { compute: { gpuDevicesRequest: 2 } },
            },
            rules: {
              leader: { security: { privileged: { canEdit: true } } },
              worker: { security: { privileged: { canEdit: false } } },
            },
          },
        };
        expect(shouldCopyLeaderToWorker(policy)).toBe(false);
      });
    });
  });

  describe("Edge cases", () => {
    it("should handle complex nested objects correctly", () => {
      const complexDefaults = {
        compute: {
          cpuCoreRequest: 2,
          cpuMemoryRequest: "4G",
          gpuDevicesRequest: 1,
        },
        storage: {
          pvc: {
            instances: [{ claimName: "pvc-1", path: "/data" }],
          },
          nfs: {
            instances: [{ server: "nfs-server", path: "/exports" }],
          },
        },
      };
      const complexRules = {
        compute: {
          cpuCoreRequest: { min: 1, max: 10 },
          gpuDevicesRequest: { min: 0, max: 8 },
        },
      };
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: complexDefaults,
            worker: complexDefaults,
          },
          rules: {
            leader: complexRules,
            worker: complexRules,
          },
        },
      };
      expect(shouldCopyLeaderToWorker(policy)).toBe(true);
    });

    it("should detect differences in storage configuration", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: {
              storage: {
                pvc: {
                  instances: [
                    { claimName: "pvc-1", path: "/data1" },
                    { claimName: "pvc-2", path: "/data2" },
                  ],
                },
              },
            },
            worker: {
              storage: {
                pvc: {
                  instances: [
                    { claimName: "pvc-2", path: "/data2" },
                    { claimName: "pvc-1", path: "/data1" },
                  ],
                },
              },
            },
          },
        },
      };
      // Arrays with different order should be considered different
      expect(shouldCopyLeaderToWorker(policy)).toBe(false);
    });

    it("should handle empty instances vs undefined", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: {
              annotations: {
                instances: [],
              },
            },
            worker: {
              annotations: undefined,
            },
          },
        },
      };
      expect(shouldCopyLeaderToWorker(policy)).toBe(false);
    });

    it("should detect differences in storage types", () => {
      const policy: UnifiedPolicyInfo = {
        worker: {
          defaults: {
            leader: {
              storage: {
                pvc: {
                  instances: [{ claimName: "pvc-1", path: "/data" }],
                },
              },
            },
            worker: {
              storage: {
                pvc: {
                  instances: [{ claimName: "pvc-different", path: "/different-path" }],
                },
              },
            },
          },
        },
      };
      expect(shouldCopyLeaderToWorker(policy)).toBe(false);
    });
  });
});
