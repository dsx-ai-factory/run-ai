import { yamlUtil } from "./yaml.util";

describe("yaml.util", () => {
  describe("objectToYaml", () => {
    it("should convert a valid object to YAML", () => {
      const inputObject = {
        name: "John Doe",
        age: 30,
        address: {
          city: "New York",
          zip: 90210,
        },
      };

      const yamlString = yamlUtil.objectToYaml(inputObject);
      const expectedYamlString = `name: John Doe
age: 30
address:
  city: New York
  zip: 90210
`;
      expect(yamlString).toEqual(expectedYamlString);
    });

    it("should handle invalid input gracefully", () => {
      // Pass an invalid object (function) to the function
      const inputObject = () => {
        return "Invalid object";
      };

      const yamlString = yamlUtil.objectToYaml(inputObject);

      expect(yamlString).toBeNull();
    });
  });

  describe("validateYaml", () => {
    it("should return null for valid YAML content", () => {
      const validYaml = `
        name: John Doe
        age: 30
        city: New York
      `;

      const validationResult = yamlUtil.validateYaml(validYaml);

      expect(validationResult).toBeNull();
    });

    it("should handle empty input gracefully", () => {
      const emptyYaml = "";

      const validationResult = yamlUtil.validateYaml(emptyYaml);

      // Expect a non-null result (error message)
      expect(validationResult).toBeNull();
    });

    it("should return an error message for invalid YAML content", () => {
      const invalidYaml = `
        name: John Doe
        age:
        New York
      `;

      const validationResult = yamlUtil.validateYaml(invalidYaml);

      // Expect a non-null result (error message)
      expect(validationResult).not.toBeNull();
    });
  });

  describe("extractWorkloadValues", () => {
    it("should extract namespace, priority, and preemptibility from metadata", () => {
      const yaml = `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: runai-project
  priority: high
  preemptibility: preemptible
spec:
  containers: []
`;

      const result = yamlUtil.extractWorkloadValues(yaml);

      expect(result.namespace).toBe("runai-project");
      expect(result.priority).toBe("high");
      expect(result.preemptibility).toBe("preemptible");
    });

    it("should return null for missing metadata fields", () => {
      const yaml = `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers: []
`;

      const result = yamlUtil.extractWorkloadValues(yaml);

      expect(result.namespace).toBeNull();
      expect(result.priority).toBeNull();
      expect(result.preemptibility).toBeNull();
    });

    it("should return defaults for empty YAML", () => {
      const result = yamlUtil.extractWorkloadValues("");

      expect(result.namespace).toBeNull();
      expect(result.priority).toBeNull();
      expect(result.preemptibility).toBeNull();
    });

    it("should return defaults for invalid YAML", () => {
      const invalidYaml = "{ invalid yaml ::";

      const result = yamlUtil.extractWorkloadValues(invalidYaml);

      expect(result.namespace).toBeNull();
      expect(result.priority).toBeNull();
      expect(result.preemptibility).toBeNull();
    });

    it("should extract only namespace when priority and preemptibility are not present", () => {
      const yaml = `
metadata:
  namespace: my-namespace
`;

      const result = yamlUtil.extractWorkloadValues(yaml);

      expect(result.namespace).toBe("my-namespace");
      expect(result.priority).toBeNull();
      expect(result.preemptibility).toBeNull();
    });
  });
});
