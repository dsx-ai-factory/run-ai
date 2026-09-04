/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  DistributedPolicyDefaultsAndRulesV2,
  WorkspacePolicyDefaultsAndRulesV2,
} from "@/swagger-models/policy-service-client";
import type { UnifiedPolicyInfo } from "@/swagger-models/workloads-client";

const simplePolicy: WorkspacePolicyDefaultsAndRulesV2 = {
  rules: {
    security: {
      allowPrivilegeEscalation: {
        required: true,
        canEdit: true,
        sourceOfRule: {
          scope: "project",
          projectId: 3,
        },
      },
    },
  },
  defaults: {
    security: {
      allowPrivilegeEscalation: true,
    },
  },
};

const policyWithNoRules: WorkspacePolicyDefaultsAndRulesV2 = {
  defaults: {
    security: {
      allowPrivilegeEscalation: true,
    },
  },
};

const policyWithNoDefaults: WorkspacePolicyDefaultsAndRulesV2 = {
  rules: {
    security: {
      allowPrivilegeEscalation: {
        sourceOfRule: {
          scope: "project",
          projectId: 3,
        },
        required: true,
        canEdit: true,
      },
    },
  },
};

const complexPolicy: WorkspacePolicyDefaultsAndRulesV2 = {
  rules: {
    environmentVariables: {
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 3,
        },
        canAdd: true,
        locked: ["a", "b"],
      },
    },
    compute: {
      gpuMemoryRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 3,
        },
        min: "4G",
        max: "8G",
      },
      gpuDevicesRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 3,
        },
        min: 10,
        max: 20,
      },
    },
  },
  defaults: {
    environmentVariables: {
      instances: [
        { name: "a", value: "aaa" },
        { name: "b", value: "bbb" },
        { name: "c", value: "ccc" },
      ],
    },
    compute: {},
  },
};

