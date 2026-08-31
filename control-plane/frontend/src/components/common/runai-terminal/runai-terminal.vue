<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="terminal-container">
    <div ref="terminalElement" class="terminal-content"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from "vue";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";

export interface Props {
  websocket: WebSocket;
}

const props = defineProps<Props>();

// Local ref for websocket to avoid mutating props
const ws = ref<WebSocket>();

const terminalElement = ref<HTMLElement>();
const terminal = ref<Terminal>();
const fitAddon = ref<FitAddon>();
let resizeTimeout: ReturnType<typeof setTimeout> | null = null;
let initialResizeTimeout: ReturnType<typeof setTimeout> | null = null;
let containerResizeTimeout: ReturnType<typeof setTimeout> | null = null;
let lastSentSize = { cols: 0, rows: 0 }; // Track last sent size to avoid redundant updates
let resizeObserver: ResizeObserver | null = null;
let isFitting = false; // Flag to prevent resize events during fit operations

const initTerminal = () => {
  if (!terminalElement.value) {
    return;
  }

  // Create terminal
  terminal.value = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
    theme: {
      background: "#1e1e1e",
      foreground: "#d4d4d4",
    },
  });

  // Load fit addon
  fitAddon.value = new FitAddon();
  terminal.value.loadAddon(fitAddon.value);

  // Open terminal
  terminal.value.open(terminalElement.value);
  try {
    isFitting = true;
    fitAddon.value.fit();
  } finally {
    isFitting = false;
  }
  terminal.value.focus();

  // Handle user input - send as plain binary (no SPDY prefix needed for this endpoint)
  terminal.value.onData((data) => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      const bytes = new TextEncoder().encode(data);
      ws.value.send(bytes.buffer);
    }
  });

  // Handle terminal resize - notify server (debounced to avoid spam)
  terminal.value.onResize(({ cols, rows }) => {
    // Skip if we're actively fitting (to avoid echo during container resize)
    if (isFitting) {
      return;
    }

    // Clear previous timeout
    if (resizeTimeout) {
      clearTimeout(resizeTimeout);
    }

    // Send resize after 300ms of no more resize events
    resizeTimeout = setTimeout(() => {
      sendResize(cols, rows);
      resizeTimeout = null;
    }, 300);
  });

  // Handle incoming messages
  setupWebSocket();
};

const sendResize = (cols: number, rows: number) => {
  // Only send if size actually changed
  if (cols === lastSentSize.cols && rows === lastSentSize.rows) {
    return;
  }

  if (ws.value?.readyState === WebSocket.OPEN) {
    // Server expects a structured message for resize
    const resizeMessage = {
      type: "terminal_size",
      params: {
        width: cols,
        height: rows,
      },
    };

    // Send as TEXT (not binary)
    ws.value.send(JSON.stringify(resizeMessage));

    // Update last sent size
    lastSentSize = { cols, rows };
  }
};

const setupWebSocket = () => {
  if (!ws.value || !terminal.value) {
    return;
  }

  // Reset size tracking for new connection
  lastSentSize = { cols: 0, rows: 0 };

  // Send initial size when WebSocket is ready
  const sendInitialSize = () => {
    if (terminal.value) {
      // Clear any existing timeout
      if (initialResizeTimeout) {
        clearTimeout(initialResizeTimeout);
      }
      // Small delay to let bash initialize after connection
      initialResizeTimeout = setTimeout(() => {
        if (terminal.value) {
          sendResize(terminal.value.cols, terminal.value.rows);
        }
        initialResizeTimeout = null;
      }, 300);
    }
  };

  // If already open, send size
  if (ws.value.readyState === WebSocket.OPEN) {
    sendInitialSize();
  } else if (ws.value.readyState === WebSocket.CONNECTING) {
    // Wait for open event
    const originalOnOpen = ws.value.onopen;
    ws.value.onopen = (event) => {
      sendInitialSize();
      // Call original handler if it exists
      if (originalOnOpen && ws.value) {
        originalOnOpen.call(ws.value, event);
      }
    };
  }

  ws.value.onmessage = (event) => {
    // Binary messages = terminal I/O
    if (event.data instanceof ArrayBuffer && event.data.byteLength > 0) {
      const text = new TextDecoder().decode(event.data);
      terminal.value?.write(text);
      return;
    }

    // Text messages = structured messages (resize confirmations, errors, etc.)
    if (typeof event.data === "string") {
      try {
        const message = JSON.parse(event.data);
        // Silently handle resize acknowledgments and other messages
        if (message.type !== "terminal_size") {
          console.log("Server message:", message);
        }
      } catch (e) {
        console.warn("Non-JSON text message:", event.data);
      }
    }
  };

  ws.value.onclose = () => {
    terminal.value?.write("\r\n\x1b[31mConnection closed\x1b[0m\r\n");
  };
};

// Handle window resize
const handleResize = () => {
  if (fitAddon.value) {
    try {
      isFitting = true;
      fitAddon.value.fit();
    } finally {
      isFitting = false;
    }
  }
};

onMounted(() => {
  initTerminal();
  window.addEventListener("resize", handleResize);

  // Watch for container size changes (e.g., drawer resize)
  if (terminalElement.value) {
    resizeObserver = new ResizeObserver(() => {
      // Debounce resize to avoid excessive refits
      if (containerResizeTimeout) {
        clearTimeout(containerResizeTimeout);
      }

      containerResizeTimeout = setTimeout(() => {
        if (fitAddon.value) {
          try {
            isFitting = true;
            fitAddon.value.fit();
          } finally {
            isFitting = false;
          }
        }
        containerResizeTimeout = null;
      }, 150);
    });
    resizeObserver.observe(terminalElement.value);
  }
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);

  // Disconnect ResizeObserver
  if (resizeObserver) {
    resizeObserver.disconnect();
    resizeObserver = null;
  }

  // Clear all pending timeouts
  if (resizeTimeout) {
    clearTimeout(resizeTimeout);
    resizeTimeout = null;
  }
  if (initialResizeTimeout) {
    clearTimeout(initialResizeTimeout);
    initialResizeTimeout = null;
  }
  if (containerResizeTimeout) {
    clearTimeout(containerResizeTimeout);
    containerResizeTimeout = null;
  }

  // Safely dispose terminal and its addons
  if (terminal.value) {
    try {
      terminal.value.dispose();
    } catch (error) {
      // Silently ignore disposal errors - component is being destroyed anyway
    }
    terminal.value = undefined;
  }

  // Clear addon refs
  fitAddon.value = undefined;

  // Don't close WebSocket - let parent manage it
});

// Watch for websocket prop changes and update local ref
watch(
  () => props.websocket,
  (newWebsocket) => {
    ws.value = newWebsocket;
    setupWebSocket();
  },
  { immediate: true },
);
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  padding: 8px;
  box-sizing: border-box;
  background-color: #1e1e1e;
}

.terminal-content {
  flex: 1;
  overflow: hidden;
}
</style>

<!-- Unscoped styles for xterm scrollbar - scoped styles don't work well with third-party library internals -->
<style>
.terminal-content .xterm-viewport::-webkit-scrollbar {
  width: 8px;
}

.terminal-content .xterm-viewport::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.3);
  border-radius: 4px;
}
</style>
