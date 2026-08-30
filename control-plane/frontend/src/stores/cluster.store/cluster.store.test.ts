import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useClusterStore } from "./cluster.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    create: vi.fn(),
    remove: vi.fn(),
    listMinimal: vi.fn(),
  },
}));

vi.mock("@/stores/auth.store", () => ({
  useAuthStore: () => ({
    loadUserOrgUnits: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock("@/services/cluster/cluster-api.service/cluster-api-old.service", () => ({
  clusterApiServiceOld: {
    setBaseURL: vi.fn(),
  },
}));

vi.mock("@/services/infra/client-apis/cluster-api/cluster-api", () => ({
  clusterApi: {
    setBaseUrl: vi.fn(),
  },
}));

const mockNewCluster: DisplayedCluster = {
  uuid: "new-cluster-uuid",
  name: "new-cluster",
  tenantId: 1,
  createdAt: "2024-01-01T00:00:00.000Z",
  version: "2.20.0",
  domain: "https://new-cluster.example.com",
};

describe("ClusterStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  describe("create", () => {
    it("should add the new cluster to the clusters array immediately after creation", async () => {
      vi.mocked(clusterService.create).mockResolvedValue(mockNewCluster);
      const store = useClusterStore();

      expect(store.clusters).toHaveLength(0);

      const result = await store.create("new-cluster");

      expect(store.clusters).toHaveLength(1);
      expect(store.clusters[0].uuid).toBe("new-cluster-uuid");
      expect(store.clusters[0].name).toBe("new-cluster");
      expect(result).toEqual(mockNewCluster);
    });

    it("should set lastCreatedClusterUuid after creation", async () => {
      vi.mocked(clusterService.create).mockResolvedValue(mockNewCluster);
      const store = useClusterStore();

      await store.create("new-cluster");

      expect(store.lastCreatedClusterUuid).toBe("new-cluster-uuid");
    });

    it("should not add to clusters array if create returns null", async () => {
      vi.mocked(clusterService.create).mockResolvedValue(null as unknown as DisplayedCluster);
      const store = useClusterStore();

      await store.create("new-cluster");

      expect(store.clusters).toHaveLength(0);
    });
  });

  describe("deleteCluster", () => {
    it("should remove the cluster from the clusters array", async () => {
      vi.mocked(clusterService.remove).mockResolvedValue(undefined);
      const store = useClusterStore();

      store.clusters = [{ uuid: "cluster-to-delete", name: "test", tenantId: 1 }];
      expect(store.clusters).toHaveLength(1);

      await store.deleteCluster("cluster-to-delete");

      expect(store.clusters).toHaveLength(0);
    });
  });

  describe("isSingleCluster", () => {
    it("should return true when there is only one cluster", () => {
      const store = useClusterStore();
      store.clusters = [{ uuid: "only-cluster", name: "test", tenantId: 1 }];

      expect(store.isSingleCluster).toBe(true);
    });

    it("should return false when there are multiple clusters", () => {
      const store = useClusterStore();
      store.clusters = [
        { uuid: "cluster-1", name: "test1", tenantId: 1 },
        { uuid: "cluster-2", name: "test2", tenantId: 1 },
      ];

      expect(store.isSingleCluster).toBe(false);
    });

    it("should return false after creating a second cluster", async () => {
      vi.mocked(clusterService.create).mockResolvedValue(mockNewCluster);
      const store = useClusterStore();

      store.clusters = [{ uuid: "existing-cluster", name: "existing", tenantId: 1 }];
      expect(store.isSingleCluster).toBe(true);

      await store.create("new-cluster");

      expect(store.isSingleCluster).toBe(false);
      expect(store.clusters).toHaveLength(2);
    });
  });
});
