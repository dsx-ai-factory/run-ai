import { describe, it, vi, beforeEach, afterEach, expect } from "vitest";
import { IdleTimer } from "./idle-timer.service";
import { useAuthStore } from "@/stores/auth.store";
import { storageUtil } from "@/utils/storage.util";
import { intervalUtil } from "@/utils/interval.util";
import type { EIntervalLabels } from "@/models/interval.model";
import { DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES } from "@/models/setting.model";

// Mock dependencies
vi.mock("@/stores/auth.store");
vi.mock("@/utils/storage.util");
vi.mock("@/utils/interval.util");

describe("IdleTimer", () => {
  let idleTimer: IdleTimer;
  let eventListeners: Map<string, (event: Event) => void>;

  beforeEach(async () => {
    vi.useFakeTimers();
    vi.clearAllMocks();

    // Reset the singleton instance for testing
    (IdleTimer as unknown as { instance: IdleTimer | null }).instance = null;
    idleTimer = IdleTimer.getInstance();

    // Mock document event listeners
    eventListeners = new Map();
    vi.spyOn(document, "addEventListener").mockImplementation((event, callback) => {
      if (typeof callback === "function") {
        eventListeners.set(event as string, callback as (event: Event) => void);
      }
    });
    vi.spyOn(document, "removeEventListener").mockImplementation((event) => {
      eventListeners.delete(event as string);
    });

    // Mock storage util
    vi.mocked(storageUtil.save).mockImplementation(() => null);
    vi.mocked(storageUtil.remove).mockImplementation(() => null);
    vi.mocked(storageUtil.get).mockImplementation(() => null);

    // Mock interval util
    vi.mocked(intervalUtil.startInterval).mockImplementation((_label: EIntervalLabels, _callback: () => void) => {
      return 1;
    });
    vi.mocked(intervalUtil.stopInterval).mockImplementation(() => null);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
    // Reset the singleton instance for testing
    (IdleTimer as unknown as { instance: IdleTimer | null }).instance = null;
  });

  it("should be a singleton", () => {
    const instance1 = IdleTimer.getInstance();
    const instance2 = IdleTimer.getInstance();
    expect(instance1).toBe(instance2);
  });

  it("should set timeout when track is called", async () => {
    await idleTimer.track(10);
    expect(idleTimer.timeout).toBe(10 * 60 * 1000);
  });

  it("should update timeout when updateTimeout is called", async () => {
    await idleTimer.track(10);
    await idleTimer.updateTimeout(20);
    expect(idleTimer.timeout).toBe(20 * 60 * 1000);
  });

  it("should cleanup and not track when timeout is 0 (disabled)", async () => {
    await idleTimer.track(DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES);

    idleTimer.events.forEach((event) => {
      expect(eventListeners.has(event)).toBe(false);
    });

    expect(intervalUtil.stopInterval).toHaveBeenCalledWith("idle-timer");
    expect(storageUtil.remove).toHaveBeenCalledWith(idleTimer.expiredTimeName);
  });

  it("should disable tracking when updateTimeout is called with 0", async () => {
    await idleTimer.track(10);
    expect(idleTimer.timeout).toBe(10 * 60 * 1000);

    await idleTimer.updateTimeout(DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES);

    idleTimer.events.forEach((event) => {
      expect(eventListeners.has(event)).toBe(false);
    });

    expect(intervalUtil.stopInterval).toHaveBeenCalledWith("idle-timer");
  });

  it("should call logout when onTimeout is called", () => {
    const mockStore = {
      logout: vi.fn(),
      $state: {},
      $patch: vi.fn(),
      $reset: vi.fn(),
      $subscribe: vi.fn(),
      $onAction: vi.fn(),
    } as unknown as ReturnType<typeof useAuthStore>;
    vi.mocked(useAuthStore).mockReturnValue(mockStore);

    idleTimer.onTimeout();
    expect(mockStore.logout).toHaveBeenCalled();
  });

  it("should have correct event list", () => {
    expect(idleTimer.events).toEqual(["mousemove", "scroll", "keydown", "click"]);
  });

  it("should add event listeners when track is called", async () => {
    await idleTimer.track(10);

    idleTimer.events.forEach((event) => {
      expect(eventListeners.has(event)).toBe(true);
    });
  });

  it("should remove event listeners when cleanUp is called through onTimeout", async () => {
    await idleTimer.track(10);
    idleTimer.onTimeout();

    idleTimer.events.forEach((event) => {
      expect(eventListeners.has(event)).toBe(false);
    });
  });

  it("should start interval when track is called", async () => {
    await idleTimer.track(10);
    expect(intervalUtil.startInterval).toHaveBeenCalledWith("idle-timer", expect.any(Function));
  });

  it("should stop interval when cleanUp is called through onTimeout", async () => {
    await idleTimer.track(10);
    idleTimer.onTimeout();
    expect(intervalUtil.stopInterval).toHaveBeenCalledWith("idle-timer");
  });

  it("should update storage when events occur", async () => {
    await idleTimer.track(10);

    // Simulate a mousemove event
    const event = new MouseEvent("mousemove");
    const callback = eventListeners.get("mousemove");
    if (callback) {
      callback(event);
    }

    // Wait for the debounce to complete (300ms + 1000ms for safety)
    vi.advanceTimersByTime(1300);

    expect(storageUtil.save).toHaveBeenCalledWith(idleTimer.expiredTimeName, expect.any(String));
  });

  it("should remove storage when cleanUp is called through onTimeout", async () => {
    await idleTimer.track(10);
    idleTimer.onTimeout();
    expect(storageUtil.remove).toHaveBeenCalledWith(idleTimer.expiredTimeName);
  });
});
