/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, beforeEach, afterEach, it, expect, vi, type Mock } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { nextTick } from "vue";

// Types for mocked instances
type MockResizeObserver = {
  observe: Mock;
  disconnect: Mock;
  unobserve: Mock;
};

type MutableWebSocket = WebSocket & {
  readyState: number;
};

// Mock terminal instance for accessing callbacks
const mockTerminalInstance = {
  onDataCallback: null as ((data: string) => void) | null,
  onResizeCallback: null as ((size: { cols: number; rows: number }) => void) | null,
  open: vi.fn(),
  dispose: vi.fn(),
  focus: vi.fn(),
  write: vi.fn(),
  writeln: vi.fn(),
  onData: vi.fn(),
  onResize: vi.fn(),
  loadAddon: vi.fn(),
  cols: 80,
  rows: 24,
  _core: { isDisposed: false },
};

// Mock xterm BEFORE importing component
vi.mock("@xterm/xterm", () => ({
  Terminal: vi.fn(() => mockTerminalInstance),
}));

vi.mock("@xterm/addon-fit", () => ({
  FitAddon: vi.fn(() => ({
    fit: vi.fn(),
    dispose: vi.fn(),
  })),
}));

// Now import after mocks are set up
import RunaiTerminal from "./runai-terminal.vue";

