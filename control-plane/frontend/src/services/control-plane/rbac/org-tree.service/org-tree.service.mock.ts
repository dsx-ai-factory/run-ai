import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";
import type { IOrgTreeNode } from "@/models/org-tree.model";

const orgUnits: Array<OrgUnit> = [
  {
    id: "1001",
    type: "tenant",
    name: "runai",
  },
  {
    id: "8a483443-8cc1-4d29-bd59-0e4e2c9082f0",
    type: "cluster",
    name: "tom-1-1695017060",
    parentId: "1001",
    parentType: "tenant",
  },
  {
    id: "1",
    type: "department",
    name: "default",
    parentId: "8a483443-8cc1-4d29-bd59-0e4e2c9082f0",
    parentType: "cluster",
  },
  {
    id: "25",
    type: "project",
    name: "p-1",
    parentId: "1",
    parentType: "department",
  },
];

const orgTree: IOrgTreeNode = {
  id: "1001",
  path: "runai",
  type: ScopeType.Tenant,
  label: "runai",
  icon: "ph-fill ph-network",
  children: [
    {
      id: "8a483443-8cc1-4d29-bd59-0e4e2c9082f0",
      path: "runai/tom-1-1695017060",
      type: ScopeType.Cluster,
      label: "tom-1-1695017060",
      icon: "runai-filled-cluster",
      children: [
        {
          id: "1",
          path: "runai/tom-1-1695017060/default",
          type: ScopeType.Department,
          label: "default",
          icon: "ph-fill ph-garage",
          children: [
            {
              id: "25",
              path: "runai/tom-1-1695017060/default/p-1",
              type: ScopeType.Project,
              label: "p-1",
              icon: "ph-fill ph-suitcase",
              children: [],
            },
          ],
        },
      ],
    },
  ],
};

const richOrgTree: IOrgTreeNode[] = [
  {
    id: "1001",
    path: "runai",
    type: "tenant",
    children: [
      {
        id: "60cfdc86-3c1e-43ff-a781-8080ea6b049e",
        path: "runai/barak-1-1717314378",
        type: "cluster",
        children: [
          {
            id: "1",
            path: "runai/barak-1-1717314378/default",
            type: "department",
            children: [
              {
                id: "1",
                path: "runai/barak-1-1717314378/default/p-0",
                type: "project",
                children: [],
                icon: "ph-fill ph-suitcase",
                label: "p-0",
              },
              {
                id: "2",
                path: "runai/barak-1-1717314378/default/test",
                type: "project",
                children: [],
                icon: "ph-fill ph-suitcase",
                label: "test",
              },
            ],
            icon: "ph-fill ph-garage",
            label: "default",
          },
        ],
        icon: "runai-filled-cluster",
        label: "barak-1-1717314378",
      },
      {
        id: "57d75727-a092-4b27-8276-10a00b9ace46",
        path: "runai/cluster-1",
        type: "cluster",
        children: [
          {
            id: "11",
            path: "runai/cluster-1/default",
            type: "department",
            children: [],
            icon: "ph-fill ph-garage",
            label: "default",
          },
        ],
        icon: "runai-filled-cluster",
        label: "cluster-1",
      },
      {
        id: "3054cb58-b967-4b4b-b7d9-8ce6e923fab2",
        path: "runai/cluster-2",
        type: "cluster",
        children: [
          {
            id: "12",
            path: "runai/cluster-2/default",
            type: "department",
            children: [],
            icon: "ph-fill ph-garage",
            label: "default",
          },
        ],
        icon: "runai-filled-cluster",
        label: "cluster-2",
      },
    ],
    icon: "ph-fill ph-network",
    label: "runai",
  },
];
export const orgTreeServiceMock = {
  orgUnits,
  orgTree,
  richOrgTree,
};
