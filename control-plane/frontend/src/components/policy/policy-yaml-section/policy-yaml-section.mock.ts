import type { WorkspacePolicyV2, WorkspacePolicyDefaultsAndRulesV2 } from "@/swagger-models/policy-service-client";
import type { IPolicyMetaAndType } from "@/views/policy/policy-edit.vue";

export const mockHugePolicy: WorkspacePolicyV2 = {
  meta: {
    id: "7796c719-0c6a-424e-8911-bfe5a4a54a13",
    clusterId: "7cf06a0e-8f67-4d91-82c3-879b4625a2b3",
    createdBy: "test@run.ai",
    createdAt: "2023-11-01T10:33:55.979929Z",
    updatedBy: "test@run.ai",
    updatedAt: "0001-01-01T00:00:00Z",
    scope: "project",
    projectId: 1,
    name: "rotem-1-workspace",
  },
  policy: {
    defaults: {
      environment: {
        command: "how-long",
        args: "x=1",
        environmentVariables: [
          {
            name: "oh-yeah",
            value: "some-val",
          },
        ],
        runAsUid: 2,
        runAsGid: 4,
        supplementalGroups: "another-group",
        image: "ubuntu",
        imagePullPolicy: "Always",
        workingDir: "/homesweet/home",
        hostIpc: true,
        hostNetwork: false,
        connections: [
          {
            name: "pyTorch",
            isExternal: true,
            internalToolInfo: {
              toolType: "pycharm",
              connectionType: "LoadBalancer",
              containerPort: 8080,
              nodePortInfo: {
                isCustomPort: true,
              },
              externalUrlInfo: {
                isCustomUrl: true,
                externalUrl: "https://rotem.com",
              },
            },
            externalToolInfo: {
              toolType: "wandb",
              externalUrl: "https://rotem.com",
            },
          },
          {
            name: "VSCode",
            isExternal: true,
            internalToolInfo: {
              toolType: "matlab",
              connectionType: "ExternalUrl",
              containerPort: 9090,
              nodePortInfo: {
                isCustomPort: true,
              },
              externalUrlInfo: {
                isCustomUrl: true,
                externalUrl: "https://external-url-info.com",
              },
            },
            externalToolInfo: {
              toolType: "wandb",
              externalUrl: "https://External-tool-info.com",
            },
          },
        ],
        createHomeDir: false,
        allowPrivilegeEscalation: true,
        uidGidSource: "fromTheImage",
        overrideUidGidInWorkspace: true,
        capabilities: ["FOWNER", "KILL"],
      },
      compute: {
        gpuRequestType: "portion",
        gpuPortion: 1,
        gpuMemory: "67",
        cpuCoreRequest: 7,
        cpuCoreLimit: 7,
        cpuMemoryRequest: "67",
        cpuMemoryLimit: "67",
        largeShmRequest: true,
        extendedResources: [
          {
            resource: "some-resource",
            quantity: "i said no",
          },
          {
            resource: "another-resource",
            quantity: "another-quantity",
          },
        ],
      },
      hostPath: {
        path: "path-this",
        readOnly: false,
        mountPath: "some-val",
      },
      nfs: {
        path: "/ahoy",
        readOnly: false,
        server: "mate",
        mountPath: "mount-this",
      },
      pvc: {
        existingPvc: false,
        claimName: "claim-this",
        path: "path-this",
        readOnly: false,
        ephemeral: false,
        claimInfo: {
          size: "2G",
          storageClass: "Storage this",
          accessModes: {
            readWriteOnce: true,
            readOnlyMany: false,
            readWriteMany: true,
          },
          volumeMode: "Filesystem",
        },
      },
      git: {
        repository: "repo-tipo",
        branch: "oy branch",
        revision: "some-rev",
        path: "some-path",
        passwordAssetId: "oh yeah",
      },
      s3: {
        bucket: "buck",
        path: "pathy",
        accessKeyAssetId: "shhh",
        url: "/https://url.com",
      },
      workspace: {
        nodeType: "node-type",
        nodePools: ["rotem", "popa", "hopa"],
        allowOverQuota: true,
        annotations: [
          {
            name: "hopa0ani0popa",
            value: "anno",
          },
        ],
        labels: [
          {
            name: "label-1",
            value: "la-be-l",
          },
        ],
        autoDeletionTimeAfterCompletionSeconds: 3,
      },
    },
    rules: {
      environment: {
        allowPrivilegeEscalation: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        args: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        capabilities: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        command: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        createHomeDir: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        environmentVariables: {
          itemRules: {
            sourceOfRule: {
              scope: "project",
              projectId: 1,
            },
            canAdd: false,
          },
        },
        hostIpc: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        hostNetwork: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        image: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        imagePullPolicy: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        overrideUidGidInWorkspace: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        runAsUid: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
          min: 1,
          max: 5,
          step: 1,
        },
        runAsGid: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
          min: 1,
          max: 5,
          step: 1,
        },
        supplementalGroups: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        uidGidSource: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        workingDir: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
      },
      compute: {
        cpuCoreRequest: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
          min: 1,
          max: 7,
          step: 1.5,
        },
        cpuCoreLimit: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
          min: 1,
          max: 7,
          step: 1.5,
        },
        cpuMemoryRequest: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
          min: "67",
          max: "72",
        },
        cpuMemoryLimit: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
          min: "67",
          max: "72",
        },
        extendedResources: {
          itemRules: {
            sourceOfRule: {
              scope: "project",
              projectId: 1,
            },
            canAdd: false,
          },
          members: {
            quantity: {
              sourceOfRule: {
                scope: "project",
                projectId: 1,
              },
              canEdit: true,
            },
          },
        },
        largeShmRequest: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        gpuRequestType: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        gpuMemory: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
          min: "67",
          max: "72",
        },
        gpuPortion: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
          min: 1,
          max: 7,
          step: 1.5,
        },
      },
      hostPath: {
        path: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        readOnly: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        mountPath: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
      },
      nfs: {
        path: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        readOnly: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        server: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        mountPath: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
      },
      pvc: {
        claimName: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        path: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        readOnly: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        claimInfo: {
          size: {
            sourceOfRule: {
              scope: "project",
              projectId: 1,
            },
            required: true,
            canEdit: true,
          },
          storageClass: {
            sourceOfRule: {
              scope: "project",
              projectId: 1,
            },
            required: true,
            canEdit: true,
          },
        },
      },
      git: {
        repository: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        branch: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        revision: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        path: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
      },
      s3: {
        bucket: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        path: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        url: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
      },
      imposedAssets: {
        datasources: {
          hostPath: {
            canAdd: true,
          },
          nfs: {
            canAdd: true,
          },
          pvc: {
            canAdd: true,
          },
          git: {
            canAdd: true,
          },
          s3: {
            canAdd: true,
          },
        },
      },
      workspace: {
        nodeType: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        nodePools: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        allowOverQuota: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
        annotations: {
          itemRules: {
            sourceOfRule: {
              scope: "project",
              projectId: 1,
            },
            canAdd: false,
          },
        },
        labels: {
          itemRules: {
            sourceOfRule: {
              scope: "project",
              projectId: 1,
            },
            canAdd: false,
          },
        },
        autoDeletionTimeAfterCompletionSeconds: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
          min: 1,
          max: 5,
          step: 1,
        },
      },
    },
  } as WorkspacePolicyDefaultsAndRulesV2,
};

