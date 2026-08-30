export interface INodeGpu {
  index: number;
  allocated: number;
  allocatedMemory: number;
  usedMemory: number;
  totalMemory: number;
  utilization: number;
  memoryUtilization: number;
  idleTime: number;
  graphicsEngineActivity?: number;
  smActivity?: number;
  smOccupancy?: number;
  tensorActivity?: number;
  fp64EngineActivity?: number;
  fp32EngineActivity?: number;
  fp16EngineActivity?: number;
  memoryBWUtilization?: number;
  nvLinkBandwidthRead?: number;
  nvLinkBandwidthTransmit?: number;
  pcieBandwidthRead?: number;
  pcieBandwidthTransmit?: number;
}
// For more details about the above metrics names see the following link: https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/feature-overview.html#metrics
