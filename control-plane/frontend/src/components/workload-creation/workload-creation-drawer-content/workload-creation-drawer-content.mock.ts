import type { IDrawerListItem } from "@/models/workload-creation-drawer-model";

export const workloadCreationDrawerContenatMock = {
  items: [
    {
      name: "chatbot-ui",
      id: "715a51c3-93ce-4838-a6c6-4d248e6b9ba0",
      kind: "environment",
      description: "",
      icons: [
        {
          name: "ChatbotUI",
          icon: "chatbotui",
        },
        {
          name: "ServingPort",
          icon: "custom-tool",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T08:34:57.923786Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "Image URL",
          content: ["registry.example.com/runai/llm-app"],
        },
        {
          title: "Image pull policy",
          content: ["Pull the image only if it's not already present on the host"],
        },
        {
          title: "ChatbotUI",
          icon: "chatbotui",
          noSeperator: true,
          content: ["Connection type: ExternalUrl", "Container port: 3000"],
        },
        {
          title: "ServingPort",
          icon: "custom-tool",
          noSeperator: false,
          content: ["Connection type: ServingPort", "Container port: 3000", "Protocol: http"],
        },
        {
          title: "Environment variables",
          content: [
            "Name: RUNAI_MODEL_NAME",
            "Source: Custom",
            "Value: MODEL_NAME_HERE",
            "<br />",
            "Name: RUNAI_MODEL_BASE_URL",
            "Source: Custom",
            "Value: http://WORKLOAD.runai-PROJECT.svc",
            "<br />",
            "Name: RUNAI_MODEL_MAX_LENGTH",
            "Source: Custom",
            "Value: 4096",
            "<br />",
            "Name: RUNAI_MODEL_TOKEN_LIMIT",
            "Source: Custom",
            "Value: 2048",
            "<br />",
            "Name: PATH_PREFIX",
            "Source: Custom",
            "Value: /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
            "<br />",
          ],
        },
      ],
    },
    {
      name: "tensorboard",
      id: "702efc41-9107-4495-b480-1e6c5547aa92",
      kind: "environment",
      description: "",
      icons: [
        {
          name: "TensorBoard",
          icon: "tensorboard",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.389022Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "Image URL",
          content: ["tensorflow/tensorflow:latest"],
        },
        {
          title: "Image pull policy",
          content: ["Pull the image only if it's not already present on the host"],
        },
        {
          title: "TensorBoard",
          icon: "tensorboard",
          noSeperator: false,
          content: ["Connection type: ExternalUrl", "Container port: 6006"],
        },
        {
          title: "Command and arguments",
          content: [
            "Command: tensorboard",
            "Arguments: --logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
          ],
        },
        {
          title: "Image pull policy",
          content: ["Defined by the image"],
        },
      ],
    },
    {
      name: "jupyter-lab",
      id: "27b00702-a224-4cdc-9fd0-01bd4c2c9594",
      kind: "environment",
      description: "",
      icons: [
        {
          name: "Jupyter",
          icon: "jupyter",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.391437Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "Image URL",
          content: ["jupyter/scipy-notebook"],
        },
        {
          title: "Image pull policy",
          content: ["Pull the image only if it's not already present on the host"],
        },
        {
          title: "Jupyter",
          icon: "jupyter",
          noSeperator: false,
          content: ["Connection type: ExternalUrl", "Container port: 8888"],
        },
        {
          title: "Command and arguments",
          content: [
            "Command: start-notebook.sh",
            "Arguments: --NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
          ],
        },
        {
          title: "Image pull policy",
          content: ["Defined by the image"],
        },
      ],
    },
    {
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
    },
  ] as IDrawerListItem[],
};
