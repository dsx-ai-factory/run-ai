<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-linear-progress animation-speed="100" indeterminate v-if="loading" />
  <div class="runai-org-tree">
    <q-tree
      ref="tree"
      class="col-12 col-sm-6"
      :nodes="nodes"
      node-key="path"
      :tick-strategy="tickStrategy"
      :ticked="ticked"
      @update:ticked="updateTicked"
      :filter="searchTerm"
      :no-results-label="noResultsLabel"
      :duration="100"
      v-model:expanded="expandedNodes"
    >
      <template v-slot:default-header="prop">
        <div class="row items-center" :aid="prop.node.path">
          <q-icon class="tree-icon" :name="prop.node.icon" size="20px" />
          <span class="text-b-subtitle2">{{ prop.node.label }}</span>
          <runai-tooltip
            v-if="prop.node.tooltip"
            class="tooltip-icon q-ml-xs runai-icon-btn--small"
            :tooltip-text="prop.node.tooltip"
            width="450px"
            size="xs"
            tooltip-position="right"
            icon="ph-bold ph-info"
            :ripple="false"
          />
        </div>
      </template>
    </q-tree>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// stores
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";

// services
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

// models
import type { IOrgTreeNode, IOrgTreeNodeId } from "@/models/org-tree.model";
import type { QTree } from "quasar";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import type { TTreeDisabledScopes, TTreeSupportedVersions } from "@/models/global.model";

export enum EDisabledScopesTooltips {
  DEFAULT = "This scope can't be selected. Try selecting its subordinates.",
  CLUSTER_VERSION = "This scope can't be selected for the current version of the cluster. Try selecting its subordinates.",
  BRANCH_CLUSTER_VERSION = "This scope and its subordinates can't be selected for the current version of the cluster.",
  PERMISSION = "You don't have permissions to select this scope. Try selecting its subordinates.",
  CLUSTER_VERSION_NO_SUBORDINATES = "This scope can't be selected for the current version of the cluster.",
}

