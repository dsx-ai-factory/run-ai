import yaml from "js-yaml";
import { YAMLException } from "js-yaml";

export interface IYamlWorkloadValues {
  namespace: string | null;
  priority: string | null;
  preemptibility: string | null;
}

export interface IYamlDryRunFields {
  name: string | null;
  apiVersion: string | null;
  kind: string | null;
  namespace: string | null;
}

export const yamlUtil = {
  objectToYaml,
  validateYaml,
  yamlToObject,
  extractWorkloadName,
  extractWorkloadKind,
  extractWorkloadValues,
  extractGpuRequest,
  extractDryRunFields,
};

function objectToYaml(data: object): string | null {
  try {
    return yaml.dump(data);
  } catch (error) {
    return null;
  }
}

function yamlToObject(data: string): object | null {
  if (!data) return {};
  try {
    return yaml.load(data) as object;
  } catch (error) {
    return null;
  }
}

function validateYaml(yamlContent: string): string | null {
  try {
    yaml.load(yamlContent);
    return null; // YAML is valid
  } catch (error) {
    // YAML is not valid
    if (error instanceof YAMLException) {
      return error.message;
    } else {
      return String(error);
    }
  }
}

/**
 * Extracts the workload name from YAML manifest metadata
 */
function extractWorkloadName(yamlContent: string): string | null {
  const parsed = yamlToObject(yamlContent);
  if (!parsed || typeof parsed !== "object") return null;

  const manifest = parsed as Record<string, unknown>;
  const metadata = manifest.metadata as Record<string, unknown> | undefined;

  if (metadata && typeof metadata.name === "string") {
    return metadata.name;
  }
  return null;
}

/**
 * Extracts the workload kind from YAML manifest
 */
function extractWorkloadKind(yamlContent: string): string | null {
  const parsed = yamlToObject(yamlContent);
  if (!parsed || typeof parsed !== "object") return null;

  const manifest = parsed as Record<string, unknown>;
  if (typeof manifest.kind === "string") {
    return manifest.kind;
  }
  return null;
}

/**
 * Extracts workload values from YAML metadata
 * Extracts namespace, priority, and preemptibility from the manifest's metadata section
 */
function extractWorkloadValues(yamlContent: string): IYamlWorkloadValues {
  const defaultValues: IYamlWorkloadValues = { namespace: null, priority: null, preemptibility: null };

  if (!yamlContent) return defaultValues;

  const parsed = yamlToObject(yamlContent);
  if (!parsed || typeof parsed !== "object") return defaultValues;

  const manifest = parsed as Record<string, unknown>;

  // Extract all values from metadata
  const metadata = manifest.metadata as Record<string, unknown> | undefined;
  const namespace = (metadata?.namespace as string) || null;
  const priority = (metadata?.priority as string) || null;
  const preemptibility = (metadata?.preemptibility as string) || null;

  return { namespace, priority, preemptibility };
}

/**
 * Extracts GPU request from YAML spec
 * Handles different formats: gpu: "1" or gpu: { value: "1" }
 */
function extractGpuRequest(yamlContent: string): number {
  if (!yamlContent) return 0;

  const parsed = yamlToObject(yamlContent);
  if (!parsed || typeof parsed !== "object") return 0;

  const spec = (parsed as Record<string, unknown>).spec as Record<string, unknown> | undefined;
  if (!spec) return 0;

  const gpu = spec.gpu;
  if (typeof gpu === "string" || typeof gpu === "number") {
    return Number(gpu) || 0;
  }
  if (typeof gpu === "object" && gpu !== null && "value" in gpu) {
    return Number((gpu as Record<string, unknown>).value) || 0;
  }
  return 0;
}

/**
 * Extracts fields used for dry-run validation (name, apiVersion, kind, namespace)
 */
function extractDryRunFields(yamlContent: string): IYamlDryRunFields {
  const defaultFields: IYamlDryRunFields = { name: null, apiVersion: null, kind: null, namespace: null };

  if (!yamlContent) return defaultFields;

  const parsed = yamlToObject(yamlContent);
  if (!parsed || typeof parsed !== "object") return defaultFields;

  const manifest = parsed as Record<string, unknown>;

  const apiVersion = typeof manifest.apiVersion === "string" ? manifest.apiVersion : null;
  const kind = typeof manifest.kind === "string" ? manifest.kind : null;

  const metadata = manifest.metadata as Record<string, unknown> | undefined;
  const name = (metadata?.name as string) || null;
  const namespace = (metadata?.namespace as string) || null;

  return { name, apiVersion, kind, namespace };
}
