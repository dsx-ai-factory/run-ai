import { GpuRequestType } from "@/swagger-models/assets-service-client";
import { type ICardListItem, TCardCmpName } from "./runai-card-list.model";

const computeResourcesMockedData = {
  list: [
    {
      id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "cpu-only",
          scope: "tenant",
          id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967182Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967182Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortionRequest: 0,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "5ac6fc47-45d8-4b06-a82d-7c5324b49851",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "small-fraction",
          scope: "tenant",
          id: "5ac6fc47-45d8-4b06-a82d-7c5324b49851",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967989Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967989Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 0.1,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "751115c8-235f-4b35-af99-3d367a51baaf",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "half-gpu",
          scope: "tenant",
          id: "751115c8-235f-4b35-af99-3d367a51baaf",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967604Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967604Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 0.5,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "1d47ae37-0ae2-4d13-a879-716472108ed6",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "one-gpu",
          scope: "tenant",
          id: "1d47ae37-0ae2-4d13-a879-716472108ed6",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.966949Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.966949Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 1,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "two-gpus",
          scope: "tenant",
          id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967809Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967809Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 2,
          cpuCoreRequest: 0.2,
          cpuMemoryRequest: "200M",
        },
      },
    },
  ] as ICardListItem[],
  largeList: [
    {
      id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "cpu-only",
          scope: "tenant",
          id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967182Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967182Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 0,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "5ac6fc47-45d8-4b06-a82d-7c5324b49851",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "small-fraction",
          scope: "tenant",
          id: "5ac6fc47-45d8-4b06-a82d-7c5324b49851",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967989Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967989Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 0.1,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "751115c8-235f-4b35-af99-3d367a51baaf",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "half-gpu",
          scope: "tenant",
          id: "751115c8-235f-4b35-af99-3d367a51baaf",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967604Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967604Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 0.5,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "1d47ae37-0ae2-4d13-a879-716472108ed6",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "one-gpu",
          scope: "tenant",
          id: "1d47ae37-0ae2-4d13-a879-716472108ed6",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.966949Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.966949Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 1,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "two-gpus",
          scope: "tenant",
          id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967809Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967809Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 2,
          cpuCoreRequest: 0.2,
          cpuMemoryRequest: "200M",
        },
      },
    },
    {
      id: "123-9a2d-40db-a123e3-123",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "test-1-gpus",
          scope: "tenant",
          id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967809Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967809Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 1,
          cpuCoreRequest: 0.1,
          cpuMemoryRequest: "100M",
        },
      },
    },
    {
      id: "546-9a2d-456-ade3-546",
      cardName: TCardCmpName.COMPUTE_RESOURCE,
      data: {
        meta: {
          name: "test-2-gpus",
          scope: "tenant",
          id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
          kind: "compute",
          tenantId: 1,
          createdBy: "runai",
          createdAt: "2023-03-07T08:34:32.967809Z",
          updatedBy: "runai",
          updatedAt: "2023-03-07T08:34:32.967809Z",
        },
        spec: {
          gpuRequestType: GpuRequestType.Portion,
          gpuPortion: 3,
          cpuCoreRequest: 0.3,
          cpuMemoryRequest: "300M",
        },
      },
    },
  ],
};

export { computeResourcesMockedData };
