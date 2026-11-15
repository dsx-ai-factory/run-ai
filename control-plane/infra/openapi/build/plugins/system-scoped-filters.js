/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */


module.exports = {
    id: 'system-scoped-filters',
    decorators: {
      oas3: {
        "filter-system-scoped-paths": () => ({
            Paths: {
              leave(paths) {
                function hasSystemScoped(obj) {
                  if (obj == null || typeof obj !== "object") return false;
                  if (obj["x-scope"] !== undefined) return true;
                  return Object.values(obj).some(hasSystemScoped);
                }
                for (const pathName of Object.keys(paths)) {
                  if (!hasSystemScoped(paths[pathName])) {
                    delete paths[pathName];
                  }
                }
              },
            },
          }),
        "remove-unused-tags-and-clean-x-tags": () => ({
            Root: {
              leave(root) {
                if (!root.tags || !root.paths) return;
  
                const usedTags = new Set();
                Object.values(root.paths).forEach(pathItem => {
                  Object.values(pathItem).forEach(operation => {
                    if (operation && Array.isArray(operation.tags)) {
                      operation.tags.forEach(tag => usedTags.add(tag));
                    }
                  });
                });
  
                root.tags = root.tags.filter(tagObj => usedTags.has(tagObj.name));
              }
            }
          }),
     }
    }
  };