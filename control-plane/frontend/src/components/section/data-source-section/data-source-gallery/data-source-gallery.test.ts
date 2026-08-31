/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import DataSourceGallery from "./data-source-gallery.vue";
import { beforeEach, describe, expect, it, test, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { AssetIdAndKind, PolicyInfo, DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import { mockDataSourceGalleryData } from "./data-source-gallery.mock";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// models
import { AssetKind } from "@/swagger-models/assets-service-client";
import { EWorkloadFormType } from "@/models/workload.model";
import { dataSourceUtil } from "@/utils/data-source.util";
import { DATA_SOURCE_TYPE, type TDataSourceKinds } from "@/models/data-source.model";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

interface IDataSourceGalleryProps {
  entityType: EWorkloadFormType;
  loading: boolean;
  dataSources: Array<DatasourceListResponseEntry>;
  selectedDataSources: AssetIdAndKind[];
  clusterId: string;
  allowedCreatedTypes: Set<TDataSourceKinds>;
  cardsMainTitle?: string;
  sectionDisabled?: boolean;
  createdDataSourceId?: string;
  policyRules?: PolicyInfo;
}

describe("data-source-gallery.vue", () => {
  let dataSourceGalleryCmp: VueWrapper<InstanceType<typeof DataSourceGallery>>;
  const clusterId = "123";

  function getMountedComponent(props: IDataSourceGalleryProps): VueWrapper<InstanceType<typeof DataSourceGallery>> {
    return mount(DataSourceGallery, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
      provide: {
        policy: {},
      },
    });
  }

  describe("general tests", () => {
    const props: IDataSourceGalleryProps = {
      entityType: EWorkloadFormType.Workspace,
      loading: false,
      dataSources: mockDataSourceGalleryData.dataSources,
      selectedDataSources: [],
      clusterId: clusterId,
      allowedCreatedTypes: new Set<TDataSourceKinds>([
        AssetKind.Pvc,
        AssetKind.Nfs,
        AssetKind.S3,
        AssetKind.Git,
        AssetKind.HostPath,
        AssetKind.SecretVolume,
        AssetKind.ConfigMap,
      ]),
    };

    test("component is loading properly", () => {
      const dataSourceGalleryCmp = getMountedComponent({
        ...props,
        cardsMainTitle: "My title",
      });
      expect(dataSourceGalleryCmp.html()).toContain("new data source");
      expect(dataSourceGalleryCmp.html()).toContain("My title");
    });

    test("component is the title properly", () => {
      const dataSourceGalleryCmp = getMountedComponent({
        ...props,
        cardsMainTitle: "Data sources example",
      });
      expect(dataSourceGalleryCmp.html()).toContain("new data source");
      expect(dataSourceGalleryCmp.html()).toContain("Data sources example");
    });

    test("the component will emit the updated selected data sources correctly when a new selection is made ", async () => {
      const dataSourceGalleryCmp = getMountedComponent(props);
      const selectedDataSource = mockDataSourceGalleryData.dataSources[0];
      await dataSourceGalleryCmp.find(`[aid=${selectedDataSource.meta.name}-card]`).trigger("click");
      const expectedEmit: AssetIdAndKind[][][] = [
        [
          [
            {
              id: props.dataSources[0].meta.id,
              kind: props.dataSources[0].meta.kind,
            },
          ],
        ],
      ];
      expect(dataSourceGalleryCmp.emitted("datasource-changed")).toEqual(expectedEmit);
    });
  });

  describe("dataSourceCards", () => {
    it("should return the correct data source cards for workspace", () => {
      const props: IDataSourceGalleryProps = {
        entityType: EWorkloadFormType.Workspace,
        loading: false,
        dataSources: mockDataSourceGalleryData.dataSources,
        selectedDataSources: [],
        clusterId: clusterId,
        allowedCreatedTypes: new Set<TDataSourceKinds>([
          AssetKind.Pvc,
          AssetKind.Nfs,
          AssetKind.S3,
          AssetKind.Git,
          AssetKind.HostPath,
          AssetKind.SecretVolume,
          AssetKind.ConfigMap,
        ]),
      };

      dataSourceGalleryCmp = getMountedComponent(props);

      const dataSourceCards = dataSourceGalleryCmp.vm.dataSourceCards;

      const expectedRes = mockDataSourceGalleryData.dataSources.map((dataSource: DatasourceListResponseEntry) => {
        const dataSourceType = dataSourceUtil.isTDataSourceKinds(dataSource.meta.kind)
          ? DATA_SOURCE_TYPE[dataSource.meta.kind]
          : "";

        return {
          id: dataSource.meta.id,
          cardName: "data-source-card",
          data: dataSource,
          searchValues: [dataSource.meta.name, dataSourceType],
          locked: false,
          disabled: false,
          tooltip: "",
          showDisabledInfo: false,
          sortInfo: {
            name: dataSource.meta.name,
            createdAt: dataSource.meta.createdAt,
            recentUsage: undefined,
          },
        };
      });

      expect(dataSourceCards).toEqual(expectedRes);
    });

    it("should return the correct data source cards for inference when cluster only: Pvc, Git and ConfigMap ", () => {
      const props: IDataSourceGalleryProps = {
        entityType: EWorkloadFormType.Inference,
        loading: false,
        dataSources: mockDataSourceGalleryData.dataSources,
        selectedDataSources: [],
        clusterId: clusterId,
        allowedCreatedTypes: new Set<TDataSourceKinds>([
          AssetKind.Pvc,
          AssetKind.Nfs,
          AssetKind.S3,
          AssetKind.Git,
          AssetKind.HostPath,
          AssetKind.SecretVolume,
          AssetKind.ConfigMap,
        ]),
      };

      const clusterStore = useClusterStore(pinia);
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);
      dataSourceGalleryCmp = getMountedComponent(props);

      const dataSourceCards = dataSourceGalleryCmp.vm.dataSourceCards;
      const expectedRes = [
        {
          id: mockDataSourceGalleryData.dataSources[0].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[0],
          searchValues: [mockDataSourceGalleryData.dataSources[0].meta.name, "Host path"],
          locked: false,
          disabled: true,
          tooltip:
            "Inference workloads don't support data sources of type Host path for the current version of the cluster.",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[0].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[0].meta.createdAt,
            recentUsage: undefined,
          },
        },
        {
          id: mockDataSourceGalleryData.dataSources[1].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[1],
          searchValues: [mockDataSourceGalleryData.dataSources[1].meta.name, "NFS"],
          locked: false,
          disabled: true,
          tooltip: "Inference workloads don't support data sources of type NFS for the current version of the cluster.",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[1].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[1].meta.createdAt,
            recentUsage: undefined,
          },
        },
        {
          id: mockDataSourceGalleryData.dataSources[2].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[2],
          searchValues: [mockDataSourceGalleryData.dataSources[2].meta.name, "Git"],
          locked: false,
          disabled: false,
          tooltip: "",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[2].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[2].meta.createdAt,
            recentUsage: undefined,
          },
        },
        {
          id: mockDataSourceGalleryData.dataSources[3].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[3],
          searchValues: [mockDataSourceGalleryData.dataSources[3].meta.name, "S3 Bucket"],
          locked: false,
          disabled: true,
          tooltip:
            "This setup cannot be selected because Inference workload does not support data sources of type S3 Bucket.",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[3].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[3].meta.createdAt,
            recentUsage: undefined,
          },
        },
      ];

      expect(dataSourceCards).toEqual(expectedRes);
    });

    it(`should return the correct data source cards for inference when cluster is supporting nfs and host-path`, () => {
      const props: IDataSourceGalleryProps = {
        entityType: EWorkloadFormType.Inference,
        loading: false,
        dataSources: mockDataSourceGalleryData.dataSources,
        selectedDataSources: [],
        clusterId: clusterId,
        allowedCreatedTypes: new Set<TDataSourceKinds>([
          AssetKind.Pvc,
          AssetKind.Nfs,
          AssetKind.S3,
          AssetKind.Git,
          AssetKind.HostPath,
          AssetKind.SecretVolume,
          AssetKind.ConfigMap,
        ]),
      };

      const clusterStore = useClusterStore(pinia);
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
      dataSourceGalleryCmp = getMountedComponent(props);

      const dataSourceCards = dataSourceGalleryCmp.vm.dataSourceCards;
      const expectedRes = [
        {
          id: mockDataSourceGalleryData.dataSources[0].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[0],
          searchValues: [mockDataSourceGalleryData.dataSources[0].meta.name, "Host path"],
          locked: false,
          disabled: false,
          tooltip: "",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[0].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[0].meta.createdAt,
            recentUsage: undefined,
          },
        },
        {
          id: mockDataSourceGalleryData.dataSources[1].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[1],
          searchValues: [mockDataSourceGalleryData.dataSources[1].meta.name, "NFS"],
          locked: false,
          disabled: false,
          tooltip: "",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[1].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[1].meta.createdAt,
            recentUsage: undefined,
          },
        },
        {
          id: mockDataSourceGalleryData.dataSources[2].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[2],
          searchValues: [mockDataSourceGalleryData.dataSources[2].meta.name, "Git"],
          locked: false,
          disabled: false,
          tooltip: "",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[2].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[2].meta.createdAt,
            recentUsage: undefined,
          },
        },
        {
          id: mockDataSourceGalleryData.dataSources[3].meta.id,
          cardName: "data-source-card",
          data: mockDataSourceGalleryData.dataSources[3],
          searchValues: [mockDataSourceGalleryData.dataSources[3].meta.name, "S3 Bucket"],
          locked: false,
          disabled: true,
          tooltip:
            "This setup cannot be selected because Inference workload does not support data sources of type S3 Bucket.",
          showDisabledInfo: false,
          sortInfo: {
            name: mockDataSourceGalleryData.dataSources[3].meta.name,
            createdAt: mockDataSourceGalleryData.dataSources[3].meta.createdAt,
            recentUsage: undefined,
          },
        },
      ];

      expect(dataSourceCards).toEqual(expectedRes);
    });
  });

  describe("getCardStateForInferenceSupport", () => {
    const props: IDataSourceGalleryProps = {
      entityType: EWorkloadFormType.Inference,
      loading: false,
      dataSources: mockDataSourceGalleryData.dataSources,
      selectedDataSources: [],
      clusterId: clusterId,
      allowedCreatedTypes: new Set<TDataSourceKinds>([
        AssetKind.Pvc,
        AssetKind.Nfs,
        AssetKind.S3,
        AssetKind.Git,
        AssetKind.HostPath,
        AssetKind.SecretVolume,
        AssetKind.ConfigMap,
      ]),
    };

    let dataSourceGalleryCmp: VueWrapper<InstanceType<typeof DataSourceGallery>>;

    beforeEach(() => {
      dataSourceGalleryCmp = mount(DataSourceGallery, {
        global: {
          plugins: [Quasar],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props,
        provide: {
          policy: {},
        },
      });
    });

    it("should return correct card state for supported inference data sources", () => {
      // AssetKind.Git
      const dataSource: DatasourceListResponseEntry | undefined = mockDataSourceGalleryData.dataSources.find(
        (dataSource: DatasourceListResponseEntry) => dataSource.meta.kind === AssetKind.Git,
      );

      expect(dataSource).toBeDefined();
      const cardState = {
        locked: false,
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      };
      const result = dataSourceGalleryCmp.vm.getCardStateForInferenceSupport(
        dataSource as DatasourceListResponseEntry,
        cardState,
      );
      expect(result).toEqual(cardState);
    });

    it("should return correct card state for unsupported inference data sources", () => {
      // AssetKind.S3
      const dataSource: DatasourceListResponseEntry | undefined = mockDataSourceGalleryData.dataSources.find(
        (dataSource: DatasourceListResponseEntry) => dataSource.meta.kind === AssetKind.S3,
      );

      expect(dataSource).toBeDefined();

      const cardState = {
        locked: false,
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      };

      const result = dataSourceGalleryCmp.vm.getCardStateForInferenceSupport(
        dataSource as DatasourceListResponseEntry,
        cardState,
      );

      expect(result).toEqual({
        locked: false,
        disabled: true,
        tooltip:
          "This setup cannot be selected because Inference workload does not support data sources of type S3 Bucket.",
        showDisabledInfo: false,
      });
    });

    it("should return correct card state for supported data sources in specific cluster version", () => {
      // AssetKind.HostPath
      const dataSource: DatasourceListResponseEntry | undefined = mockDataSourceGalleryData.dataSources.find(
        (dataSource: DatasourceListResponseEntry) => dataSource.meta.kind === AssetKind.HostPath,
      );

      expect(dataSource).toBeDefined();

      const clusterStore = useClusterStore(pinia);
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);

      const cardState = {
        locked: false,
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      };
      const result = dataSourceGalleryCmp.vm.getCardStateForInferenceSupport(
        dataSource as DatasourceListResponseEntry,
        cardState,
      );
      expect(result).toEqual(cardState);
    });

    it("should return correct card state for unsupported data sources in specific cluster version", () => {
      // AssetKind.HostPath
      const dataSource: DatasourceListResponseEntry | undefined = mockDataSourceGalleryData.dataSources.find(
        (dataSource: DatasourceListResponseEntry) => dataSource.meta.kind === AssetKind.HostPath,
      );

      expect(dataSource).toBeDefined();

      const clusterStore = useClusterStore(pinia);
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);

      const cardState = {
        locked: false,
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      };
      const result = dataSourceGalleryCmp.vm.getCardStateForInferenceSupport(
        dataSource as DatasourceListResponseEntry,
        cardState,
      );

      expect(result).toEqual({
        locked: false,
        disabled: true,
        tooltip:
          "Inference workloads don't support data sources of type Host path for the current version of the cluster.",
        showDisabledInfo: false,
      });
    });
  });
});
