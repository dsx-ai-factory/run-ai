/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeAll, beforeEach, vi } from "vitest";
import { config } from "@vue/test-utils";
import { createApp } from "vue";
import { Quasar } from "quasar";
import axios from "axios";
import type { ComponentPublicInstance } from "vue";
import { VueDOMPurifyHTML, VueDOMPurifyHTMLOptions } from "@/plugins/vue-dom-purify-html";
import { iconMapFn } from "@/plugins/icon-map-fn";

// Register the v-dompurify-html directive globally (as in production) so any
// component using it resolves the directive during tests.
config.global.plugins = [...(config.global.plugins ?? []), [VueDOMPurifyHTML, VueDOMPurifyHTMLOptions]];

// Seed Quasar's iconMapFn (as in production via main.ts) before any test mounts
// Quasar. Quasar freezes its global config on the first install, so this throwaway
// install seeds iconMapFn into the shared config; every per-test `plugins: [Quasar]`
// install then resolves it. Without this, q-icon renders Phosphor names like
// "ph-bold ph-info" as ligature text instead of applying them as CSS classes,
// which breaks icon-class assertions and leaks the name into `.text()`.
createApp({}).use(Quasar, { config: { iconMapFn } });

vi.mock("@/plugins/echarts", () => ({}));

vi.mock("vue-echarts", async () => {
  const { defineComponent, h } = await import("vue");
  return {
    default: defineComponent({
      name: "VChart",
      props: { option: { type: Object, default: () => ({}) }, autoresize: Boolean },
      emits: ["click"],
      setup(props) {
        return () => h("div", { class: "mock-echarts", "data-option": JSON.stringify(props.option) });
      },
    }),
  };
});

const mockAxios = () => {
  axios.get = vi.fn();
  axios.post = vi.fn();
  axios.put = vi.fn();
  axios.delete = vi.fn();
};

function disableScrollTo(): void {
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  Object.defineProperty(window, "scrollTo", { value: () => {}, writable: true });
}

// Returns the component name from either options API ($options.name) or composition API ($.type.name).
// Used to identify which component triggered a Vue warning or error in the test output.
/* eslint-disable @typescript-eslint/no-explicit-any */
function getComponentName(instance: ComponentPublicInstance | null): string {
  if (!instance) return "unknown component";
  return (instance as any).$options?.name ?? (instance as any).$?.type?.name ?? "anonymous component";
}
/* eslint-enable @typescript-eslint/no-explicit-any */

// jsdom doesn't implement HTMLCanvasElement.getContext; mock it to prevent xterm.js stack traces.
HTMLCanvasElement.prototype.getContext = vi.fn() as never;

// Network-related strings emitted by ErrorAlert when a component handles an API failure.
// These are expected in error-path tests and suppressed to avoid noise; tests that want to
// assert on them should spy on console.error locally.
function isSuppressedConsoleError(error: unknown): boolean {
  if (typeof error === "string" && error.startsWith("Anchor:")) return true;
  const text = typeof error === "string" ? error : error instanceof Error ? error.message : String(error);
  return (
    text.includes("Error: connect ECONNREFUSED") || text.includes("Network Error") || text.includes("[status-code: -1]")
  );
}

// V8's default stackTraceLimit is 10. Vue internals consume all 10 frames before reaching
// the test file, so we temporarily increase the limit to capture the full chain.
function captureDeepError(message: string): Error {
  const prev = Error.stackTraceLimit;
  Error.stackTraceLimit = 50;
  const err = new Error(message);
  Error.stackTraceLimit = prev;
  return err;
}

// Strips node_modules and setup.ts frames so vitest points at the test file line
// that triggered the mount/interaction, not at Vue internals or this setup.
function toTestStack(err: Error): Error {
  if (!err.stack) return err;
  const [message, ...frames] = err.stack.split("\n");
  const testFrames = frames.filter((line) => !line.includes("/node_modules/") && !line.includes("unit-test/setup.ts"));
  err.stack = [message, ...testFrames].join("\n");
  return err;
}

beforeAll(() => {
  disableScrollTo();
  mockAxios();
  config.global.stubs = {
    ...config.global.stubs,
    "v-chart": { template: "<div />" },
  };
});

let warnError: Error | null;
let errorMsg: unknown;
let logMsg: string;
let consoleMock: ReturnType<typeof vi.spyOn>;
let consoleMockWarn: ReturnType<typeof vi.spyOn>;
let consoleMockLog: ReturnType<typeof vi.spyOn>;

beforeEach(() => {
  warnError = null;
  errorMsg = "";
  logMsg = "";

  config.global.config.warnHandler = (warn: string, instance: ComponentPublicInstance | null) => {
    // QTooltip emits "Anchor: target not found" in jsdom because it cannot resolve
    // DOM position in a detached test environment. Suppress it; tests that assert
    // tooltip presence/text do so via DOM assertions, not positioning behaviour.
    if (warn.includes("Anchor: target not found")) return;
    const name = getComponentName(instance);
    // Create the Error here (while the test call is still on the stack) then strip
    // node_modules + setup frames so vitest points at the test file line.
    warnError = toTestStack(captureDeepError(`[Vue warn] in <${name}>: ${warn}`));
    process.stderr.write(`[Vue warn] in <${name}>: ${warn}\n`);
  };

  // `info` is the lifecycle hook or event name where the error was caught
  // (e.g. "created hook", "v-on handler"), which helps locate the source quickly.
  config.global.config.errorHandler = (error: unknown, instance: ComponentPublicInstance | null, info: string) => {
    const name = getComponentName(instance);
    if (error instanceof Error) {
      errorMsg = toTestStack(error);
    } else {
      errorMsg = toTestStack(captureDeepError(`[Vue error] in <${name}> (${info}): ${String(error)}`));
    }
    process.stderr.write(`[Vue error] in <${name}> (${info}): ${String(error)}\n`);
  };

  // Suppress console.warn from production code; Vue-level warnings are captured above via warnHandler.
  consoleMockWarn = vi.spyOn(console, "warn").mockImplementation(() => {});

  // Capture console.error calls. Expected network-failure strings from ErrorAlert are suppressed;
  // anything else is stored and will fail the test in afterEach — this catches missing service mocks
  // that let XHR errors propagate silently.
  consoleMock = vi.spyOn(console, "error").mockImplementation((error: unknown) => {
    if (isSuppressedConsoleError(error)) return;
    errorMsg = error;
  });

  // Any console.log in production code during a test is unexpected and will fail the test.
  // Tests that intentionally call console.log should spy on it locally.
  consoleMockLog = vi.spyOn(console, "log").mockImplementation((...args: unknown[]) => {
    logMsg = args.map(String).join(" ");
  });
});

afterEach(() => {
  if (warnError) throw warnError;

  if (errorMsg) {
    if (errorMsg instanceof Error) throw errorMsg;
    throw new Error(`[Vue error]: ${String(errorMsg)}`);
  }

  if (logMsg) {
    throw new Error(
      `Unexpected console.log: "${logMsg}". Mock console in your test with vi.spyOn(console, 'log').mockImplementation(() => {})`,
    );
  }

  consoleMock?.mockRestore();
  consoleMockWarn?.mockRestore();
  consoleMockLog?.mockRestore();
});
