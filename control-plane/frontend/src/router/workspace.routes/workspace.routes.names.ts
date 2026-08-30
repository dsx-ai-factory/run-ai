// This file is a workaround for a bug in vite.
// The bug:
// When you import the WORKSPACE_ROUTE_NAMES from the workspace.route.ts file into
// component, during the work you do some changes to the component (or to his parents),
// HMR (used by vite) show an exception about "initialization" and "syntax errors or importing non-existent modules".
// So I moved the names out of the route file, which fix this error.
export const WORKSPACE_ROUTE_NAMES = {
  WORKSPACE_NEW: "workspace-new",
  WORKSPACE_ASSETS_EDIT: "workspace-assets-edit",
  WORKSPACE_HYBRID: "workspace-hybrid",
};