const fullyFilledPolicy: WorkspacePolicyDefaultsAndRulesV2 = {
  rules: {
    command: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: true,
      options: [
        {
          value: "command-opt-1",
          displayed: "display-this-command",
        },
      ],
    },
    args: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: true,
      options: [
        {
          value: "args-opt-1",
          displayed: "display-this-args",
        },
      ],
    },
    image: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: true,
      options: [
        {
          value: "image-opt-1",
          displayed: "display-this-image",
        },
      ],
    },
    workingDir: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: true,
      options: [
        {
          value: "dir-opt-1",
          displayed: "display-this-dir",
        },
      ],
    },
    nodeType: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: true,
      options: [
        {
          value: "node-type-opt-1",
          displayed: "display-this-node-type",
        },
      ],
    },
    autoDeletionTimeAfterCompletionSeconds: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: false,
      min: 4,
      max: 10,
      step: 2,
    },
    backoffLimit: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: false,
      min: 4,
      max: 10,
      step: 2,
    },
    createHomeDir: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: false,
    },
    terminateAfterPreemption: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: false,
    },
    nodePools: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: false,
    },
    environmentVariables: {
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canAdd: true,
        locked: ["this", "that", "and-this-also"],
      },
    },
    imagePullPolicy: {
      sourceOfRule: {
        scope: "project",
        projectId: 12,
        departmentId: "44",
        clusterId: "clusteruuid-123123123",
      },
      required: true,
      canEdit: true,
      options: [
        {
          value: "IfNotPresent",
          displayed: "If not present",
        },
        {
          value: "Always",
          displayed: "Always",
        },
        {
          value: "Never",
          displayed: "Never",
        },
      ],
    },
    annotations: {
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canAdd: true,
        locked: ["this", "that", "and-this-also"],
      },
    },
    labels: {
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canAdd: true,
        locked: ["this", "that", "and-this-also"],
      },
    },
    security: {
      allowPrivilegeEscalation: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
      },
      runAsNonRoot: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
      },
      hostIpc: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
      },
      hostNetwork: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
      },
      supplementalGroups: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: true,
        options: [
          {
            value: "supplementalGroups-opt-1",
            displayed: "display-this-supplementalGroups",
          },
        ],
      },
      capabilities: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: true,
      },
      uidGidSource: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: false,
        canEdit: true,
        options: [
          {
            value: "fromTheImage",
            displayed: "that-image",
          },
          {
            value: "fromIdpToken",
            displayed: "this-idp",
          },
          {
            value: "custom",
            displayed: "custom-this",
          },
        ],
      },
      seccompProfileType: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: true,
        options: [
          {
            value: "RuntimeDefault",
            displayed: "RuntimeDefault-image",
          },
          {
            value: "Unconfined",
            displayed: "Unconfined-idp",
          },
          {
            value: "Localhost",
            displayed: "Localhost-this",
          },
        ],
      },
      runAsUid: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
      runAsGid: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
    },
    compute: {
      cpuCoreRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
      cpuCoreLimit: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
      gpuPortionLimit: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
      gpuPortionRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
      cpuMemoryRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: "4G",
        max: "10G",
      },
      cpuMemoryLimit: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: "4G",
        max: "10G",
      },
      gpuMemoryRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: "4G",
        max: "10G",
      },
      gpuMemoryLimit: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: "4G",
        max: "10G",
      },
      largeShmRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
      },
      gpuDevicesRequest: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: false,
        min: 4,
        max: 10,
        step: 2,
      },
      gpuRequestType: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        required: true,
        canEdit: true,
        options: [
          {
            value: "portion",
            displayed: "portion-image",
          },
          {
            value: "memory",
            displayed: "memory-idp",
          },
        ],
      },
      extendedResources: {
        attributes: {
          quantity: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            options: [
              { value: "this", displayed: "thisDis" },
              { value: "that", displayed: "thatDis" },
              { value: "and-this-also", displayed: "and-this-also-Dis" },
            ],
          },
        },
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["this", "that", "and-this-also"],
        },
      },
    },
    exposedUrls: {
      attributes: {
        container: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          min: 4,
          max: 10,
          step: 2,
        },
        customUrl: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
        },
        authorizedUsers: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
        },
        url: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "memory",
              displayed: "memory-idp",
            },
          ],
        },
        toolType: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "memory",
              displayed: "memory-idp",
            },
          ],
        },
        toolName: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "memory",
              displayed: "memory-idp",
            },
          ],
        },
      },
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canAdd: true,
        locked: ["memory"],
      },
    },
    relatedUrls: {
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canAdd: true,
        locked: ["NA-ME"],
      },
      attributes: {
        url: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "url",
              displayed: "url",
            },
          ],
        },
        type: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "type",
              displayed: "type",
            },
          ],
        },
        name: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "NA-ME",
              displayed: "NAME",
            },
          ],
        },
      },
    },
    ports: {
      attributes: {
        container: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          min: 4,
          max: 10,
          step: 2,
        },
        external: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          min: 4,
          max: 10,
          step: 2,
        },
        customExternalPort: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
        },
        toolType: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "tool",
              displayed: "tool",
            },
          ],
        },
        toolName: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "tool-NA-ME",
              displayed: "tool-NAME",
            },
          ],
        },
        serviceType: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: false,
          options: [
            {
              value: "LoadBalancer",
              displayed: "lb",
            },
            {
              value: "NodePort",
              displayed: "np",
            },
            {
              value: "ClusterIP",
              displayed: "cip",
            },
          ],
        },
      },
      instances: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        canAdd: true,
        locked: ["LoadBalancer"],
      },
    },
    storage: {
      hostPath: {
        attributes: {
          path: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          readOnly: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          mountPath: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
        },
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["hp-nonesense"],
        },
      },
      nfs: {
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["nfs-nonesense"],
        },
        attributes: {
          path: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          readOnly: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          server: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          mountPath: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
        },
      },
      pvc: {
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["pvc-nonesense"],
        },
        attributes: {
          claimName: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          path: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          readOnly: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          claimInfo: {
            size: {
              sourceOfRule: {
                scope: "project",
                projectId: 12,
                departmentId: "44",
                clusterId: "clusteruuid-123123123",
              },
              required: true,
              canEdit: true,
              options: [
                {
                  value: "claim-1",
                  displayed: "display-this-claim",
                },
              ],
            },
            storageClass: {
              sourceOfRule: {
                scope: "project",
                projectId: 12,
                departmentId: "44",
                clusterId: "clusteruuid-123123123",
              },
              required: true,
              canEdit: true,
              options: [
                {
                  value: "storage-opt-1",
                  displayed: "display-this-storage",
                },
              ],
            },
          },
        },
      },
      git: {
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["git-nonesense"],
        },
        attributes: {
          repository: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
            options: [
              {
                value: "repo-opt-1",
                displayed: "display-this-repo",
              },
            ],
          },
          branch: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
            options: [
              {
                value: "branch-opt-1",
                displayed: "display-this-branch",
              },
            ],
          },
          revision: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
            options: [
              {
                value: "rev-opt-1",
                displayed: "display-this-rev",
              },
            ],
          },
          path: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
            options: [
              {
                value: "path-opt-1",
                displayed: "display-this-path",
              },
            ],
          },
        },
      },
      s3: {
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["s3-nonesense"],
        },
        attributes: {
          bucket: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
            options: [
              {
                value: "bucket-opt-1",
                displayed: "display-this-bucket",
              },
            ],
          },
          path: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          url: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
        },
      },
      configMapVolume: {
        instances: {
          sourceOfRule: {
            scope: "project",
            projectId: 12,
            departmentId: "44",
            clusterId: "clusteruuid-123123123",
          },
          canAdd: true,
          locked: ["cm-nonesense"],
        },
        attributes: {
          configMap: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
          mountPath: {
            sourceOfRule: {
              scope: "project",
              projectId: 12,
              departmentId: "44",
              clusterId: "clusteruuid-123123123",
            },
            required: true,
            canEdit: false,
          },
        },
      },
    },
  },
  defaults: {
    command: "some-command",
    args: "some-args",
    environmentVariables: {
      instances: [
        { name: "a", value: "aaa", exclude: true },
        { name: "b", value: "bbb", exclude: false },
        { name: "c", value: "ccc", exclude: true },
      ],
    },
    image: "some-image",
    workingDir: "some-dir",
    nodeType: "some-node-type",
    createHomeDir: true,
    nodePools: ["node-1", "node-2"],
    imagePullPolicy: "IfNotPresent",
    compute: {
      gpuDevicesRequest: 4,
      gpuRequestType: "portion",
      gpuPortionRequest: 32,
      gpuPortionLimit: 55,
      gpuMemoryRequest: "4G",
      gpuMemoryLimit: "8G",
      cpuCoreRequest: 3,
      cpuCoreLimit: 6,
      cpuMemoryRequest: "3M",
      cpuMemoryLimit: "12M",
      largeShmRequest: true,
      extendedResources: {
        instances: [
          {
            resource: "instance-1-res",
            quantity: "instance-1-quan",
            exclude: true,
          },
          {
            resource: "instance-2-res",
            quantity: "instance-2-quan",
            exclude: false,
          },
        ],
        attributes: {
          resource: "ER-resource-attribute",
          quantity: "ER-quantity-attribute",
          exclude: true,
        },
      },
    },
    storage: {
      hostPath: {
        instances: [
          {
            path: "hp-instance-path-1",
            readOnly: true,
            mountPath: "hp-instance-mount-1",
          },
          {
            path: "hp-instance-path-2",
            readOnly: false,
            mountPath: "hp-instance-mount-2",
          },
        ],
        attributes: {
          path: "hp-attr-path",
          readOnly: false,
          mountPath: "hp-attr-mountPath",
        },
      },
      nfs: {
        instances: [
          {
            path: "nfs-instance-path",
            readOnly: true,
            server: "nfs-instance-server",
            mountPath: "nfs-instance-mount",
          },
          {
            path: "nfs-instance-path-2",
            readOnly: false,
            server: "nfs-instance-server-2",
            mountPath: "nfs-instance-mount-2",
          },
        ],
        attributes: {
          path: "nfs-attr-path",
          readOnly: false,
          server: "nfs-attr-server",
          mountPath: "nfs-attr-mount",
        },
      },
      pvc: {
        instances: [
          {
            existingPvc: false,
            claimName: "pvc-instance-claim",
            path: "pvc-instance-path",
            readOnly: true,
            ephemeral: false,
            claimInfo: {
              size: "3G",
              storageClass: "pvc-instance-strage-class",
              accessModes: {
                readWriteOnce: true,
                readOnlyMany: true,
                readWriteMany: true,
              },
              volumeMode: "Filesystem",
            },
          },
          {
            existingPvc: false,
            claimName: "pvc-instance-claim-2",
            path: "pvc-instance-path-2",
            readOnly: true,
            ephemeral: false,
            claimInfo: {
              size: "3G",
              storageClass: "pvc-instance-strage-class-2",
              accessModes: {
                readWriteOnce: true,
                readOnlyMany: true,
                readWriteMany: true,
              },
              volumeMode: "Filesystem",
            },
          },
        ],
        attributes: {
          existingPvc: false,
          claimName: "pvc-attr-claim",
          path: "pvc-attr-path",
          readOnly: true,
          ephemeral: false,
          claimInfo: {
            size: "3G",
            storageClass: "pvc-attr-stgclass",
            accessModes: {
              readWriteOnce: true,
              readOnlyMany: true,
              readWriteMany: true,
            },
            volumeMode: "Filesystem",
          },
        },
      },
      git: {
        instances: [
          {
            repository: "git-instance-rep",
            branch: "git-instance-branch",
            revision: "git-instance-rev",
            path: "git-instance-path",
            passwordSecret: "git-instance-pass-sec",
            secretKeyOfUser: "git-instance-secret-user",
            secretKeyOfPassword: "git-instance-secret-pass",
          },
          {
            repository: "git-instance-2-rep",
            branch: "git-instance-2-branch",
            revision: "git-instance-2-rev",
            path: "git-instance-2-path",
            passwordSecret: "git-instance-2-pass-sec",
            secretKeyOfUser: "git-instance-2-secret-user",
            secretKeyOfPassword: "git-instance-2-secret-pass",
          },
        ],
        attributes: {
          repository: "git-attr-rep",
          branch: "git-attr-branch",
          revision: "git-attr-rev",
          path: "git-attr-path",
          passwordSecret: "git-attr-pass-sec",
          secretKeyOfUser: "git-attr-secret-user",
          secretKeyOfPassword: "git-attr-secret-pass",
        },
      },
      s3: {
        instances: [
          {
            bucket: "s3-instance-buck",
            path: "s3-instance-path",
            url: "s3-instance-url",
            accessKeySecret: "s3-instance-access-key",
            secretKeyOfAccessKeyId: "s3-instance-secret-id",
            secretKeyOfSecretKey: "s3-instance-secret-key",
          },
        ],
        attributes: {
          bucket: "s3-attr-buck",
          path: "s3-attr-path",
          url: "s3-attr-url",
          accessKeySecret: "s3-attr-access-key",
          secretKeyOfAccessKeyId: "s3-attr-secret-id",
          secretKeyOfSecretKey: "s3-attr-secret-key",
        },
      },
      configMapVolume: {
        instances: [
          {
            configMap: "CM-instance-configmap",
            mountPath: "CM-instance-mount",
          },
          {
            configMap: "CM-instance-2-map",
            mountPath: "CM-instance-2-mount",
          },
        ],
        attributes: {
          configMap: "CM-attr-configMap",
          mountPath: "CM-attr-mountPath",
        },
      },
    },
    security: {
      allowPrivilegeEscalation: true,
      uidGidSource: "fromIdpToken",
      capabilities: ["BLOCK_SUSPEND", "DAC_OVERRIDE", "LINUX_IMMUTABLE", "AUDIT_WRITE", "IPC_LOCK"],
      seccompProfileType: "RuntimeDefault",
      runAsNonRoot: false,
      hostIpc: true,
      hostNetwork: false,
      runAsUid: 1111,
      runAsGid: 2222,
      supplementalGroups: "suppp-grp",
    },
    annotations: {
      instances: [
        { name: "anno-a", value: "ann-aaa", exclude: true },
        { name: "anno-b", value: "ann-bbb", exclude: false },
        { name: "anno-c", value: "ann-ccc", exclude: true },
      ],
    },
    labels: {
      instances: [
        { name: "label-a", value: "label-aaa", exclude: true },
        { name: "label-b", value: "label-bbb", exclude: false },
        { name: "label-c", value: "label-ccc", exclude: true },
      ],
    },
    ports: {
      instances: [
        {
          container: 21312312,
          serviceType: "LoadBalancer",
          external: 44444,
          toolType: "port-tool-type-instance",
          toolName: "port-tool-name-instance",
        },
        {
          container: 123123,
          serviceType: "NodePort",
          external: 443,
          toolType: "port-tool-type-instance-2",
          toolName: "port-tool-name-instance-2",
        },
      ],
      attributes: {
        container: 123123,
        serviceType: "NodePort",
        external: 443,
        toolType: "port-tool-type-attr",
        toolName: "port-tool-name-attr",
      },
    },
    exposedUrls: {
      instances: [
        {
          container: 12345,
          url: "https://url.com",
          authorizedUsers: ["user-1", "user-2", "user-3"],
          toolType: "some-tool-instance",
          toolName: "some-name-instace",
        },
        {
          container: 54321,
          url: "https://url-2.com",
          authorizedUsers: ["user-4", "user-5", "user-6"],
          toolType: "some-tool-instance-2",
          toolName: "some-name-instance-2",
        },
      ],
      attributes: {
        container: 54321,
        url: "https://url-attr.com",
        authorizedUsers: ["user-attr-4", "user-attr-5", "user-attr-6"],
        toolType: "some-tool-attr",
        toolName: "some-name-attr",
      },
    },
    relatedUrls: {
      instances: [
        {
          url: "https://some-url-instance.com",
          type: "some-relatedUrl-type-instance",
          name: "some-relatedUrl-name-instance",
        },
        {
          url: "https://some-url-instance-2.com",
          type: "some-relatedUrl-type-instance-2",
          name: "some-relatedUrl-name-instance-2",
        },
      ],
      attributes: {
        url: "https://some-url-attr.com",
        type: "some-relatedUrl-type-attr",
        name: "some-relatedUrl-name-attr",
      },
    },
  },
  imposedAssets: ["data-1-id", "data-2-id", "data-3-id"],
};

