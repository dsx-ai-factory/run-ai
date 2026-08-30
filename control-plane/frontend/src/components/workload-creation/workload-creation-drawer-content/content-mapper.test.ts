import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { useSettingStore } from "@/stores/setting.store";
import { EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import type { IAssetsFilter } from "@/models/filter.model";
import { workloadCreationDrawerEntityMapper } from "./content-mapper";
import { WorkloadCreationDrawerEntities, type IDrawerListItem } from "@/models/workload-creation-drawer-model";
import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import { AssetKind } from "@/swagger-models/assets-service-client";

// Mock external dependencies
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/utils/workload-hybrid.util/workload-hybrid.util");
vi.mock("@/stores/setting.store");

describe("content-mapper loadDataSources", () => {
  const mockDataSources: DatasourceListResponseEntry[] = [
    {
      meta: {
        id: "1",
        name: "test-ds",
        kind: AssetKind.Nfs,
        scope: "tenant",
        tenantId: 1001,
        createdBy: "test@run.ai",
        createdAt: "2024-12-23T14:14:53.104798Z",
        updatedBy: "test@run.ai",
        updatedAt: "2024-12-23T14:14:53.104798Z",
      },
      spec: {
        nfs: {
          path: "/test-path",
          readOnly: false,
          server: "test-server",
          mountPath: "/mount-path",
        },
      },
    },
    {
      meta: {
        id: "2",
        name: "test-ds-2",
        kind: AssetKind.S3,
        scope: "tenant",
        tenantId: 1001,
        createdBy: "test@run.ai",
        createdAt: "2024-12-23T14:14:53.104798Z",
        updatedBy: "test@run.ai",
        updatedAt: "2024-12-23T14:14:53.104798Z",
      },
      spec: {
        s3: {
          bucket: "test-bucket",
          path: "/s3-mount",
          url: "https://s3.amazonaws.com",
        },
      },
    },
  ];

  const mockDrawerItems: IDrawerListItem[] = [
    {
      id: "1",
      name: "test-ds",
      kind: AssetKind.Nfs,
      details: [],
      description: "",
      icons: [],
      policy: { compliance: true },
      createdAt: "2024-12-23T14:14:53.104798Z",
      lastUsed: "",
      disabled: false,
    },
    {
      id: "2",
      name: "test-ds-2",
      kind: AssetKind.S3,
      details: [],
      description: "",
      icons: [],
      policy: { compliance: true },
      createdAt: "2024-12-23T14:14:53.104798Z",
      lastUsed: "",
      disabled: false,
    },
  ];

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    // Mock dataSourceService.list
    vi.mocked(dataSourceService.list).mockResolvedValue(mockDataSources);

    // Mock workloadHybridUtil.dataSourceListToDrawerItems
    vi.mocked(workloadHybridUtil.dataSourceListToDrawerItems).mockReturnValue(mockDrawerItems);

    // Mock useSettingStore
    vi.mocked(useSettingStore).mockReturnValue({
      isDataVolumeEnabled: true,
    } as ReturnType<typeof useSettingStore>);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("dataVolumes parameter handling", () => {
    it("should include dataVolumes=false when no projectId is provided (cluster scope)", async () => {
      // Arrange
      const filter: IAssetsFilter = {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        clusterUuid: "cluster-123", // Cluster scope - no projectId
      };

      // Act
      const loadDataSources = workloadCreationDrawerEntityMapper[WorkloadCreationDrawerEntities.DataSource].loadFunction;
      await loadDataSources(filter);

      // Assert
      expect(dataSourceService.list).toHaveBeenCalledWith({
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        clusterUuid: "cluster-123",
        dataVolumes: false, // Current implementation includes dataVolumes: false
        statusInfo: true,
      });
      expect(dataSourceService.list).toHaveBeenCalledTimes(1);
    });

    it("should include dataVolumes=false when no projectId is provided (department scope)", async () => {
      // Arrange
      const filter: IAssetsFilter = {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Training,
        departmentId: "dept-456", // Department scope - no projectId
      };

      // Act
      const loadDataSources = workloadCreationDrawerEntityMapper[WorkloadCreationDrawerEntities.DataSource].loadFunction;
      await loadDataSources(filter);

      // Assert
      expect(dataSourceService.list).toHaveBeenCalledWith({
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Training,
        departmentId: "dept-456",
        dataVolumes: false, // Current implementation includes dataVolumes: false
        statusInfo: true,
      });
    });

    it("should include dataVolumes=true when projectId is provided and data volumes are enabled", async () => {
      // Arrange
      const filter: IAssetsFilter = {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        projectId: 123,
        complyToProject: 123,
      };
      //mock isDataVolumeEnabled
      vi.mocked(useSettingStore).mockReturnValue({
        isDataVolumeEnabled: true,
      } as ReturnType<typeof useSettingStore>);

      // Act
      const loadDataSources = workloadCreationDrawerEntityMapper[WorkloadCreationDrawerEntities.DataSource].loadFunction;
      await loadDataSources(filter);

      // Assert
      expect(dataSourceService.list).toHaveBeenCalledWith({
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        projectId: 123,
        complyToProject: 123,
        dataVolumes: true, // Should be included when projectId exists
        statusInfo: true,
      });
    });

    it("should include dataVolumes=false when projectId is provided but data volumes are disabled", async () => {
      // Arrange
      vi.mocked(useSettingStore).mockReturnValue({
        isDataVolumeEnabled: false, // Data volumes disabled
      } as ReturnType<typeof useSettingStore>);

      const filter: IAssetsFilter = {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        projectId: 123,
        complyToProject: 123,
      };

      // Act
      const loadDataSources = workloadCreationDrawerEntityMapper[WorkloadCreationDrawerEntities.DataSource].loadFunction;
      await loadDataSources(filter);

      // Assert
      expect(dataSourceService.list).toHaveBeenCalledWith({
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        projectId: 123,
        complyToProject: 123,
        dataVolumes: false, // Current implementation includes dataVolumes: false when disabled
        statusInfo: true,
      });
    });
  });

  describe("data transformation", () => {
    it("should transform data sources to drawer items", async () => {
      // Arrange
      const filter: IAssetsFilter = {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        projectId: 123,
      };

      // Act
      const loadDataSources = workloadCreationDrawerEntityMapper[WorkloadCreationDrawerEntities.DataSource].loadFunction;
      const result = await loadDataSources(filter);

      // Assert
      expect(workloadHybridUtil.dataSourceListToDrawerItems).toHaveBeenCalledWith(mockDataSources);
      expect(result).toEqual(mockDrawerItems);
    });
  });

  describe("inference workload specific logic", () => {
    it("should apply inference-specific filtering when workload type is Inference", async () => {
      // This test would verify the inference-specific logic in the loadDataSources function
      // For now, just ensuring it doesn't break
      const filter: IAssetsFilter = {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
        projectId: 123,
      };

      const loadDataSources = workloadCreationDrawerEntityMapper[WorkloadCreationDrawerEntities.DataSource].loadFunction;
      const result = await loadDataSources(filter);

      expect(result).toEqual(mockDrawerItems);
    });
  });
});
