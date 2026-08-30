import type { WorkloadTemplate } from "@/swagger-models/assets-service-client";

export const templatesMockData: WorkloadTemplate[] = [
  {
    meta: {
      name: "test-template",
      scope: "project",
      clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
      projectId: 1,
      id: "d3481435-996e-4ca2-9e52-00f8d8203725",
      kind: "workload-template",
      tenantId: 1,
      createdBy: "test@run.ai",
      createdAt: "2023-03-15T15:29:13.340953Z",
      updatedBy: "test@run.ai",
      updatedAt: "2023-03-15T15:29:13.340953Z",
      projectName: "test-1",
    },
    spec: {
      assets: {
        environment: {
          id: "3c6fc0e5-effa-4a55-85f9-5adb6278ce50",
          name: "test-1-1-1",
          toolTypes: ["jupyter-notebook", "mlflow", "rstudio"],
        },
        compute: {
          id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
          name: "cpu-only",
        },
      },
      specificEnv: {
        command: "test",
        args: "arg1",
        environmentVariables: [
          {
            name: "env-var",
            value: "12",
          },
        ],
        connections: [],
        runAsUid: 5,
        runAsGid: 3,
        supplementalGroups: "5,5,5",
        allowOverQuota: false,
      },
    },
  },
];
