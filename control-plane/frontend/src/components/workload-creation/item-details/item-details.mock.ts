import type { IDrawerListItem } from "@/models/workload-creation-drawer-model";

export const itemDetailsMock: IDrawerListItem = {
  name: "jupyter-tensorboard",
  id: "ef7a9843-1640-4f4f-97ea-1f2b35fe7042",
  kind: "environment",
  description: "",
  icons: [
    {
      name: "Jupyter",
      icon: "jupyter",
    },
    {
      name: "TensorBoard",
      icon: "tensorboard",
    },
  ],
  policy: {
    compliance: true,
  },
  createdAt: "2024-12-23T09:29:09.393236Z",
  lastUsed: "",
  disabled: false,
  details: [
    {
      title: "Image URL",
      content: ["gcr.io/run-ai-demo/jupyter-tensorboard"],
    },
    {
      title: "Image pull policy",
      content: ["Pull the image only if it's not already present on the host"],
    },
    {
      title: "Jupyter",
      icon: "jupyter",
      noSeperator: true,
      content: ["Connection type: ExternalUrl", "Container port: 8888"],
    },
    {
      title: "TensorBoard",
      icon: "tensorboard",
      noSeperator: false,
      content: ["Connection type: ExternalUrl", "Container port: 6006"],
    },
    {
      title: "Environment variables",
      content: ["Name: LOG_DIR", "Source: Custom", "Value: ./", "<br />"],
    },
    {
      title: "Image pull policy",
      content: ["Defined by the image"],
    },
  ],
};
