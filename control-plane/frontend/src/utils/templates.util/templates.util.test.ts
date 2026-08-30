import { describe, it, expect } from "vitest";
import { Scope, WorkloadType } from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { TemplateMetaFields } from "@/swagger-models/workloads-client";
import { templatesUtil } from "./templates.util";

describe("templates.util", () => {
  describe("isPermittedTemplate", () => {
    const createTemplateMeta = (scopeType: Scope, scopeId: string): TemplateMetaFields => ({
      scopeType,
      scopeId,
      id: "123",
      name: "test",
      workloadType: WorkloadType.Training,
      createdBy: "test",
      createdAt: "2021-01-01",
    });

    describe("System scope", () => {
      it("should return true when user has tenant-level permission and tenantId is provided", () => {
        const templateMeta = createTemplateMeta(Scope.System, "system-id");
        const permissions: PermittedScopes = { tenants: ["123"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions, "123")).toBe(true);
      });

      it("should return false when user has tenant-level permission but tenantId is not in the list", () => {
        const templateMeta = createTemplateMeta(Scope.System, "system-id");
        const permissions: PermittedScopes = { tenants: ["456"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions, "123")).toBe(false);
      });

      it("should return false when tenantId is not provided", () => {
        const templateMeta = createTemplateMeta(Scope.System, "system-id");
        const permissions: PermittedScopes = { tenants: ["123"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should return false when tenants array is empty", () => {
        const templateMeta = createTemplateMeta(Scope.System, "system-id");
        const permissions: PermittedScopes = { tenants: [] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions, "123")).toBe(false);
      });

      it("should return false when tenants is undefined", () => {
        const templateMeta = createTemplateMeta(Scope.System, "system-id");
        const permissions: PermittedScopes = {};

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions, "123")).toBe(false);
      });
    });

    describe("Tenant scope", () => {
      it("should return true when scopeId is in tenants array", () => {
        const templateMeta = createTemplateMeta(Scope.Tenant, "tenant-123");
        const permissions: PermittedScopes = { tenants: ["tenant-123", "tenant-456"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(true);
      });

      it("should return false when scopeId is not in tenants array", () => {
        const templateMeta = createTemplateMeta(Scope.Tenant, "tenant-123");
        const permissions: PermittedScopes = { tenants: ["tenant-456"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should return false when tenants is undefined", () => {
        const templateMeta = createTemplateMeta(Scope.Tenant, "tenant-123");
        const permissions: PermittedScopes = {};

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should handle empty scopeId", () => {
        const templateMeta = createTemplateMeta(Scope.Tenant, "");
        const permissions: PermittedScopes = { tenants: [""] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(true);
      });

      it("should handle undefined scopeId by checking for empty string", () => {
        const templateMeta = createTemplateMeta(Scope.Tenant, "");
        const permissions: PermittedScopes = { tenants: [""] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(true);
      });
    });

    describe("Cluster scope", () => {
      it("should return true when scopeId is in clusters array", () => {
        const templateMeta = createTemplateMeta(Scope.Cluster, "cluster-abc");
        const permissions: PermittedScopes = { clusters: ["cluster-abc", "cluster-xyz"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(true);
      });

      it("should return false when scopeId is not in clusters array", () => {
        const templateMeta = createTemplateMeta(Scope.Cluster, "cluster-abc");
        const permissions: PermittedScopes = { clusters: ["cluster-xyz"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should return false when clusters is undefined", () => {
        const templateMeta = createTemplateMeta(Scope.Cluster, "cluster-abc");
        const permissions: PermittedScopes = {};

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });
    });

    describe("Department scope", () => {
      it("should return true when scopeId is in departments array", () => {
        const templateMeta = createTemplateMeta(Scope.Department, "dept-1");
        const permissions: PermittedScopes = { departments: ["dept-1", "dept-2"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(true);
      });

      it("should return false when scopeId is not in departments array", () => {
        const templateMeta = createTemplateMeta(Scope.Department, "dept-1");
        const permissions: PermittedScopes = { departments: ["dept-2"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should return false when departments is undefined", () => {
        const templateMeta = createTemplateMeta(Scope.Department, "dept-1");
        const permissions: PermittedScopes = {};

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });
    });

    describe("Project scope", () => {
      it("should return true when scopeId is in projects array", () => {
        const templateMeta = createTemplateMeta(Scope.Project, "project-a");
        const permissions: PermittedScopes = { projects: ["project-a", "project-b"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(true);
      });

      it("should return false when scopeId is not in projects array", () => {
        const templateMeta = createTemplateMeta(Scope.Project, "project-a");
        const permissions: PermittedScopes = { projects: ["project-b"] };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should return false when projects is undefined", () => {
        const templateMeta = createTemplateMeta(Scope.Project, "project-a");
        const permissions: PermittedScopes = {};

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });
    });

    describe("Unknown scope", () => {
      it("should return false for unknown scope type", () => {
        const templateMeta = createTemplateMeta("unknown" as Scope, "some-id");
        const permissions: PermittedScopes = {
          tenants: ["some-id"],
          clusters: ["some-id"],
          departments: ["some-id"],
          projects: ["some-id"],
        };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });
    });

    describe("Edge cases", () => {
      it("should handle empty permissions object", () => {
        const templateMeta = createTemplateMeta(Scope.Project, "project-a");
        const permissions: PermittedScopes = {};

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should handle permissions with all empty arrays", () => {
        const templateMeta = createTemplateMeta(Scope.Tenant, "tenant-123");
        const permissions: PermittedScopes = {
          tenants: [],
          clusters: [],
          departments: [],
          projects: [],
        };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });

      it("should not check wrong scope arrays", () => {
        const templateMeta = createTemplateMeta(Scope.Project, "project-a");
        const permissions: PermittedScopes = {
          tenants: ["project-a"],
          clusters: ["project-a"],
          departments: ["project-a"],
          projects: [],
        };

        expect(templatesUtil.isPermittedTemplate(templateMeta, permissions)).toBe(false);
      });
    });
  });
});
