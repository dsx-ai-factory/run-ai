import { type ExtendedResource, GpuRequestType } from "@/swagger-models/assets-service-client";
import { EMemoryUnitValue, type ISelectOption, type IUIAssetCreationMeta } from "./global.model";

export const cpuUnitOptions: Array<ISelectOption> = [
  { label: "Cores", value: 1 },
  { label: "Millicores", value: 1000 },
];

export const gpuRequestTypeOptions = [
  { label: "% of device", value: GpuRequestType.Portion },
  { label: "MB", value: EMemoryUnitValue.MB },
  { label: "GB", value: EMemoryUnitValue.GB },
];

export interface IUIComputeResourceCreation {
  meta: IUIAssetCreationMeta;
  spec: {
    gpuDevicesRequest: number;
    gpuRequestType?: GpuRequestType | null;
    gpuPortionRequest?: number | null;
    gpuPortionLimit?: number | null;
    gpuMemoryRequest?: string | null;
    gpuMemoryLimit?: string | null;
    cpuCoreRequest?: number | null;
    cpuMemoryRequest?: string | null;
    cpuCoreLimit?: number | null;
    cpuMemoryLimit?: string | null;
    largeShmRequest?: boolean | null;
    extendedResources?: Array<ExtendedResource> | null;
  };
}

export interface ICpuCore {
  request?: number | null;
  limit?: number | null;
}

export interface ICpuMemory {
  request?: string | null;
  limit?: string | null;
}

export interface IGpuPortion {
  request: number;
  limit?: number | null | undefined;
}

export interface IGpuMemory {
  request: string;
  limit?: string | null | undefined;
}
export interface IResourcesSectionModel {
  gpuDevicesRequest: number;
  gpuRequestType?: GpuRequestType | null;
  gpuPortionRequest?: number | null;
  gpuPortionLimit?: number | null;
  gpuMemoryRequest?: string | null;
  gpuMemoryLimit?: string | null;
  cpuCore: ICpuCore;
  cpuMemory: ICpuMemory;
  minWorkers?: number | null;
  maxWorkers?: number | null;
}

export interface IGpuRequestModel {
  gpuDevicesRequest: number;
  gpuRequestType?: GpuRequestType | null;
  gpuPortionRequest?: number | null;
  gpuMemoryRequest?: string | null;
  gpuPortionLimit?: number | null;
  gpuMemoryLimit?: string | null;
}

export interface ICpuRequestModel {
  cpuCore: ICpuCore;
  cpuMemory: ICpuMemory;
}
