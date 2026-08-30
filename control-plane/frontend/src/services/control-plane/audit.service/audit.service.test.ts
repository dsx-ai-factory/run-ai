import { describe, it, expect, vi } from "vitest";
import { auditService } from "./audit.service";
import { filterService } from "@/services/filter.service/filter.service";
import { auditServiceApi } from "@/services/infra/client-apis/audit-service-api/audit-service-api";
import {
  AuditLogActionEnum,
  AuditLogResultEnum,
  type AuditLog,
  type GetAuditLogs200Response,
} from "@/swagger-models/audit-service-client";
import type { IFilterBy } from "@/models/filter.model";

vi.mock("@/services/filter.service/filter.service", () => ({
  filterService: {
    mapColumnsFilterToFilterParams: vi.fn(),
  },
}));

vi.mock("@/services/infra/client-apis/audit-service-api/audit-service-api", () => ({
  auditServiceApi: {
    auditLogsApi: {
      getAuditLogs: vi.fn(),
    },
  },
}));

describe("auditService", () => {
  it("should call getAuditLogs with correct parameters when sortBy, descending, page, and rowsPerPage are provided", async () => {
    const mockStart = "2024-01-01";
    const mockEnd = "2024-01-02";
    const filterBy: IFilterBy = {
      sortBy: "timestamp",
      descending: true,
      page: 2,
      rowsPerPage: 20,
    };

    const mockFilters = { offset: 20, limit: 20, sortOrder: "desc", sortBy: "timestamp", filterBy: {} };
    filterService.mapColumnsFilterToFilterParams = vi.fn().mockReturnValue(mockFilters);

    const mockResponse: GetAuditLogs200Response = {
      total: 50,
      audit_logs: [
        {
          id: "1",
          timestamp: "2024-01-01T12:00:00",
          tenant_id: "tenant-1",
          subject: "admin",
          subject_type: "User",
          source_ip: "192.168.1.1",
          action: AuditLogActionEnum.SignIn,
          http_method: "POST",
          url: "/api/login",
          entity_type: "User",
          entity_name: "admin",
          entity_id: "user-1",
          result: AuditLogResultEnum.Succeeded,
          http_status_code: 200,
          metadata: { key: "value" },
        },
      ],
    };

    auditServiceApi.auditLogsApi.getAuditLogs = vi.fn().mockResolvedValue({ data: mockResponse });

    const result = await auditService.listAudits(filterBy, mockStart, mockEnd);

    expect(auditServiceApi.auditLogsApi.getAuditLogs).toHaveBeenCalledWith(
      mockStart,
      mockEnd,
      mockFilters.offset,
      mockFilters.limit,
      mockFilters.sortOrder,
      mockFilters.sortBy,
      mockFilters.filterBy,
    );

    expect(result).toEqual(mockResponse);
    const auditLog: AuditLog | undefined = result?.audit_logs ? result.audit_logs[0] : undefined;
    expect(auditLog).toHaveProperty("id");
    expect(auditLog).toHaveProperty("timestamp");
    expect(auditLog).toHaveProperty("tenant_id");
    expect(auditLog).toHaveProperty("subject");
    expect(auditLog).toHaveProperty("subject_type");
    expect(auditLog).toHaveProperty("source_ip");
    expect(auditLog).toHaveProperty("action");
    expect(auditLog).toHaveProperty("http_method");
    expect(auditLog).toHaveProperty("url");
    expect(auditLog).toHaveProperty("entity_type");
    expect(auditLog).toHaveProperty("entity_name");
    expect(auditLog).toHaveProperty("entity_id");
    expect(auditLog).toHaveProperty("result");
    expect(auditLog).toHaveProperty("http_status_code");
    expect(auditLog).toHaveProperty("metadata");
  });

  it("should handle multiple audit logs", async () => {
    const mockStart = "2024-01-01";
    const mockEnd = "2024-01-02";
    const filterBy: IFilterBy = {};

    const mockFilters = { offset: 0, limit: 10, sortOrder: "asc", sortBy: "timestamp", filterBy: {} };
    filterService.mapColumnsFilterToFilterParams = vi.fn().mockReturnValue(mockFilters);

    const mockResponse: GetAuditLogs200Response = {
      total: 2,
      audit_logs: [
        {
          id: "1",
          timestamp: "2024-01-01T12:00:00",
          tenant_id: "tenant-1",
          subject: "admin",
          subject_type: "User",
          source_ip: "192.168.1.1",
          action: AuditLogActionEnum.SignIn,
          http_method: "POST",
          url: "/api/login",
          entity_type: "User",
          entity_name: "admin",
          entity_id: "user-1",
          result: AuditLogResultEnum.Succeeded,
          http_status_code: 200,
          metadata: { key: "value" },
        },
        {
          id: "2",
          timestamp: "2024-01-01T13:00:00",
          tenant_id: "tenant-2",
          subject: "user",
          subject_type: "App",
          source_ip: "192.168.1.2",
          action: AuditLogActionEnum.SignOut,
          http_method: "GET",
          url: "/api/logout",
          entity_type: "User",
          entity_name: "user",
          entity_id: "user-2",
          result: AuditLogResultEnum.Failed,
          http_status_code: 400,
          metadata: { key2: "value2" },
        },
      ],
    };

    auditServiceApi.auditLogsApi.getAuditLogs = vi.fn().mockResolvedValue({ data: mockResponse });

    const result = await auditService.listAudits(filterBy, mockStart, mockEnd);

    expect(result?.audit_logs?.length).toBe(2);
    expect(result?.audit_logs?.[1].id).toBe("2");
    expect(result?.audit_logs?.[1].subject).toBe("user");
  });
});
