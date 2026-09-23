/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { computeResourceIndexColumns } from "@/table-models/compute-resource.table-model";
import { dataSourceIndexColumns } from "@/table-models/data-source.table-model";
import { environmentIndexColumns } from "@/table-models/environment.table-model";
import { credentialIndexColumns } from "@/table-models/credential.table-model";
import { workloadTemplateIndexColumns } from "@/table-models/workload-template.table-model";

import {
  GpuRequestType,
  type ComputeAsset,
  type CredentialsListResponseEntry,
  type DatasourceListResponseEntry,
  type EnvironmentAsset,
  type WorkloadTemplate,
} from "@/swagger-models/assets-service-client";
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";

function getDisplayedColumns(columns: Array<ITableColumn>): Array<string> {
  return columns.filter((column) => column.display && column.name !== "scope").map((column) => column.name);
}

const environmentColumns = environmentIndexColumns.map((c) => {
  if (c.name === "workloads") return { ...c, display: true };
  return c;
});
const computeColumns = computeResourceIndexColumns.map((c) => {
  if (c.name === "workloads") return { ...c, display: true };
  return c;
});
const dataSourceColumns = dataSourceIndexColumns.map((c) => {
  if (c.name === "workloads") return { ...c, display: true };
  return c;
});

const environmentModel = {
  rows: [
    {
      meta: {
        name: "comet-env",
        scope: "tenant",
        id: "ca182e3b-4d8e-482a-afcb-0ee165a0fa15",
        kind: "environment",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-04-04T07:19:22.818056Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-04-04T07:19:22.818056Z",
      },
      spec: {
        environmentVariables: [],
        image: "ubuntu",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "Comet",
            isExternal: true,
            externalToolInfo: {
              toolType: "comet",
              externalUrl: "https://www.com.comet/comet",
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
      usedBy: {
        workspaces: [
          {
            id: "73180787-cb59-4306-a138-dd0334b85260",
            name: "comet-ws",
            status: "Succeeded",
          },
        ],
        assets: {},
      },
    },
    {
      meta: {
        name: "env-policy-test",
        scope: "project",
        clusterId: "1cac0b32-1989-4efe-88b2-ed6ad11ebcc5",
        projectId: 1,
        id: "c1a5d1b2-2dce-44f1-8d69-d0e0bef59a90",
        kind: "environment",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-04-04T05:57:32.20271Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-04-04T05:57:32.20271Z",
        projectName: "project-1",
      },
      spec: {
        environmentVariables: [
          {
            name: "WANDB_PROJECT",
            value: "",
            deleted: true,
          },
        ],
        image: "ubuntu",
        imagePullPolicy: "IfNotPresent",
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
      usedBy: {
        workspaces: [
          {
            id: "166a674b-ecac-45f6-a30b-648ba092a227",
            name: "whatwhat",
            status: "Succeeded",
          },
        ],
        assets: {},
      },
    },
    {
      meta: {
        name: "jupyter-lab",
        scope: "tenant",
        id: "457de34a-8729-4230-b3c6-9b34ded8e11f",
        kind: "environment",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-04-02T12:50:49.03551Z",
        updatedBy: "runai",
        updatedAt: "2023-04-02T12:50:49.03551Z",
      },
      spec: {
        command: "start-notebook.sh",
        args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
        environmentVariables: [],
        image: "jupyter/scipy-notebook",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "Jupyter",
            internalToolInfo: {
              toolType: "jupyter-notebook",
              connectionType: "ExternalUrl",
              containerPort: 8888,
              externalUrlInfo: {},
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
    },
    {
      meta: {
        name: "jupyter-tensorboard",
        scope: "tenant",
        id: "ed761ddd-2172-4f66-9292-eecf203b9c01",
        kind: "environment",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-04-02T12:50:49.036604Z",
        updatedBy: "runai",
        updatedAt: "2023-04-02T12:50:49.036605Z",
      },
      spec: {
        environmentVariables: [
          {
            name: "LOG_DIR",
            value: "./",
          },
        ],
        image: "quay.io/jupyter/scipy-notebook",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "Jupyter",
            internalToolInfo: {
              toolType: "jupyter-notebook",
              connectionType: "ExternalUrl",
              containerPort: 8888,
              externalUrlInfo: {},
            },
          },
          {
            name: "TensorBoard",
            internalToolInfo: {
              toolType: "tensorboard",
              connectionType: "ExternalUrl",
              containerPort: 6006,
              externalUrlInfo: {},
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
    },
    {
      meta: {
        name: "metrics",
        scope: "tenant",
        id: "f03cc847-4ca6-41bf-8766-02de1505372b",
        kind: "environment",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-04-02T12:51:21.441283Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-04-02T12:51:21.441283Z",
      },
      spec: {
        environmentVariables: [],
        image: "ubuntu",
        imagePullPolicy: "IfNotPresent",
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
      usedBy: {
        workspaces: [
          {
            id: "6ca2cf67-8c44-4834-855b-e5eaf58bb2d5",
            name: "test-charts",
            status: "Succeeded",
          },
          {
            id: "8f1e95c2-fd03-4a35-b8a7-535a08031f09",
            name: "test",
            status: "Succeeded",
          },
          {
            id: "09eaf2f0-2fbf-4e0c-8284-911b5128c820",
            name: "zzz",
            status: "Succeeded",
          },
          {
            id: "f5443b15-db27-490a-9ace-f89a65adcaf3",
            name: "wsss",
            status: "Succeeded",
          },
          {
            id: "31863e0d-9e54-4f8a-aa48-0cd5d3abb535",
            name: "with-gpu",
            status: "UnexpectedAdmissionError",
          },
          {
            id: "a810f5fc-915d-4246-a3da-21610b40b86f",
            name: "cha",
            status: "Succeeded",
          },
        ],
        assets: {},
      },
    },
    {
      meta: {
        name: "tensorboard",
        scope: "tenant",
        id: "ccd343b6-58fe-44dc-be33-eaae9dc296cd",
        kind: "environment",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-04-02T12:50:49.033149Z",
        updatedBy: "runai",
        updatedAt: "2023-04-02T12:50:49.033149Z",
      },
      spec: {
        command: "tensorboard",
        args: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
        environmentVariables: [],
        image: "tensorflow/tensorflow:latest",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "TensorBoard",
            internalToolInfo: {
              toolType: "tensorboard",
              connectionType: "ExternalUrl",
              containerPort: 6006,
              externalUrlInfo: {},
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
    },
  ] as Array<EnvironmentAsset>,
  filterBy: {
    sortBy: "environment",
    displayedColumns: getDisplayedColumns(environmentColumns),
    rowsPerPage: 10,
  } as IFilterBy,
};

const computeResourceModel = {
  rows: [
    {
      meta: {
        name: "cpu-only",
        scope: "tenant",
        id: "03690406-d1c8-436d-8861-fdf6fb65a1fb",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-06-04T08:11:02.704793Z",
        updatedBy: "runai",
        updatedAt: "2023-06-04T08:11:02.704793Z",
      },
      spec: {
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
    },
    {
      meta: {
        name: "half-gpu",
        scope: "tenant",
        id: "14770594-8744-4e60-b9b8-da828e21c3da",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-06-04T08:11:02.705402Z",
        updatedBy: "runai",
        updatedAt: "2023-06-04T08:11:02.705402Z",
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
        id: "0da76f69-d7ba-4c10-b8f8-80c3dc52a04e",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-06-04T08:11:02.704504Z",
        updatedBy: "runai",
        updatedAt: "2023-06-04T08:11:02.704504Z",
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
        name: "small-fraction",
        scope: "tenant",
        id: "ef6e915f-5216-415f-b67a-051382d1e603",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-06-04T08:11:02.705963Z",
        updatedBy: "runai",
        updatedAt: "2023-06-04T08:11:02.705963Z",
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
        name: "two-gpus",
        scope: "tenant",
        id: "967b7f1c-d4ed-493e-bca1-f1e7afa75ff7",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-06-04T08:11:02.705796Z",
        updatedBy: "runai",
        updatedAt: "2023-06-04T08:11:02.705796Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Memory,
        gpuMemory: "2M",
        cpuCoreRequest: 0.2,
        cpuMemoryRequest: "200M",
      },
    },
  ] as Array<ComputeAsset>,
  filterBy: {
    sortBy: "compute-resource",
    displayedColumns: getDisplayedColumns(computeColumns),
    rowsPerPage: 10,
  } as IFilterBy,
};

const dataSourceModel = {
  rows: [
    {
      meta: {
        name: "git",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "cdcbd072-3a80-420b-844b-18199610e57c",
        kind: "git",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T07:40:18.330024Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T07:40:18.330024Z",
        projectName: "project-1",
      },
      spec: {
        git: {
          repository: "https://www.com",
          branch: "main",
          revision: "adassa2342sdas32423",
          path: "/home",
        },
      },
    },
    {
      meta: {
        name: "host-path",
        scope: "tenant",
        id: "f82a8034-7e86-4416-aad5-7b2c7fa235ad",
        kind: "hostPath",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T07:40:38.46479Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T07:40:38.46479Z",
      },
      spec: {
        hostPath: {
          path: "/usr",
          readOnly: false,
          mountPath: "/home",
        },
      },
      usedBy: {
        workspaces: [
          {
            id: "f79c6ce7-9e4d-431c-b02e-f0424851f565",
            name: "wsws",
            status: "Running",
          },
          {
            id: "9396e050-6b58-4a1b-883b-12712ae90b6c",
            name: "wswsws",
            status: "Pending",
          },
        ],
        assets: {},
      },
    },
    {
      meta: {
        name: "nfs",
        scope: "department",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        departmentId: "1",
        id: "7ebfa7a8-292d-4eda-adc7-e0e923a37fdc",
        kind: "nfs",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T07:39:03.488922Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T07:39:03.488922Z",
      },
      spec: {
        nfs: {
          path: "/usr",
          readOnly: false,
          server: "myServer",
          mountPath: "/home/local",
        },
      },
      usedBy: {
        workspaces: [
          {
            id: "a23f447d-d5d7-4002-8cd0-64981b739dcc",
            name: "ws-3",
            status: "ContainerCreating",
          },
        ],
        assets: {},
      },
    },
    {
      meta: {
        name: "pvc",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "83e06652-74f0-4bf7-ac62-4bb734de5f55",
        kind: "pvc",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T07:39:30.729546Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T07:39:30.729546Z",
        projectName: "project-1",
      },
      spec: {
        pvc: {
          existingPvc: false,
          claimName: "new-pvc",
          path: "/home/local",
          readOnly: false,
          claimInfo: {
            size: "4G",
            storageClass: "standard",
            accessModes: {
              readWriteOnce: true,
            },
          },
        },
      },
    },
    {
      meta: {
        name: "s3",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "f03e2d09-11d1-460c-b71e-b6d0fe8d4594",
        kind: "s3",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T07:39:53.252819Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T07:39:53.252819Z",
        projectName: "project-1",
      },
      spec: {
        s3: {
          bucket: "my-bucket",
          path: "/home/data",
          url: "https://www.com",
        },
      },
    },
  ] as Array<DatasourceListResponseEntry>,
  filterBy: {
    sortBy: "data-source",
    displayedColumns: getDisplayedColumns(dataSourceColumns),
    rowsPerPage: 10,
  } as IFilterBy,
};

const credentialModel = {
  rows: [
    {
      meta: {
        name: "access-key",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "090faf85-ab3a-41e8-b76b-b62b23408736",
        kind: "accessKey",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T08:25:01.125198Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T08:25:01.125198Z",
        projectName: "project-1",
      },
      spec: {
        accessKey: {
          accessKeyId: "access-key",
          secretAccessKey: "********",
          keyId: 0,
        },
      },
      usedBy: {
        assets: {
          datasources: [
            {
              id: "c0b37553-d388-4185-ab94-d02721faee6f",
              name: "s3-with-cred",
              kind: "s3",
            },
          ],
        },
      },
    },
    {
      meta: {
        name: "docker-registry",
        scope: "tenant",
        id: "26bbaf63-d2c0-4c2b-be97-47df5c2e699d",
        kind: "dockerRegistry",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T08:24:31.58213Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T08:24:31.58213Z",
      },
      spec: {
        dockerRegistry: {
          user: "user",
          password: "********",
          url: "https://www.com",
          keyId: 0,
        },
      },
    },
    {
      meta: {
        name: "username-password",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "5d4d671f-f5ae-4985-91e6-fe5b5d44dbd2",
        kind: "password",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T08:25:24.642178Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T08:25:24.642178Z",
        projectName: "project-1",
      },
      spec: {
        password: {
          user: "username",
          password: "********",
          keyId: 0,
        },
      },
    },
  ] as Array<CredentialsListResponseEntry>,
  filterBy: {
    sortBy: "data-source",
    displayedColumns: getDisplayedColumns(credentialIndexColumns),
    rowsPerPage: 10,
  } as IFilterBy,
};

const templateModel = {
  rows: [
    {
      meta: {
        name: "temp-1",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "c25d764e-dcc3-4266-a320-c392e71f6323",
        kind: "workload-template",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T13:47:25.782095Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T13:47:25.782095Z",
        projectName: "project-1",
      },
      spec: {
        assets: {
          environment: {
            id: "f296d002-a943-4a81-bc15-b1678dcbd75d",
            name: "tensorboard",
            toolTypes: ["tensorboard"],
          },
          compute: {
            id: "03690406-d1c8-436d-8861-fdf6fb65a1fb",
            name: "cpu-only",
          },
          datasources: [
            {
              id: "83e06652-74f0-4bf7-ac62-4bb734de5f55",
              name: "pvc",
              kind: "pvc",
            },
            {
              id: "c0b37553-d388-4185-ab94-d02721faee6f",
              name: "s3-with-cred",
              kind: "s3",
            },
            {
              id: "cdcbd072-3a80-420b-844b-18199610e57c",
              name: "git",
              kind: "git",
            },
          ],
        },
        specificEnv: {
          command: "tensorboard",
          args: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
          environmentVariables: [],
          connections: [],
          allowOverQuota: false,
          nodePools: [],
        },
      },
    },
    {
      meta: {
        name: "temp-2",
        scope: "project",
        clusterId: "37e0838e-2161-45ab-ab63-259d82392fc5",
        projectId: 1,
        id: "fcfd942e-a1d1-478a-afe0-d92b53781b57",
        kind: "workload-template",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-06-05T13:47:39.799499Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-06-05T13:47:39.799499Z",
        projectName: "project-1",
      },
      spec: {
        assets: {
          environment: {
            id: "59a56cf3-9af7-42a9-ad91-be99693a6dd3",
            name: "jupyter-lab",
            toolTypes: ["jupyter-notebook"],
          },
          compute: {
            id: "967b7f1c-d4ed-493e-bca1-f1e7afa75ff7",
            name: "two-gpus",
          },
        },
        specificEnv: {
          command: "start-notebook.sh",
          args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
          environmentVariables: [],
          connections: [],
          allowOverQuota: false,
          nodePools: [],
        },
      },
    },
  ] as Array<WorkloadTemplate>,
  filterBy: {
    sortBy: "template",
    displayedColumns: getDisplayedColumns(workloadTemplateIndexColumns),
    rowsPerPage: 10,
  } as IFilterBy,
};

export const tableModels = {
  environmentModel,
  computeResourceModel,
  dataSourceModel,
  credentialModel,
  templateModel,
  environmentColumns,
  computeColumns,
  dataSourceColumns,
};
