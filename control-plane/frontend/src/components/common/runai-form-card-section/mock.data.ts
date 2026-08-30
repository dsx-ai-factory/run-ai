import { TCardCmpName } from "@/components/common/runai-card-list";
import { GpuRequestType } from "@/swagger-models/assets-service-client";

export const cardsList = [
  {
    meta: {
      name: "cpu-only",
      scope: "tenant",
      id: "1a500b1b-6a39-4737-a8d4-32fce6883a7b",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-04-16T14:20:40.150338Z",
      updatedBy: "runai",
      updatedAt: "2023-04-16T14:20:40.150338Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 0,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    },
  },
  {
    meta: {
      name: "testt",
      scope: "tenant",
      id: "97070476-e86b-44f9-be0b-f3b5c96db2c6",
      kind: "compute",
      tenantId: 1,
      createdBy: "test@run.ai",
      createdAt: "2023-04-17T11:58:40.639365Z",
      updatedBy: "test@run.ai",
      updatedAt: "2023-04-17T11:58:40.639365Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 0,
      cpuCoreRequest: 0,
      cpuCoreLimit: 0,
      cpuMemoryRequest: "10M",
      largeShmRequest: false,
    },
  },
  {
    meta: {
      name: "small-fraction",
      scope: "tenant",
      id: "7d937f29-f557-40dc-ae79-585217a5c9ee",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-04-16T14:20:40.151259Z",
      updatedBy: "runai",
      updatedAt: "2023-04-16T14:20:40.151259Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 0.1,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    },
  },
  {
    meta: {
      name: "half-gpu",
      scope: "tenant",
      id: "442c15b1-1ddd-4825-b309-3d1773fb5759",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-04-16T14:20:40.150798Z",
      updatedBy: "runai",
      updatedAt: "2023-04-16T14:20:40.150798Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 0.5,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    },
  },
  {
    meta: {
      name: "one-gpu",
      scope: "tenant",
      id: "04c3f5eb-bb48-464b-8290-bffc9dc9289f",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-04-16T14:20:40.150077Z",
      updatedBy: "runai",
      updatedAt: "2023-04-16T14:20:40.150077Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 1,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    },
  },
  {
    meta: {
      name: "two-gpus",
      scope: "tenant",
      id: "c239d795-1e75-4c5d-9484-5d35ac695e55",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-04-16T14:20:40.151057Z",
      updatedBy: "runai",
      updatedAt: "2023-04-16T14:20:40.151057Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 2,
      cpuCoreRequest: 0.2,
      cpuMemoryRequest: "200M",
    },
  },
  {
    meta: {
      name: "kjh",
      scope: "tenant",
      id: "5bb14779-4258-47f3-a6e6-ae069ce513bc",
      kind: "compute",
      tenantId: 1,
      createdBy: "test@run.ai",
      createdAt: "2023-04-16T17:18:10.702405Z",
      updatedBy: "test@run.ai",
      updatedAt: "2023-04-16T17:18:10.702405Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 5,
      cpuCoreRequest: 0,
      cpuCoreLimit: 0,
      cpuMemoryRequest: "10M",
      largeShmRequest: false,
    },
  },
  {
    meta: {
      name: "jllkj",
      scope: "tenant",
      id: "0970f0e1-1b18-434d-bc75-962a2f25b444",
      kind: "compute",
      tenantId: 1,
      createdBy: "test@run.ai",
      createdAt: "2023-04-16T17:19:57.810155Z",
      updatedBy: "test@run.ai",
      updatedAt: "2023-04-16T17:19:57.810155Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortion: 5,
      cpuCoreRequest: 0,
      cpuCoreLimit: 0,
      cpuMemoryRequest: "0M",
      cpuMemoryLimit: "0M",
      largeShmRequest: false,
    },
  },
].map((computeResource) => {
  const name = computeResource.meta.name;
  const cpu = (computeResource.spec.cpuCoreRequest || "").toString();
  const memory = computeResource.spec.cpuMemoryRequest || "";
  const gpu = (computeResource.spec.gpuPortion || 0).toString();
  return {
    id: computeResource.meta.id,
    cardName: TCardCmpName.COMPUTE_RESOURCE,
    data: computeResource,
    searchValues: [name, cpu, memory, gpu],
  };
});
