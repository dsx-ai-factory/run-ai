import type { WorkloadSpec } from "@/utils/workload-details.util";
import { EWorkloadType } from "@/models/workload.model";

type ManifestObject = Record<string, unknown>;
type WorkloadV2CrudLike = {
  spec?: WorkloadSpec;
  masterSpec?: WorkloadSpec;
};

const DISTRIBUTED_REPLICA_SPEC_KEYS = [
  "pytorchReplicaSpecs",
  "tfReplicaSpecs",
  "xgbReplicaSpecs",
  "jaxReplicaSpecs",
  "mpiReplicaSpecs",
];

const DISTRIBUTED_REPLICA_TYPE_KEYS = ["Worker", "Launcher", "Chief", "Master", "PS"];

export const workloadV2ManifestUtil = {
  getCrudWorkloadFromManifest,
  getSpecFromManifest,
  getCommandAndArgumentsFromManifest,
};

function getCrudWorkloadFromManifest(
  manifest?: ManifestObject | null,
  workloadType?: string,
): WorkloadV2CrudLike | undefined {
  const spec = mapValue(manifest, "spec");
  if (!spec) return undefined;

  if (workloadType === EWorkloadType.Distributed) {
    return getDistributedCrudWorkloadFromSpec(spec);
  }

  const workloadSpec = getSpecFromManifest(manifest);
  if (!workloadSpec) return undefined;

  return { spec: workloadSpec };
}

function getSpecFromManifest(manifest?: ManifestObject | null): WorkloadSpec | undefined {
  const spec = mapValue(manifest, "spec");
  if (!spec) return undefined;

  const template = mapValue(spec, "template");
  if (template) {
    const workloadSpec = getSpecFromPodTemplate(template);
    if (workloadSpec) return workloadSpec;
  }

  return getSpecFromReplicaSpecs(spec);
}

function getCommandAndArgumentsFromManifest(manifest?: ManifestObject | null): { command: string; arguments: string } {
  const workloadSpec = getSpecFromManifest(manifest);
  return { command: workloadSpec?.command ?? "", arguments: workloadSpec?.args ?? "" };
}

function getDistributedCrudWorkloadFromSpec(spec: ManifestObject): WorkloadV2CrudLike | undefined {
  for (const replicaSpecKey of DISTRIBUTED_REPLICA_SPEC_KEYS) {
    const replicaSpecs = mapValue(spec, replicaSpecKey);
    if (!replicaSpecs) continue;

    const masterSpec = getSpecFromReplicaSpec(replicaSpecs, "Master");
    const workerSpec = getSpecFromReplicaSpec(replicaSpecs, "Worker");
    if (masterSpec || workerSpec) {
      return { masterSpec, spec: workerSpec };
    }
  }

  return undefined;
}

function getSpecFromReplicaSpecs(spec: ManifestObject): WorkloadSpec | undefined {
  for (const replicaSpecKey of DISTRIBUTED_REPLICA_SPEC_KEYS) {
    const replicaSpecs = mapValue(spec, replicaSpecKey);
    if (!replicaSpecs) continue;

    for (const replicaTypeKey of DISTRIBUTED_REPLICA_TYPE_KEYS) {
      const workloadSpec = getSpecFromReplicaSpec(replicaSpecs, replicaTypeKey);
      if (workloadSpec) return workloadSpec;
    }

    for (const replicaTypeKey of Object.keys(replicaSpecs).sort()) {
      const workloadSpec = getSpecFromReplicaSpec(replicaSpecs, replicaTypeKey);
      if (workloadSpec) return workloadSpec;
    }
  }

  return undefined;
}

function getSpecFromReplicaSpec(replicaSpecs: ManifestObject, replicaTypeKey: string): WorkloadSpec | undefined {
  const replicaSpec = mapValue(replicaSpecs, replicaTypeKey);
  const template = mapValue(replicaSpec, "template");
  if (!template) return undefined;
  return getSpecFromPodTemplate(template);
}

function getSpecFromPodTemplate(template: ManifestObject): WorkloadSpec | undefined {
  const spec = mapValue(template, "spec");
  const containers = arrayValue(spec?.containers);
  const container = mapFromUnknown(containers?.[0]);
  if (!container) return undefined;

  const image = stringValue(container.image);
  const command = stringArrayValue(container.command).join(" ");
  const args = stringArrayValue(container.args).join(" ");
  if (!image && !command && !args) return undefined;

  return {
    ...(image ? { image } : {}),
    ...(command ? { command } : {}),
    ...(args ? { args } : {}),
  } as WorkloadSpec;
}

function mapValue(source: ManifestObject | null | undefined, key: string): ManifestObject | undefined {
  if (!source) return undefined;
  return mapFromUnknown(Object.getOwnPropertyDescriptor(source, key)?.value);
}

function mapFromUnknown(value: unknown): ManifestObject | undefined {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as ManifestObject;
  }
  return undefined;
}

function arrayValue(value: unknown): unknown[] | undefined {
  return Array.isArray(value) ? value : undefined;
}

function stringArrayValue(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}
