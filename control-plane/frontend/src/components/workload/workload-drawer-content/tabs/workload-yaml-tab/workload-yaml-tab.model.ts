// What the selected YAML target resolves to (drives which k8s-object API is called).
export enum ETargetKind {
  Workload = "workload",
  Pod = "pod",
  Element = "element",
}

// Selector value encoding: "<kind>" for the workload root, "<kind>:<id>" for a
// specific pod/element. Parsed back via the first separator.
export const KEY_SEPARATOR = ":";

// UI labels / display strings.
export const OBJECT_SELECT_LABEL = "Object";
export const DOWNLOAD_TOOLTIP = "Download YAML";
export const WORKLOAD_OPTION_LABEL = "Workload";
export const POD_OPTION_PREFIX = "Pod";
export const DEFAULT_ELEMENT_KIND = "Element";
export const LABEL_SEPARATOR = ": ";

// Download.
export const DEFAULT_FILENAME = "workload";
export const YAML_FILE_EXTENSION = ".yaml";
export const YAML_MIME_TYPE = "application/yaml";

export enum EYamlTabView {
  // Dropdown = Workload + Pods (side-panel drawer).
  Collapsed = "collapsed",
  // Dropdown is scoped by the structure-map's selected element.
  Expanded = "expanded",
}