const distributedPolicy: DistributedPolicyDefaultsAndRulesV2 = {
  defaults: {
    master: {
      command: "cmd-default-master",
      image: "image-default-master",
    },
    worker: {
      command: "cmd-default-worker",
      image: "image-default-worker",
      security: {
        supplementalGroups: "supplementalGroups-opt-1",
        capabilities: ["BLOCK_SUSPEND", "DAC_OVERRIDE", "LINUX_IMMUTABLE", "AUDIT_WRITE", "IPC_LOCK"],
        // allowPrivilegeEscalation: true,
        // uidGidSource: "fromIdpToken",
      },
    },
  },
  rules: {
    master: {
      image: {
        sourceOfRule: {
          scope: "project",
          projectId: 12,
          departmentId: "44",
          clusterId: "clusteruuid-123123123",
        },
        options: [
          { displayed: "master-image-1", value: "image-default-master" },
          { displayed: "master-image-2", value: "image-default-master-2" },
        ],
      },
    },
    worker: {
      security: {
        supplementalGroups: {
          sourceOfRule: {
            scope: "department",
            departmentId: "2",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: true,
          options: [
            {
              value: "supplementalGroups-opt-1",
              displayed: "display-this-supplementalGroups",
            },
          ],
        },
        capabilities: {
          sourceOfRule: {
            scope: "department",
            departmentId: "2",
            clusterId: "clusteruuid-123123123",
          },
          required: true,
          canEdit: true,
        },
      },
    },
  },
};

export const inferenceDistributedPolicy: UnifiedPolicyInfo = {
  worker: {
    defaults: {
      security: {
        allowPrivilegeEscalation: true,
      },
      image: "shared-default-image",
      imagePullPolicy: "IfNotPresent",
      leader: {
        image: "leader-default-image",
        command: "leader-cmd",
      },
      worker: {
        image: "worker-default-image",
        command: "worker-cmd",
      },
    },
    rules: {
      security: {
        allowPrivilegeEscalation: {
          sourceOfRule: { scope: "project", projectId: 1 },
          required: true,
          canEdit: true,
        },
      },
      image: {
        sourceOfRule: { scope: "project", projectId: 1 },
        options: [{ displayed: "Shared", value: "shared-default-image" }],
      },
      leader: {
        image: {
          sourceOfRule: { scope: "project", projectId: 2 },
          options: [{ displayed: "Leader Image", value: "leader-default-image" }],
        },
      },
      worker: {
        image: {
          sourceOfRule: { scope: "department", departmentId: "dept-1" },
          options: [{ displayed: "Worker Image", value: "worker-default-image" }],
        },
      },
    },
  },
};

export const policyUtilMock = {
  simplePolicy,
  policyWithNoRules,
  policyWithNoDefaults,
  complexPolicy,
  fullyFilledPolicy,
  distributedPolicy,
  inferenceDistributedPolicy,
};
