import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useUserDisplayNameStore } from "./user-display-name.store";
import { userService } from "@/services/control-plane/user.service/user.service";
import type { UserMinimal } from "@/swagger-models/identity-manager-client";

vi.mock("@/services/control-plane/user.service/user.service", () => ({
  userService: {
    resolveUsersByUsernames: vi.fn(),
  },
}));

const flushMicrotasks = (): Promise<void> => new Promise((resolve) => queueMicrotask(resolve));
const userMinimal = (overrides: Partial<UserMinimal> & { username: string }): UserMinimal => ({
  id: `id-${overrides.username}`,
  firstName: null,
  lastName: null,
  ...overrides,
});

describe("UserDisplayNameStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("starts empty", () => {
    const store = useUserDisplayNameStore();
    expect(store.isResolved("alice@x.io")).toBe(false);
    expect(store.isPending("alice@x.io")).toBe(false);
    expect(store.getDisplayLabel("alice@x.io")).toBeNull();
  });

  it("marks pending immediately and fetches on next microtask", async () => {
    let resolveFetch: (users: UserMinimal[]) => void = () => undefined;
    vi.mocked(userService.resolveUsersByUsernames).mockImplementation(
      () => new Promise<UserMinimal[]>((res) => (resolveFetch = res)),
    );

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["alice@x.io"]);

    expect(store.isPending("alice@x.io")).toBe(true);
    expect(store.isResolved("alice@x.io")).toBe(false);

    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledWith(["alice@x.io"], 3000);

    resolveFetch([userMinimal({ username: "alice@x.io", firstName: "Alice", lastName: "Smith" })]);
    await flushMicrotasks();
    await flushMicrotasks();

    expect(store.isPending("alice@x.io")).toBe(false);
    expect(store.isResolved("alice@x.io")).toBe(true);
    expect(store.getDisplayLabel("alice@x.io")).toEqual({
      label: "alice@x.io Alice Smith",
      primaryLabel: "alice@x.io",
      secondaryLabel: "Alice Smith",
    });
  });

  it("coalesces multiple resolveUsernames calls within the same tick into a single batch", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValue([
      userMinimal({ username: "a@x.io", firstName: "A", lastName: "One" }),
      userMinimal({ username: "b@x.io", firstName: "B", lastName: "Two" }),
      userMinimal({ username: "c@x.io" }),
    ]);

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["a@x.io"]);
    store.resolveUsernames(["b@x.io", "c@x.io"]);
    store.resolveUsernames(["a@x.io"]);

    await flushMicrotasks();
    await flushMicrotasks();

    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);
    const arg = vi.mocked(userService.resolveUsersByUsernames).mock.calls[0][0];
    expect([...arg].sort()).toEqual(["a@x.io", "b@x.io", "c@x.io"]);
  });

  it("does not refetch already-resolved usernames", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValue([
      userMinimal({ username: "a@x.io", firstName: "A", lastName: "One" }),
    ]);

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["a@x.io"]);
    await flushMicrotasks();
    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);

    store.resolveUsernames(["a@x.io"]);
    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);
  });

  it("does not refetch in-flight usernames", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockImplementation(() => new Promise<UserMinimal[]>(() => undefined));

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["a@x.io"]);
    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);

    store.resolveUsernames(["a@x.io"]);
    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);
  });

  it("splits requests >BATCH_SIZE into parallel batches of 25", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValue([]);

    const store = useUserDisplayNameStore();
    const names = Array.from({ length: 55 }, (_, i) => `user${i}@x.io`);
    store.resolveUsernames(names);
    await flushMicrotasks();
    await flushMicrotasks();

    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(3);
    const callSizes = vi.mocked(userService.resolveUsersByUsernames).mock.calls.map((c) => c[0].length);
    expect(callSizes.sort((a, b) => a - b)).toEqual([5, 25, 25]);
  });

  it("caches not-found usernames as null and clears pending", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValue([]);

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["ghost@x.io"]);
    await flushMicrotasks();
    await flushMicrotasks();

    expect(store.isResolved("ghost@x.io")).toBe(true);
    expect(store.isPending("ghost@x.io")).toBe(false);
    expect(store.getDisplayLabel("ghost@x.io")).toBeNull();
    expect(store.getDisplayInfo("ghost@x.io")).toBeNull();
  });

  it("on permanent error (403/404), caches the username as null (does not retry)", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockRejectedValue({ response: { status: 403 } });

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["alice@x.io"]);
    await flushMicrotasks();
    await flushMicrotasks();

    expect(store.isPending("alice@x.io")).toBe(false);
    expect(store.isResolved("alice@x.io")).toBe(true);
    expect(store.getDisplayLabel("alice@x.io")).toBeNull();

    store.resolveUsernames(["alice@x.io"]);
    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(1);
  });

  it("on transient error (timeout / 5xx / network), does NOT cache — next resolve retries", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockRejectedValueOnce(new Error("timeout"));

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["alice@x.io"]);
    await flushMicrotasks();
    await flushMicrotasks();

    // pending cleared so the cell exits the skeleton, but NOT resolved so it stays retry-able.
    expect(store.isPending("alice@x.io")).toBe(false);
    expect(store.isResolved("alice@x.io")).toBe(false);
    expect(store.getDisplayLabel("alice@x.io")).toBeNull();

    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValueOnce([
      userMinimal({ username: "alice@x.io", firstName: "Alice", lastName: "Smith" }),
    ]);
    store.resolveUsernames(["alice@x.io"]);
    await flushMicrotasks();
    await flushMicrotasks();

    expect(userService.resolveUsersByUsernames).toHaveBeenCalledTimes(2);
    expect(store.isResolved("alice@x.io")).toBe(true);
    expect(store.getDisplayLabel("alice@x.io")?.secondaryLabel).toBe("Alice Smith");
  });

  it("getDisplayLabel returns username-only when first/last are empty", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValue([
      userMinimal({ username: "noname@x.io", firstName: null, lastName: null }),
    ]);

    const store = useUserDisplayNameStore();
    store.resolveUsernames(["noname@x.io"]);
    await flushMicrotasks();
    await flushMicrotasks();

    expect(store.getDisplayLabel("noname@x.io")).toEqual({
      label: "noname@x.io",
      primaryLabel: "noname@x.io",
    });
  });

  it("ignores falsy usernames", async () => {
    vi.mocked(userService.resolveUsersByUsernames).mockResolvedValue([]);
    const store = useUserDisplayNameStore();
    store.resolveUsernames(["", "valid@x.io"]);
    await flushMicrotasks();
    expect(userService.resolveUsersByUsernames).toHaveBeenCalledWith(["valid@x.io"], 3000);
  });
});