export default defineComponent({
  components: { RunaiTooltip },
  emits: { "update:selected": (selected: Array<IOrgTreeNodeId>) => selected },
  props: {
    selected: {
      type: Array as PropType<IOrgTreeNodeId[]>,
      default: [] as IOrgTreeNodeId[],
    },
    readonly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    searchTerm: {
      type: String as PropType<string>,
      default: "",
    },
    noResultsLabel: {
      type: String as PropType<string>,
      default: "No cluster, department, or project match your filters",
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
    },
    scopesSupportedVersions: {
      type: Object as PropType<TTreeSupportedVersions>,
      default: () => ({}),
    },
    disableTenant: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableCluster: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    forcedDisabledScopes: {
      type: Object as PropType<TTreeDisabledScopes>,
      default: () => ({}),
    },
    hiddenScopeTypes: {
      type: [Object, null] as PropType<Record<ScopeType, boolean> | null>,
      default: null,
    },
    tickStrategy: {
      type: String as PropType<"leaf" | "leaf-filtered" | "none" | "strict" | undefined>,
      default: "strict",
    },
    selectedClusterId: {
      type: String as PropType<string>,
      required: false,
    },
    multiple: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      selectedNode: {} as IOrgTreeNodeId,
      ticked: [] as string[],
      lastTicked: [] as string[],
      nodes: [] as IOrgTreeNode[],
      loading: false as boolean,
      clusterStore: useClusterStore(),
      authStore: useAuthStore(),
      expandedAllowedScopes: {} as PermittedScopes,
      expandedNodes: [] as string[],
      selectedNodePaths: [] as string[],
    };
  },
  async created() {
    this.initTreeNodes();
    if (!this.allowedScopes) return;
    this.expandedAllowedScopes = orgTreeService.getExpandedPermittedScopes(
      this.allowedScopes,
      this.authStore.orgUnitList,
    );
  },
  mounted() {
    this.markSelectedNode();
    this.disableNodes(this.nodes);
  },
  methods: {
    async initTreeNodes(): Promise<void> {
      const orgTree: IOrgTreeNode = orgTreeService.getOrgTree(this.authStore.orgUnitList);
      if (this.selectedClusterId) {
        orgTree.children = this.removeOtherClusters(orgTree);
      }
      if (this.hiddenScopeTypes !== null) {
        this.nodes = orgTreeService.filterHiddenScopes([orgTree], this.hiddenScopeTypes);
      } else {
        this.nodes = [orgTree];
      }
      this.nodes = orgTreeService.sortOrgTree(this.nodes);
      this.setDefaultExpandedNodes();
    },
    removeOtherClusters(orgTree: IOrgTreeNode) {
      const relevantClusterId = this.selectedClusterId;
      return orgTree.children.filter((node: IOrgTreeNode) => {
        return node.id === relevantClusterId;
      });
    },
    markSelectedNode(): void {
      if (this.readonly) {
        const rootNode: IOrgTreeNode = !!this.$refs.tree && (this.$refs.tree as QTree).getNodeByKey(this.nodes[0].path);
        this.changeTickable(rootNode, false);
        rootNode.tickable = false;
      }
      if (!this.selected.length) return;
      const paths: string[] = this.selected
        .map((node) => orgTreeService.getNodePathById(this.nodes[0], node))
        .filter((path) => !!path) as string[];
      if (!paths.length) return;

      // Marking the selected node
      this.selectedNodePaths = [...paths];
      this.ticked = [...paths];

      // Disabling the children of the selected node
      const tree = this.$refs.tree as QTree;
      paths.forEach((path) => this.disableNodeChildren(tree.getNodeByKey(path), true));

      // Expanding the path of the selected node
      paths.forEach((path) => this.setPathExpanded(path));
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    updateTicked(ticked: readonly any[]): void {
      const tree: QTree = this.$refs.tree as QTree;

      // When ticking a node
      if (ticked.length && ticked.length > this.ticked.length) {
        const newTickedPath = ticked[ticked.length - 1];

        if (this.multiple) {
          this.ticked.push(newTickedPath);
        } else {
          this.ticked[0] && this.unTickAndEnableChildren(tree.getNodeByKey(this.ticked[0]));
          this.ticked = [newTickedPath];
          this.selectedNodePaths = [];
        }

        this.addSelectedPath(newTickedPath);
        const lastTickedNode: IOrgTreeNode = !!this.$refs.tree && tree.getNodeByKey(newTickedPath);
        if (lastTickedNode?.children) this.disableNodeChildren(lastTickedNode, true);
      } else {
        const untickedNodePaths = this.ticked.filter((path) => !ticked.includes(path));
        untickedNodePaths.forEach((untickedPath) => {
          this.unTickAndEnableChildren(tree.getNodeByKey(untickedPath));
          this.selectedNodePaths = this.selectedNodePaths.filter((path) => path !== untickedPath);
          this.ticked = this.ticked.filter((path) => path !== untickedPath);
        });
      }

      const selectedNodes = this.selectedNodePaths.map((path) => {
        const node = tree.getNodeByKey(path);
        return {
          type: node?.type,
          id: node?.id,
          path: node?.path,
        };
      });
      this.$emit("update:selected", selectedNodes);
    },
    addSelectedPath(path: string): void {
      const subPaths = this.selectedNodePaths.filter((p) => p.includes(path));
      subPaths.forEach((subPath) => {
        this.selectedNodePaths = this.selectedNodePaths.filter((p) => p !== subPath);
      });
      this.selectedNodePaths.push(path);
    },
    disableNodeChildren(node: IOrgTreeNode, shouldTick = false): void {
      const tickedSet = new Set(this.ticked);
      node.children?.forEach((child: IOrgTreeNode) => {
        shouldTick && !tickedSet.has(child.path) && this.ticked.push(child.path);
        child.tickable = false;
        this.disableNodeChildren(child, shouldTick);
      });
    },
    changeTickable(node: IOrgTreeNode, tickable: boolean): void {
      node.children?.forEach((child: IOrgTreeNode) => {
        child.tickable = tickable;
        this.changeTickable(child, tickable);
      });
    },
    unTickAndEnableChildren(node: IOrgTreeNode): void {
      node.children?.forEach((child: IOrgTreeNode) => {
        this.ticked = this.ticked.filter((id) => id !== child.path);
        child.tickable = true;
        this.unTickAndEnableChildren(child);
      });
      this.disableNodes(this.nodes);
    },
    setPathExpanded(path: string): void {
      const subPaths = path.split("/");
      let currentSubPath = "";
      for (const subPath of subPaths) {
        if (subPath) {
          currentSubPath += currentSubPath !== "" ? `/${subPath}` : subPath;
          this.expandedNodes.push(currentSubPath);
        }
      }
    },
    disableNodes(nodes: Array<IOrgTreeNode>): void {
      if (!nodes || !nodes.length || !this.allowedScopes) return;
      nodes.forEach((node: IOrgTreeNode) => {
        const parentClusterUuid = orgTreeService.getParentClusterUuid(this.authStore.orgUnitList, node.id, node.type);

        // Disabling scopes due to special reasons (not related to version or permission)
        if (this.forcedDisabledScopes[node.type]) {
          this.disableNodeWithTooltip(node, EDisabledScopesTooltips.DEFAULT);
          if (node.children?.length) this.disableNodes(node.children);
          return;
        }

        // Disabling due permissions
        if (!this.isPermittedNode(node)) {
          this.disableNodeWithTooltip(node, EDisabledScopesTooltips.PERMISSION);
          if (node.children?.length) this.disableNodes(node.children);
          return;
        }

        // Disabling all non-allowed node scopes by prop or version
        if (node.type === ScopeType.Tenant) {
          if (this.disableTenant) {
            this.disableNodeWithTooltip(node, EDisabledScopesTooltips.DEFAULT);
          } else if (
            this.scopesSupportedVersions.tenant &&
            !this.clusterStore.isAtleastOneClusterSufficientForVersion(this.scopesSupportedVersions.tenant)
          ) {
            this.disableNodeWithTooltip(node, EDisabledScopesTooltips.CLUSTER_VERSION);
          }
        }
        if (node.type === ScopeType.Cluster) {
          if (this.scopesSupportedVersions.cluster || this.disableCluster) {
            const shouldDisableNode =
              (this.scopesSupportedVersions.cluster &&
                !this.clusterStore.isClusterVersionSufficient(
                  parentClusterUuid,
                  this.scopesSupportedVersions.cluster,
                )) ||
              this.disableCluster;
            if (shouldDisableNode) {
              if (this.versionSupportsChildNodes(parentClusterUuid, node)) {
                this.disableNodeWithTooltip(node, EDisabledScopesTooltips.CLUSTER_VERSION);
              } else {
                // Disabling all the branch below
                this.disableNodeWithTooltip(node, EDisabledScopesTooltips.BRANCH_CLUSTER_VERSION);
                this.disableNodeChildren(node);
                return;
              }
            }
          }
        }
        if (node.type === ScopeType.Department && this.scopesSupportedVersions.department) {
          const shouldDisableNode = !this.clusterStore.isClusterVersionSufficient(
            parentClusterUuid,
            this.scopesSupportedVersions.department,
          );
          if (shouldDisableNode) {
            if (this.versionSupportsChildNodes(parentClusterUuid, node)) {
              this.disableNodeWithTooltip(node, EDisabledScopesTooltips.CLUSTER_VERSION);
            } else {
              // Disabling all the branch below
              this.disableNodeWithTooltip(node, EDisabledScopesTooltips.BRANCH_CLUSTER_VERSION);
              this.disableNodeChildren(node);
              return;
            }
          }
        }
        if (node.type === ScopeType.Project && this.scopesSupportedVersions.project) {
          const shouldDisableNode = !this.clusterStore.isClusterVersionSufficient(
            parentClusterUuid,
            this.scopesSupportedVersions.project,
          );
          shouldDisableNode &&
            this.disableNodeWithTooltip(node, EDisabledScopesTooltips.CLUSTER_VERSION_NO_SUBORDINATES);
        }

        if (node.children?.length) this.disableNodes(node.children);
      });
    },
    disableNodeWithTooltip(node: IOrgTreeNode, reason: EDisabledScopesTooltips) {
      node.tickable = false;
      node.tooltip = reason;
    },
    versionSupportsChildNodes(parentClusterId: string | null, node: IOrgTreeNode): boolean {
      if (!this.scopesSupportedVersions.department && !this.scopesSupportedVersions.project) return true;
      const supportDepartmentScope: boolean =
        !this.scopesSupportedVersions.department ||
        this.clusterStore.isClusterVersionSufficient(parentClusterId, this.scopesSupportedVersions.department);
      const supportProjectScope: boolean =
        !this.scopesSupportedVersions.project ||
        this.clusterStore.isClusterVersionSufficient(parentClusterId, this.scopesSupportedVersions.project);

      if (node.type === ScopeType.Cluster) return supportDepartmentScope || supportProjectScope;
      if (node.type === ScopeType.Department) return supportProjectScope;
      return true;
    },
    isPermittedNode(node: IOrgTreeNode): boolean {
      switch (node.type) {
        case ScopeType.Tenant:
          return this.expandedAllowedScopes?.tenants?.includes(node.id) || false;
        case ScopeType.Cluster:
          return this.expandedAllowedScopes?.clusters?.includes(node.id) || false;
        case ScopeType.Department:
          return this.expandedAllowedScopes?.departments?.includes(node.id) || false;
        case ScopeType.Project:
          return this.expandedAllowedScopes?.projects?.includes(node.id) || false;
        default:
          return false;
      }
    },
    expandFilteredNodes(nodes: Array<IOrgTreeNode>, filter: string): void {
      const nodesToExpand = orgTreeService.findMatchingNodePaths(nodes, [], filter);
      if (nodesToExpand.length) {
        this.expandedNodes = nodesToExpand;
      } else {
        // Reseting to the default expanding
        this.setDefaultExpandedNodes();
      }
    },
    setDefaultExpandedNodes(): void {
      if (this.nodes[0]?.path) {
        this.expandedNodes = [this.nodes[0].path];
      } else {
        this.expandedNodes = [];
      }
    },
  },
  watch: {
    searchTerm: {
      handler(newVal: string): void {
        this.expandFilteredNodes(this.nodes, newVal);
      },
      immediate: true,
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-org-tree {
  :deep(.ph-info) {
    font-weight: 900;
    font-size: 11px !important;
  }
  :deep(.q-tree__icon) {
    color: var(--icon-default);
    font-size: 16px !important;
    margin-left: var(--space-2xs);
    margin-right: var(--space-2xs);
  }
  :deep(.q-tree__arrow) {
    padding-right: var(--space-2xs) !important; // overrides Quasar's inline tree arrow spacing
  }
  :deep(.q-tree) {
    padding: var(--space-s);
    text-align: center;
    align-items: center;
    display: grid;

    &:not(:has(.q-tree__node)) {
      padding-top: 26px;
      padding-bottom: 26px;
    }
  }
  :deep(.q-tree__node-header) {
    border-radius: var(--b-border-radius-radius-xs);

    &:hover {
      background: var(--bg-primary-hover);
    }
  }
  :deep(.q-checkbox__bg) {
    border-color: var(--interactive-secondary-default-border);
  }
  :deep(.q-checkbox.disabled .q-checkbox__bg) {
    border-color: var(--interactive-secondary-disabled-border);
  }
  :deep(.q-checkbox .q-checkbox__inner--truthy .q-checkbox__bg),
  :deep(.q-checkbox .q-checkbox__inner--indet .q-checkbox__bg) {
    border-color: currentColor;
  }
  :deep(.tree-icon) {
    color: var(--icon-default);
    margin-left: var(--space-2xs);
    margin-right: var(--space-2xs);
  }
  :deep(.tooltip-icon .q-icon) {
    font-size: 16px;
    color: var(--icon-info);
  }
}
</style>