describe("runai-terminal.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiTerminal>>;
  let mockWebSocket: WebSocket;

  beforeEach(() => {
    vi.useFakeTimers();

    // Reset mock terminal instance
    mockTerminalInstance.onDataCallback = null;
    mockTerminalInstance.onResizeCallback = null;
    mockTerminalInstance.open.mockClear();
    mockTerminalInstance.dispose.mockClear();
    mockTerminalInstance.focus.mockClear();
    mockTerminalInstance.write.mockClear();
    mockTerminalInstance.writeln.mockClear();
    mockTerminalInstance.loadAddon.mockClear();

    // Set up callbacks
    mockTerminalInstance.onData.mockImplementation((callback) => {
      mockTerminalInstance.onDataCallback = callback;
    });
    mockTerminalInstance.onResize.mockImplementation((callback) => {
      mockTerminalInstance.onResizeCallback = callback;
    });

    // Create a mock WebSocket
    mockWebSocket = {
      readyState: WebSocket.OPEN,
      send: vi.fn(),
      close: vi.fn(),
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
      binaryType: "arraybuffer",
    } as unknown as WebSocket;

    // Mock ResizeObserver
    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      disconnect: vi.fn(),
      unobserve: vi.fn(),
    }));
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("should mount and initialize terminal", () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    expect(wrapper.exists()).toBe(true);
  });

  it("should create terminal and load fit addon", () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    expect(mockTerminalInstance.loadAddon).toHaveBeenCalled();
    expect(mockTerminalInstance.open).toHaveBeenCalled();
    expect(mockTerminalInstance.focus).toHaveBeenCalled();
  });

  it("should send user input through websocket when connected", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    await nextTick();

    // Simulate user typing
    if (mockTerminalInstance.onDataCallback) {
      mockTerminalInstance.onDataCallback("hello");
    }

    expect(mockWebSocket.send).toHaveBeenCalled();
    const sendMock = mockWebSocket.send as Mock;
    const sentData = sendMock.mock.calls[0][0];
    // Verify it's an ArrayBuffer-like object (has byteLength)
    expect(sentData).toHaveProperty("byteLength");
  });

  it("should not send input when websocket is not open", async () => {
    const closedWebSocket = {
      ...mockWebSocket,
      readyState: WebSocket.CLOSED,
    } as unknown as WebSocket;

    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: closedWebSocket,
      },
    });

    await nextTick();

    if (mockTerminalInstance.onDataCallback) {
      mockTerminalInstance.onDataCallback("hello");
    }

    expect(mockWebSocket.send).not.toHaveBeenCalled();
  });

  it("should handle binary messages from websocket", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    await nextTick();

    const binaryData = new ArrayBuffer(5);
    const view = new Uint8Array(binaryData);
    view.set([104, 101, 108, 108, 111]); // "hello" in bytes

    const mockEvent = {
      data: binaryData,
    } as MessageEvent;

    // Trigger onmessage
    if (mockWebSocket.onmessage) {
      mockWebSocket.onmessage(mockEvent);
    }

    // Should write to terminal
    expect(mockTerminalInstance.write).toHaveBeenCalled();
  });

  it("should handle text messages from websocket", async () => {
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    await nextTick();

    const textMessage = JSON.stringify({ type: "info", message: "test" });
    const mockEvent = {
      data: textMessage,
    } as MessageEvent;

    // Should not throw error
    expect(() => {
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage(mockEvent);
      }
    }).not.toThrow();
    logSpy.mockRestore();
  });

  it("should debounce resize events", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    await nextTick();

    // Clear initial sends
    vi.clearAllTimers();
    mockWebSocket.send = vi.fn();

    // Trigger multiple resize events
    if (mockTerminalInstance.onResizeCallback) {
      mockTerminalInstance.onResizeCallback({ cols: 80, rows: 24 });
      mockTerminalInstance.onResizeCallback({ cols: 90, rows: 30 });
      mockTerminalInstance.onResizeCallback({ cols: 100, rows: 40 });
    }

    // Should not send immediately
    expect(mockWebSocket.send).not.toHaveBeenCalled();

    // Advance timers past debounce delay (300ms)
    vi.advanceTimersByTime(350);

    // Should send only once with final size
    expect(mockWebSocket.send).toHaveBeenCalledTimes(1);
    expect(mockWebSocket.send).toHaveBeenCalledWith(
      JSON.stringify({
        type: "terminal_size",
        params: {
          width: 100,
          height: 40,
        },
      }),
    );
  });

  it("should not send resize when size hasn't changed", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    await nextTick();

    // First resize
    if (mockTerminalInstance.onResizeCallback) {
      mockTerminalInstance.onResizeCallback({ cols: 80, rows: 24 });
    }
    vi.advanceTimersByTime(350);
    expect(mockWebSocket.send).toHaveBeenCalledTimes(1);

    // Same size again
    if (mockTerminalInstance.onResizeCallback) {
      mockTerminalInstance.onResizeCallback({ cols: 80, rows: 24 });
    }
    vi.advanceTimersByTime(350);

    // Should not send again
    expect(mockWebSocket.send).toHaveBeenCalledTimes(1);
  });

  it("should setup ResizeObserver for container resize", () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    expect(global.ResizeObserver).toHaveBeenCalled();
    const observerConstructor = global.ResizeObserver as unknown as Mock;
    const observerInstance = observerConstructor.mock.results[0].value as MockResizeObserver;
    expect(observerInstance.observe).toHaveBeenCalled();
  });

  it("should handle websocket prop changes", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    const newMockWebSocket = {
      ...mockWebSocket,
      readyState: WebSocket.CONNECTING,
    } as unknown as WebSocket;

    await wrapper.setProps({ websocket: newMockWebSocket });

    // Should re-setup websocket
    await nextTick();
    expect(newMockWebSocket.onmessage).toBeDefined();
  });

  it("should clean up on unmount", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    const observerConstructor = global.ResizeObserver as unknown as Mock;
    const observerInstance = observerConstructor.mock.results[0].value as MockResizeObserver;

    wrapper.unmount();

    expect(observerInstance.disconnect).toHaveBeenCalled();
    expect(mockTerminalInstance.dispose).toHaveBeenCalled();
  });

  it("should write connection closed message on websocket close", async () => {
    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: mockWebSocket,
      },
    });

    await nextTick();

    // Trigger onclose
    if (mockWebSocket.onclose) {
      mockWebSocket.onclose({} as CloseEvent);
    }

    expect(mockTerminalInstance.write).toHaveBeenCalledWith("\r\n\x1b[31mConnection closed\x1b[0m\r\n");
  });

  it("should send initial size when websocket is already open", async () => {
    const openWebSocket = {
      ...mockWebSocket,
      readyState: WebSocket.OPEN,
    } as unknown as WebSocket;

    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: openWebSocket,
      },
    });

    await nextTick();

    // Advance past initial size delay (300ms)
    vi.advanceTimersByTime(350);

    expect(openWebSocket.send).toHaveBeenCalledWith(expect.stringContaining("terminal_size"));
  });

  it("should wait for websocket to open before sending initial size", async () => {
    const sendSpy = vi.fn();
    const connectingWebSocket = {
      readyState: WebSocket.CONNECTING,
      send: sendSpy,
      close: vi.fn(),
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
      binaryType: "arraybuffer",
    } as unknown as WebSocket;

    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: connectingWebSocket,
      },
    });

    await nextTick();

    // Should set onopen handler
    expect(connectingWebSocket.onopen).toBeDefined();

    // Update readyState and simulate websocket opening
    const mutableWs = connectingWebSocket as unknown as MutableWebSocket;
    mutableWs.readyState = WebSocket.OPEN;
    if (connectingWebSocket.onopen) {
      connectingWebSocket.onopen({} as Event);
    }

    // Advance past initial size delay
    vi.advanceTimersByTime(350);

    expect(sendSpy).toHaveBeenCalledWith(expect.stringContaining("terminal_size"));
  });

  it("should preserve original onopen handler if it exists", async () => {
    const originalOnOpen = vi.fn();
    const connectingWebSocket = {
      ...mockWebSocket,
      readyState: WebSocket.CONNECTING,
      onopen: originalOnOpen,
    } as unknown as WebSocket;

    wrapper = mount(RunaiTerminal, {
      props: {
        websocket: connectingWebSocket,
      },
    });

    await nextTick();

    // Trigger the new onopen
    if (connectingWebSocket.onopen && connectingWebSocket.onopen !== originalOnOpen) {
      connectingWebSocket.onopen({} as Event);
    }

    expect(originalOnOpen).toHaveBeenCalled();
  });
});
