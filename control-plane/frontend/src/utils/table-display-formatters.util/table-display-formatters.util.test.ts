import { describe, it, expect } from "vitest";
import { tableDisplayFormatters } from "./table-display-formatters.util";
import { Scope } from "@/swagger-models/assets-service-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import type { ITableColumn, IStatusColOptions } from "@/models/table.model";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { type Tenant, TenantStatusEnum, TenantContractTypeEnum } from "@/swagger-models/tenants-manager-client";

describe("tableDisplayFormatters", () => {
  const mockTenant: Tenant = {
    id: 123,
    name: "Test Tenant",
    uuid: "tenant-uuid-123",
    smg: false,
    status: TenantStatusEnum.Ready,
    createdAt: "2023-01-01T00:00:00Z",
    updatedAt: "2023-01-01T00:00:00Z",
    contractType: TenantContractTypeEnum.Normal,
  };

  const mockOrgUnitList: OrgUnit[] = [
    { id: "proj-1", name: "Project Alpha", type: Scope.Project } as OrgUnit,
    { id: "dept-1", name: "Department Beta", type: Scope.Department } as OrgUnit,
    { id: "cluster-1", name: "Cluster Gamma", type: Scope.Cluster } as OrgUnit,
    { id: "123", name: "Test Tenant", type: Scope.Tenant } as OrgUnit,
  ];

  const mockClusterList: DisplayedCluster[] = [
    {
      uuid: "cluster-uuid-1",
      name: "Cluster One",
      tenantId: 123,
      createdAt: "2023-01-01T00:00:00Z",
    },
    {
      uuid: "cluster-uuid-2",
      name: "Cluster Two",
      tenantId: 123,
      createdAt: "2023-01-01T00:00:00Z",
    },
  ];

  describe("getScopeName", () => {
    it("should return project name for project scope", () => {
      const row = {
        meta: {
          scope: Scope.Project,
          projectId: "proj-1",
        },
      };

      const result = tableDisplayFormatters.getScopeName(row, mockTenant, mockOrgUnitList);
      expect(result).toBe("Project Alpha");
    });

    it("should return department name for department scope", () => {
      const row = {
        meta: {
          scope: Scope.Department,
          departmentId: "dept-1",
        },
      };

      const result = tableDisplayFormatters.getScopeName(row, mockTenant, mockOrgUnitList);
      expect(result).toBe("Department Beta");
    });

    it("should return tenant name for system scope", () => {
      const row = {
        meta: {
          scope: Scope.System,
        },
      };

      const result = tableDisplayFormatters.getScopeName(row, mockTenant, mockOrgUnitList);
      expect(result).toBe("Test Tenant");
    });

    it("should return empty string when org unit not found", () => {
      const row = {
        meta: {
          scope: Scope.Project,
          projectId: "unknown-proj",
        },
      };

      const result = tableDisplayFormatters.getScopeName(row, mockTenant, mockOrgUnitList);
      expect(result).toBe("");
    });
  });

  describe("getScopeId", () => {
    it("should return project ID for project scope", () => {
      const row = {
        meta: {
          scope: Scope.Project,
          projectId: 123,
        },
      };

      const result = tableDisplayFormatters.getScopeId(row, mockTenant);
      expect(result).toBe("123");
    });

    it("should return tenant ID for default case", () => {
      const row = {
        meta: {
          scope: Scope.Tenant,
        },
      };

      const result = tableDisplayFormatters.getScopeId(row, mockTenant);
      expect(result).toBe("123");
    });
  });

  describe("getClusterName", () => {
    it("should return cluster name when found", () => {
      const result = tableDisplayFormatters.getClusterName("cluster-uuid-1", mockClusterList);
      expect(result).toBe("Cluster One");
    });

    it("should return dash when cluster not found", () => {
      const result = tableDisplayFormatters.getClusterName("unknown-uuid", mockClusterList);
      expect(result).toBe("-");
    });

    it("should return dash when cluster UUID is null", () => {
      const result = tableDisplayFormatters.getClusterName(null, mockClusterList);
      expect(result).toBe("-");
    });

    it("should return ALL label when ignoreFilter condition is met", () => {
      interface TestRow {
        isSpecial: boolean;
      }
      const column: Partial<ITableColumn> = {
        ignoreFilter: (row: TestRow) => row.isSpecial,
      };
      const row: TestRow = { isSpecial: true };

      const result = tableDisplayFormatters.getClusterName(null, mockClusterList, column as ITableColumn, row);
      expect(result).toBe("All");
    });
  });

  describe("getListPreview", () => {
    it("should handle string input", () => {
      const result = tableDisplayFormatters.getListPreview("single-item");
      expect(result).toBe("single-item");
    });

    it("should show preview for arrays with more than one item", () => {
      const result = tableDisplayFormatters.getListPreview(["item1", "item2", "item3"]);
      expect(result).toBe("item1 + 2 more");
    });

    it("should show single item without preview", () => {
      const result = tableDisplayFormatters.getListPreview(["single"]);
      expect(result).toBe("single");
    });
  });

  describe("getStatusText", () => {
    it("should return status text from options", () => {
      const statusOptions: IStatusColOptions = {
        status: "Running",
        displayAnimation: false,
      };

      const result = tableDisplayFormatters.getStatusText(statusOptions);
      expect(result).toBe("Running");
    });

    it("should return dash when status is empty", () => {
      const statusOptions: IStatusColOptions = {
        status: "",
        displayAnimation: false,
      };

      const result = tableDisplayFormatters.getStatusText(statusOptions);
      expect(result).toBe("-");
    });
  });
});
