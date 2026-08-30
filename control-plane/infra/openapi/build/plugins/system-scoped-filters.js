

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
  
                const parentTags = new Set();
                if (Array.isArray(root.tags)) {
                  root.tags.forEach(tagObj => {
                    if (tagObj['x-parent']) {
                      parentTags.add(tagObj['x-parent']);
                    }
                  });
                }
  
                root.tags = root.tags.filter(tagObj => usedTags.has(tagObj.name) || parentTags.has(tagObj.name));
                const validTagNames = new Set(root.tags.map(t => t.name));
  
                // For each operation, cleanup its x-tags to leave only valid global tags
                if (Array.isArray(root['x-tagGroups'])) {
                    root['x-tagGroups'] = root['x-tagGroups']
                      .map(group => ({
                        ...group,
                        tags: group.tags.filter(tag => usedTags.has(tag))
                      }))
                      .filter(group => group.tags.length > 0);
                  
                    if (root['x-tagGroups'].length === 0) {
                      delete root['x-tagGroups'];
                    }
                  }
              }
            }
          }),
     }
    }
  };