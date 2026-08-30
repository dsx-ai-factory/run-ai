export const AllocatedGpuWidgetTelemetryMock = {
  type: "GPU_ALLOCATION",
  timestamp: "2024-04-01T11:35:00.875659781Z",
  values: [
    {
      value: "89.00",
      groups: [
        {
          key: "ClusterId",
          value: "b13c3cf9-4cfe-47f6-a6ca-15e2950668d7",
        },
      ],
    },
  ],
};

export const mockAllocatedGpuTelemetryFilterByNodepool = {
  type: "GPU_ALLOCATION",
  timestamp: "2024-04-01T11:36:20.708633653Z",
  values: [
    {
      value: "2.00",
      groups: [
        {
          key: "ClusterId",
          value: "b13c3cf9-4cfe-47f6-a6ca-15e2950668d7",
        },
        {
          key: "Type",
          value: "Workspace",
        },
      ],
    },
  ],
};
