/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { OrgUnit, ScopeType } from '../api/apis/backendApi';

export function getTypeAndIdString(id: string, type: ScopeType): string {
    return type + '-' + id;
}

export function validateTree(orgUnits: OrgUnit[]): boolean {
    const nodeMap = new Map<string, OrgUnit>();
    const nodeToChildren = new Map<string, OrgUnit[]>();

    const rootNodes = new Set<OrgUnit>();

    for (const orgUnit of orgUnits) {
        const { id, type, parentId, parentType } = orgUnit;
        const typeAndId = getTypeAndIdString(id, type);

        if (nodeMap.has(typeAndId)) {
            console.error(`Duplicate node found with Type ${type} and Id ${id}`);
            return false;
        }

        nodeMap.set(typeAndId, orgUnit);

        if (!parentId) {
            rootNodes.add(orgUnit);
        } else {
            const parentTypeAndId = getTypeAndIdString(parentId, parentType);
            const children = nodeToChildren.get(parentTypeAndId) ?? [];
            children.push(orgUnit);
            nodeToChildren.set(parentTypeAndId, children);
        }
    }

    // Perform a depth-first traversal to check the tree structure
    function isSubtreeValid(nodeTypeAndId: string): boolean {
        const node = nodeMap.get(nodeTypeAndId);
        if (!node) {
            console.error(`Node with Type-Id ${nodeTypeAndId} not found`);
            return false;
        }

        if (nodeToChildren.has(nodeTypeAndId)) {
            const children = nodeToChildren.get(nodeTypeAndId);

            for (const child of children) {
                if (!isSubtreeValid(getTypeAndIdString(child.id, child.type))) {
                    console.error(`Invalid subtree for child with Type-Id ${nodeTypeAndId}`);
                    return false;
                }
            }
        }

        return true;
    }

    for (const rootNode of rootNodes) {
        if (!isSubtreeValid(getTypeAndIdString(rootNode.id, rootNode.type))) {
            return false;
        }
    }

    return true;
}
