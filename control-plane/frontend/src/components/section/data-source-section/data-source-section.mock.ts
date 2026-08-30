import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";

const dataSources: DatasourceListResponseEntry[] = [
  {
    meta: {
      name: "host-path",
      scope: "cluster",
      clusterId: "c56f038f-df9a-47a1-83f6-8dd4a4530909",
      id: "cce99c0f-70e3-48bd-ac42-89e9017d7b67",
      kind: "hostPath",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-07-18T13:10:37.87588Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-07-18T13:10:37.87588Z",
    },
    spec: {
      hostPath: {
        path: "/usr",
        readOnly: false,
        mountPath: "/home",
      },
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "nfs-100",
      scope: "cluster",
      clusterId: "c56f038f-df9a-47a1-83f6-8dd4a4530909",
      id: "6e082428-a0c2-4748-b240-934bd484207d",
      kind: "nfs",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-07-18T13:10:16.23967Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-07-18T13:10:16.23967Z",
    },
    spec: {
      nfs: {
        path: "/usr",
        readOnly: false,
        server: "myserver",
        mountPath: "/home",
      },
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "git",
      scope: "cluster",
      clusterId: "c56f038f-df9a-47a1-83f6-8dd4a4530909",
      id: "c845ec5d-7d96-4f03-82c1-1ff3563fb8d7",
      kind: "git",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-07-18T13:11:19.31101Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-07-18T13:11:19.31101Z",
    },
    spec: {
      git: {
        repository: "https://git.com",
        revision: "master",
        path: "/home",
      },
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "s3",
      scope: "cluster",
      clusterId: "c56f038f-df9a-47a1-83f6-8dd4a4530909",
      id: "eb057439-9202-4cde-9e12-00034c42ee45",
      kind: "s3",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-07-18T13:12:06.677876Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-07-18T13:12:06.677876Z",
    },
    spec: {
      s3: {
        bucket: "my-bucket",
        path: "/home",
        url: "https://s3.com",
      },
    },
    compliance: {
      compliance: true,
    },
  },
];

export const mockDataSourceSectionData = {
  dataSources,
};
