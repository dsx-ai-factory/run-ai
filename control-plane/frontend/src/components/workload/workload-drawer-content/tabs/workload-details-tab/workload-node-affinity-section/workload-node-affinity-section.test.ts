import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import WorkloadNodeAffinitySection from "./workload-node-affinity-section.vue";
import type { PodNodeAffinity } from "@/swagger-models/workloads-client";
import { MatchExpressionOperator } from "@/swagger-models/workloads-client";

describe("WorkloadNodeAffinitySection", () => {
  const createWrapper = (nodeAffinity: PodNodeAffinity) => {
    return mount(WorkloadNodeAffinitySection, {
      props: { nodeAffinity },
    });
  };

  describe("required affinity", () => {
    it("should render required section with match expressions", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "topology.kubernetes.io/zone",
                  operator: MatchExpressionOperator.In,
                  values: ["us-east-1a", "us-east-1b"],
                },
              ],
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);

      expect(wrapper.text()).toContain("Required:");
      expect(wrapper.text()).toContain("topology.kubernetes.io/zone");
      expect(wrapper.text()).toContain("in");
      expect(wrapper.text()).toContain("[us-east-1a, us-east-1b]");
    });

    it("should render AND between multiple expressions in a term", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "topology.kubernetes.io/zone",
                  operator: MatchExpressionOperator.In,
                  values: ["us-east-1a"],
                },
                {
                  key: "capacity-type",
                  operator: MatchExpressionOperator.NotIn,
                  values: ["spot"],
                },
              ],
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);

      expect(wrapper.text()).toContain("AND");
      expect(wrapper.text()).toContain("topology.kubernetes.io/zone");
      expect(wrapper.text()).toContain("capacity-type");
      expect(wrapper.text()).toContain("not in");
    });

    it("should render OR separator between multiple terms", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "zone",
                  operator: MatchExpressionOperator.In,
                  values: ["us-east-1a"],
                },
              ],
            },
            {
              matchExpressions: [
                {
                  key: "hostname",
                  operator: MatchExpressionOperator.In,
                  values: ["prod-worker-01"],
                },
              ],
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);

      expect(wrapper.text()).toContain("--- OR ---");
      expect(wrapper.text()).toContain("zone");
      expect(wrapper.text()).toContain("hostname");
    });
  });

  describe("preferred affinity", () => {
    it("should render preferred section with weight and expressions", () => {
      const nodeAffinity: PodNodeAffinity = {
        preferred: {
          nodeSelectorTerms: [
            {
              weight: 80,
              preference: {
                matchExpressions: [
                  {
                    key: "nvidia.com/gpu",
                    operator: MatchExpressionOperator.Exists,
                  },
                ],
              },
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);

      expect(wrapper.text()).toContain("Preferred:");
      expect(wrapper.text()).toContain("Weight 80:");
      expect(wrapper.text()).toContain("nvidia.com/gpu");
      expect(wrapper.text()).toContain("exists");
    });

    it("should not show values for Exists operator", () => {
      const nodeAffinity: PodNodeAffinity = {
        preferred: {
          nodeSelectorTerms: [
            {
              weight: 1,
              preference: {
                matchExpressions: [
                  {
                    key: "some-key",
                    operator: MatchExpressionOperator.Exists,
                    values: [],
                  },
                ],
              },
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);
      const badges = wrapper.findAll(".expression-badge");

      // Only the key badge, no values badge
      expect(badges).toHaveLength(1);
      expect(badges[0].text()).toBe("some-key");
    });

    it("should not show values for DoesNotExist operator", () => {
      const nodeAffinity: PodNodeAffinity = {
        preferred: {
          nodeSelectorTerms: [
            {
              weight: 1,
              preference: {
                matchExpressions: [
                  {
                    key: "some-key",
                    operator: MatchExpressionOperator.DoesNotExist,
                  },
                ],
              },
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);

      expect(wrapper.text()).toContain("does not exist");
      const badges = wrapper.findAll(".expression-badge");
      expect(badges).toHaveLength(1);
    });
  });

  describe("operator formatting", () => {
    const testCases: Array<{ operator: MatchExpressionOperator; expected: string }> = [
      { operator: MatchExpressionOperator.In, expected: "in" },
      { operator: MatchExpressionOperator.NotIn, expected: "not in" },
      { operator: MatchExpressionOperator.Exists, expected: "exists" },
      { operator: MatchExpressionOperator.DoesNotExist, expected: "does not exist" },
      { operator: MatchExpressionOperator.Gt, expected: ">" },
      { operator: MatchExpressionOperator.Lt, expected: "<" },
    ];

    testCases.forEach(({ operator, expected }) => {
      it(`should format ${operator} as "${expected}"`, () => {
        const nodeAffinity: PodNodeAffinity = {
          required: {
            nodeSelectorTerms: [
              {
                matchExpressions: [
                  {
                    key: "test-key",
                    operator,
                    values:
                      operator === MatchExpressionOperator.Exists || operator === MatchExpressionOperator.DoesNotExist
                        ? undefined
                        : ["val"],
                  },
                ],
              },
            ],
          },
        };

        const wrapper = createWrapper(nodeAffinity);
        expect(wrapper.text()).toContain(expected);
      });
    });
  });

  describe("value formatting", () => {
    it("should format single value without brackets", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "key",
                  operator: MatchExpressionOperator.In,
                  values: ["single-value"],
                },
              ],
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);
      expect(wrapper.text()).toContain("single-value");
      expect(wrapper.text()).not.toContain("[single-value]");
    });

    it("should format multiple values with brackets", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "key",
                  operator: MatchExpressionOperator.In,
                  values: ["val1", "val2"],
                },
              ],
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);
      expect(wrapper.text()).toContain("[val1, val2]");
    });
  });

  describe("empty states", () => {
    it("should not render required section when not present", () => {
      const nodeAffinity: PodNodeAffinity = {
        preferred: {
          nodeSelectorTerms: [
            {
              weight: 1,
              preference: {
                matchExpressions: [
                  {
                    key: "key",
                    operator: MatchExpressionOperator.Exists,
                  },
                ],
              },
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);
      expect(wrapper.text()).not.toContain("Required:");
      expect(wrapper.text()).toContain("Preferred:");
    });

    it("should not render preferred section when not present", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "key",
                  operator: MatchExpressionOperator.In,
                  values: ["val"],
                },
              ],
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);
      expect(wrapper.text()).toContain("Required:");
      expect(wrapper.text()).not.toContain("Preferred:");
    });

    it("should render both sections when both present", () => {
      const nodeAffinity: PodNodeAffinity = {
        required: {
          nodeSelectorTerms: [
            {
              matchExpressions: [
                {
                  key: "req-key",
                  operator: MatchExpressionOperator.In,
                  values: ["val"],
                },
              ],
            },
          ],
        },
        preferred: {
          nodeSelectorTerms: [
            {
              weight: 50,
              preference: {
                matchExpressions: [
                  {
                    key: "pref-key",
                    operator: MatchExpressionOperator.Exists,
                  },
                ],
              },
            },
          ],
        },
      };

      const wrapper = createWrapper(nodeAffinity);
      expect(wrapper.text()).toContain("Required:");
      expect(wrapper.text()).toContain("Preferred:");
      expect(wrapper.text()).toContain("req-key");
      expect(wrapper.text()).toContain("pref-key");
    });
  });
});
