import { describe, vi, test, expect } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import RolePermissions from "@/views/rbac/roles/role-permissions.vue";
import { Quasar } from "quasar";
import type { TestingPinia } from "@pinia/testing";
import { createTestingPinia } from "@pinia/testing";
import type { RoleV1 } from "@/swagger-models/authorization-client";
import { ResourceType } from "@/swagger-models/authorization-client";
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const systemAdmin: RoleV1 = {
  id: 1,
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
  custom: false,
  effectiveEnabled: true,
  name: "System administrator",
  description:
    "A system administrator can:\nControl all aspects of the system: users, organizations, compute resources, general settings, etc.\nThis role has global system control and should be limited to a small group of skilled IT administrators.",
  permissions: [
    { resourceType: "users", displayName: "Users", groupId: "iam", actions: ["create", "read", "update", "delete"] },
    {
      resourceType: "apps",
      displayName: "Service accounts",
      groupId: "iam",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "access_rules",
      displayName: "Access rules",
      groupId: "iam",
      actions: ["create", "read", "update", "delete"],
    },
    { resourceType: "roles", displayName: "Roles", groupId: "iam", actions: ["read"] },
    {
      resourceType: "department",
      displayName: "Departments",
      groupId: "organization",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "project",
      displayName: "Projects",
      groupId: "organization",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "workloads",
      displayName: "Workloads",
      groupId: "workload",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "credentials",
      displayName: "Credentials",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "workspaces",
      displayName: "Workspaces",
      groupId: "workload",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "trainings",
      displayName: "Trainings",
      groupId: "workload",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "inferences",
      displayName: "Inferences",
      groupId: "workload",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "environments",
      displayName: "Environments",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "host-path-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "nfs-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "pvc-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "cm-volume-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "secret-volume-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "git-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "s3-assets",
      displayName: "Data sources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "compute-resources",
      displayName: "Compute resources",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "templates",
      displayName: "Templates",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "cluster",
      displayName: "Clusters",
      groupId: "physical-resource",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "nodepools",
      displayName: "Node pools",
      groupId: "physical-resource",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "settings",
      displayName: "Settings",
      groupId: "organization",
      actions: ["create", "read", "update", "delete"],
    },
    { resourceType: "events-history", displayName: "Event history", groupId: "organization", actions: ["read"] },
    {
      resourceType: "policies",
      displayName: "Policies",
      groupId: "organization",
      actions: ["create", "read", "update", "delete"],
    },
    {
      resourceType: "datavolumes",
      displayName: "Data volumes",
      groupId: "workload-asset",
      actions: ["create", "read", "update", "delete"],
    },
    { resourceType: "tenant", displayName: "Account", groupId: "organization", actions: ["read"] },
    { resourceType: "nodes", displayName: "Nodes", groupId: "physical-resource", actions: ["read"] },
    {
      resourceType: "cluster-config",
      displayName: "Cluster configuration",
      groupId: "physical-resource",
      actions: ["create", "read", "update", "delete"],
    },
  ],
};
describe("role-permission.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof RolePermissions>>;
  const mountComponent = (props: { role: RoleV1 }): void => {
    wrapper = mount(RolePermissions, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
      },
      props: props,
    });
  };

  describe("RolePermissions component", () => {
    test("should render all displayName text", async () => {
      mountComponent({ role: systemAdmin });
      wrapper.vm.groupAndSortPermissions.forEach((permission) => {
        const resourceTypeElement = wrapper.find(`[aid="${permission.resourceType}"]`);
        expect(resourceTypeElement.text()).eql(permission.displayName);
      });
    });
    test("validate that all data sources are under the data sources category", async () => {
      mountComponent({ role: systemAdmin });
      const resourceTypeElement = wrapper.find(`[aid="${ResourceType.HostPathAssets}"]`);
      expect(resourceTypeElement.text()).eql("Data sources");

      expect(wrapper.findAll(`[aid="${ResourceType.S3Assets}"]`).length).eql(0);
      expect(wrapper.findAll(`[aid="${ResourceType.NfsAssets}"]`).length).eql(0);
      expect(wrapper.findAll(`[aid="${ResourceType.GitAssets}"]`).length).eql(0);
      expect(wrapper.findAll(`[aid="${ResourceType.CmVolumeAssets}"]`).length).eql(0);
      expect(wrapper.findAll(`[aid="${ResourceType.SecretVolumeAssets}"]`).length).eql(0);
    });
    test("should display PVC separately from data sources", async () => {
      mountComponent({ role: systemAdmin });
      const pvcElement = wrapper.find(`[aid="${ResourceType.PvcAssets}"]`);
      expect(pvcElement.exists()).toBe(true);
      expect(pvcElement.text()).eql("PVC");
    });
    test("should display Workspaces and Trainings permissions", async () => {
      mountComponent({ role: systemAdmin });
      const workspacesElement = wrapper.find(`[aid="${ResourceType.Workspaces}"]`);
      const trainingsElement = wrapper.find(`[aid="${ResourceType.Trainings}"]`);

      expect(workspacesElement.exists()).toBe(true);
      expect(workspacesElement.text()).eql("Workspaces");
      expect(trainingsElement.exists()).toBe(true);
      expect(trainingsElement.text()).eql("Trainings");
    });
  });
});