export const mockSimplePolicy: WorkspacePolicyV2 = {
  policy: {
    defaults: {
      environment: {
        command: "some-command",
        args: "y=4",
        environmentVariables: [
          {
            name: "das-ist-ein-env-var",
            value: "some-val",
          },
        ],
        overrideUidGidInWorkspace: true,
        capabilities: ["FOWNER"],
      },
    },
    rules: {
      environment: {
        allowPrivilegeEscalation: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: true,
          canEdit: true,
        },
        args: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        command: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          canEdit: true,
        },
        workingDir: {
          sourceOfRule: {
            scope: "project",
            projectId: 1,
          },
          required: false,
          canEdit: true,
        },
      },
    },
  } as WorkspacePolicyDefaultsAndRulesV2,
};

export const mockPolicyMetaAndType: IPolicyMetaAndType = {
  policyType: "workspace",
  policyMeta: {
    id: "7796c719-0c6a-424e-8911-bfe5a4a54a13",
    clusterId: "7cf06a0e-8f67-4d91-82c3-879b4625a2b3",
    createdBy: "test@run.ai",
    createdAt: "2023-11-01T10:33:55.979929Z",
    updatedBy: "test@run.ai",
    updatedAt: "0001-01-01T00:00:00Z",
    scope: "project",
    projectId: 1,
    name: "some-proj-workspace",
  },
};
