/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import type { IOrgTreeNode, IOrgTreeNodeId } from "@/models/org-tree.model";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { ScopeType } from "@/swagger-models/authorization-client";

describe("orgTreeService", () => {
  describe("getOrgTree", () => {
    it("should return the correct organizational tree for a tenant with departments and projects", () => {
      // Act
      const orgTree = orgTreeService.getOrgTree(orgTreeServiceMock.orgUnits);

      // Assert
      expect(orgTree).toEqual(orgTreeServiceMock.orgTree);
    });

    describe("getNodePathById", () => {
      const hierarchy: IOrgTreeNode[] = [
        {
          id: "1",
          path: "runai",
          type: ScopeType.Tenant,
          label: "runai",
          selectable: false,
          icon: "ph-bold ph-tree-structure",
          children: [
            {
              id: "1",
              path: "runai/default",
              type: ScopeType.Department,
              label: "default",
              icon: "ph-bold ph-garage",
              selectable: false,
              children: [
                {
                  id: "1",
                  path: "runai/default/project1",
                  type: ScopeType.Project,
                  label: "project1",
                  icon: "ph-bold ph-suitcase",
                  children: [],
                },
              ],
            },
            {
              id: "2",
              path: "runai/test123",
              type: ScopeType.Department,
              label: "test123",
              icon: "ph-bold ph-garage",
              selectable: false,
              children: [],
            },
          ],
        },
      ];

      it("returns null if the node is not found", () => {
        const result = orgTreeService.getNodePathById(hierarchy[0], { id: "0", type: ScopeType.Project });
        expect(result).toBeNull();
      });

      it("returns the correct path for a top-level tenant node", () => {
        const result = orgTreeService.getNodePathById(hierarchy[0], { id: "1", type: ScopeType.Tenant });
        expect(result).toEqual("runai");
      });

      it("returns the correct path for a department node", () => {
        const result = orgTreeService.getNodePathById(hierarchy[0], { id: "1", type: ScopeType.Department });
        expect(result).toEqual("runai/default");
      });

      it("returns the correct path for a project node", () => {
        const result = orgTreeService.getNodePathById(hierarchy[0], { id: "1", type: ScopeType.Project });
        expect(result).toEqual("runai/default/project1");
      });
    });
  });

  describe("getExpandedPermittedScopes", () => {
    it("Should return the correct permitted scopes for a tenant allowed action", () => {
      // Arrange
      const permittedScopes = {
        tenants: ["1001"],
        clusters: [],
        departments: [],
        projects: [],
      };
      const orgUnits = orgTreeServiceMock.orgUnits;

      // Act
      const result = orgTreeService.getExpandedPermittedScopes(permittedScopes, orgUnits);

      // Assert
      const expectedResult = {
        tenants: ["1001"],
        clusters: ["8a483443-8cc1-4d29-bd59-0e4e2c9082f0"],
        departments: ["1"],
        projects: ["25"],
      };
      expect(result).toEqual(expectedResult);
    });

    it("Should return the correct permitted scopes for a cluster allowed action", () => {
      // Arrange
      const permittedScopes = {
        clusters: ["8a483443-8cc1-4d29-bd59-0e4e2c9082f0"],
        departments: [],
        projects: [],
      };
      const orgUnits = orgTreeServiceMock.orgUnits;

      // Act
      const result = orgTreeService.getExpandedPermittedScopes(permittedScopes, orgUnits);

      // Assert
      const expectedResult = {
        clusters: ["8a483443-8cc1-4d29-bd59-0e4e2c9082f0"],
        departments: ["1"],
        projects: ["25"],
      };
      expect(result).toEqual(expectedResult);
    });
  });

  describe("getOrgTreeNodeId", () => {
    it("should get the correct cluster org tree node id", () => {
      const result: IOrgTreeNodeId | null = orgTreeService.getOrgTreeNodeId(
        orgTreeServiceMock.orgUnits,
        "8a483443-8cc1-4d29-bd59-0e4e2c9082f0",
        ScopeType.Cluster,
      );

      expect(result).toEqual({
        id: "8a483443-8cc1-4d29-bd59-0e4e2c9082f0",
        path: "runai/tom-1-1695017060",
        type: "cluster",
      });
    });
    it("should get the correct cluster org tree node id", () => {
      const result: IOrgTreeNodeId | null = orgTreeService.getOrgTreeNodeId(
        orgTreeServiceMock.orgUnits,
        "1",
        ScopeType.Department,
      );

      expect(result).toEqual({
        id: "1",
        type: "department",
        path: "runai/tom-1-1695017060/default",
      });
    });
    it("should return null if the node is not found", () => {
      const result: IOrgTreeNodeId | null = orgTreeService.getOrgTreeNodeId(
        orgTreeServiceMock.orgUnits,
        "1",
        ScopeType.Cluster,
      );

      expect(result).toBeNull();
    });
  });
  describe("filterHiddenScopes", () => {
    it("should filter out cluster department and project scope types", () => {
      const hiddenScopes = {
        system: false,
        tenant: false,
        cluster: true,
        department: true,
        project: true,
      };
      const result = orgTreeService.filterHiddenScopes(orgTreeServiceMock.richOrgTree, hiddenScopes);
      result.forEach((tenantScope) => {
        expect(tenantScope.children.length).eql(0);
      });
    });

    it("should filter out department and project scope types", () => {
      const hiddenScopes = {
        system: false,
        tenant: false,
        cluster: false,
        department: true,
        project: true,
      };
      const result = orgTreeService.filterHiddenScopes(orgTreeServiceMock.richOrgTree, hiddenScopes);
      result.forEach((tenantScope) => {
        expect(tenantScope.children.length).eql(3);
        tenantScope.children.forEach((clusterScope) => {
          expect(clusterScope.children.length).eql(0);
        });
      });
    });

    it("should filter out project scope types", () => {
      const hiddenScopes = {
        system: false,
        tenant: false,
        cluster: false,
        department: false,
        project: true,
      };
      const result: IOrgTreeNode[] = orgTreeService.filterHiddenScopes(orgTreeServiceMock.richOrgTree, hiddenScopes);
      result.forEach((tenantScope: IOrgTreeNode) => {
        expect(tenantScope.children.length).eql(3);
        tenantScope.children.forEach((clusterScope: IOrgTreeNode) => {
          expect(clusterScope.children.length).eql(1);
          clusterScope.children.forEach((departmentScope: IOrgTreeNode) => {
            expect(departmentScope.children.length).eql(0);
          });
        });
      });
    });

    it("should not filter out any scope types", () => {
      const hiddenScopes = {
        system: false,
        tenant: false,
        cluster: false,
        department: false,
        project: false,
      };
      const result: IOrgTreeNode[] = orgTreeService.filterHiddenScopes(orgTreeServiceMock.richOrgTree, hiddenScopes);
      expect(result).toEqual(orgTreeServiceMock.richOrgTree);
    });
  });

  describe("findMatchingNodePaths", () => {
    const nodes: IOrgTreeNode[] = [
      {
        id: "1",
        path: "/runai",
        type: ScopeType.Tenant,
        label: "runai",
        children: [
          {
            id: "5",
            path: "/runai/tom-1-1695017060",
            type: ScopeType.Cluster,
            label: "tom-1-1695017060",
            children: [
              {
                id: "2",
                path: "/runai/tom-1-1695017060/default",
                type: ScopeType.Department,
                label: "default",
                children: [
                  {
                    id: "3",
                    path: "/runai/tom-1-1695017060/default/project1",
                    type: ScopeType.Project,
                    label: "project1",
                    children: [],
                  },
                  {
                    id: "6",
                    path: "/runai/tom-1-1695017060/default/project2",
                    type: ScopeType.Project,
                    label: "project2",
                    children: [],
                  },
                ],
              },
              {
                id: "4",
                path: "/runai/tom-1-1695017060/test123",
                type: ScopeType.Department,
                label: "test123",
                children: [],
              },
            ],
          },
        ],
      },
    ];

    it("should return the correct node paths for a tenant", () => {
      let result = orgTreeService.findMatchingNodePaths(nodes, [], "runai");
      expect(result).toEqual(["/runai"]);

      result = orgTreeService.findMatchingNodePaths(nodes, [], "tom-1-1695017060");
      expect(result).toEqual(["/runai", "/runai/tom-1-1695017060"]);

      result = orgTreeService.findMatchingNodePaths(nodes, [], "test123");
      expect(result).toEqual(["/runai", "/runai/tom-1-1695017060", "/runai/tom-1-1695017060/test123"]);

      result = orgTreeService.findMatchingNodePaths(nodes, [], "project1");
      expect(result).toEqual([
        "/runai",
        "/runai/tom-1-1695017060",
        "/runai/tom-1-1695017060/default",
        "/runai/tom-1-1695017060/default/project1",
      ]);
    });
  });

  describe("sortOrgTree", () => {
    it("should recursively sort children nodes", () => {
      const input: IOrgTreeNode[] = [
        {
          id: "1",
          path: "/A",
          type: "department",
          label: "B",
          children: [
            {
              id: "3",
              path: "/A/1",
              type: "department",
              label: "C",
              children: [],
            },
            {
              id: "2",
              path: "/A/2",
              type: "department",
              label: "A",
              children: [],
            },
          ],
        },
      ];

      const expected: IOrgTreeNode[] = [
        {
          id: "1",
          path: "/A",
          type: "department",
          label: "B",
          children: [
            {
              id: "2",
              path: "/A/2",
              type: "department",
              label: "A",
              children: [],
            },
            {
              id: "3",
              path: "/A/1",
              type: "department",
              label: "C",
              children: [],
            },
          ],
        },
      ];

      expect(orgTreeService.sortOrgTree(input)).toEqual(expected);
    });

    it("should handle deeply nested structures", () => {
      const input: IOrgTreeNode[] = [
        {
          id: "1",
          path: "/A",
          type: "department",
          label: "C",
          children: [
            {
              id: "2",
              path: "/A/1",
              type: "department",
              label: "B",
              children: [
                {
                  id: "3",
                  path: "/A/1/1",
                  type: "department",
                  label: "D",
                  children: [],
                },
                {
                  id: "4",
                  path: "/A/1/2",
                  type: "department",
                  label: "A",
                  children: [],
                },
              ],
            },
          ],
        },
      ];

      const expected: IOrgTreeNode[] = [
        {
          id: "1",
          path: "/A",
          type: "department",
          label: "C",
          children: [
            {
              id: "2",
              path: "/A/1",
              type: "department",
              label: "B",
              children: [
                {
                  id: "4",
                  path: "/A/1/2",
                  type: "department",
                  label: "A",
                  children: [],
                },
                {
                  id: "3",
                  path: "/A/1/1",
                  type: "department",
                  label: "D",
                  children: [],
                },
              ],
            },
          ],
        },
      ];

      expect(orgTreeService.sortOrgTree(input)).toEqual(expected);
    });
  });
});
