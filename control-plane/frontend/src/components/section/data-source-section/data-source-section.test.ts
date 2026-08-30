import { mount } from "@vue/test-utils";
import DataSourceSection from "./data-source-section.vue";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { AssetIdAndKind, PolicyInfo, DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import { mockDataSourceSectionData } from "./data-source-section.mock";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { useClusterStore } from "@/stores/cluster.store";

// models
import { AssetKind } from "@/swagger-models/assets-service-client";
import { EWorkloadFormType } from "@/models/workload.model";
import type { IWorkloadDataSourceSectionOptions } from "./data-source-section.model";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

interface IDataSourceSectionProps {
  entityType: EWorkloadFormType;
  loading: boolean;
  dataSources: Array<DatasourceListResponseEntry>;
  selectedDataSources: AssetIdAndKind[];
  defaultOpened?: boolean;
  sectionOptions?: IWorkloadDataSourceSectionOptions;
  sectionDisabled?: boolean;
  clusterId: string;
  policyRules?: PolicyInfo | null;
}

describe("data-source-section.vue", () => {
  let dataSourceSectionCmp: VueWrapper<InstanceType<typeof DataSourceSection>>;
  const clusterId = "123";
  const clusterStore = useClusterStore(pinia);

  function getComponent(
    props: IDataSourceSectionProps,
    policy = {},
  ): VueWrapper<InstanceType<typeof DataSourceSection>> {
    return mount(DataSourceSection, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
      },
      props,
      provide: {
        policy,
      },
    });
  }

  describe("general tests", () => {
    const props: IDataSourceSectionProps = {
      entityType: EWorkloadFormType.Workspace,
      loading: false,
      dataSources: mockDataSourceSectionData.dataSources,
      defaultOpened: false,
      selectedDataSources: [],
      clusterId: clusterId,
    };

    beforeEach(() => {
      dataSourceSectionCmp = mount(DataSourceSection, {
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

    it("component is loading properly", () => {
      expect(dataSourceSectionCmp.html()).toContain("Data sources");
      expect(dataSourceSectionCmp.html()).toContain("new data source");
      expect(dataSourceSectionCmp.html()).toContain("Select the data sources your workload needs to access");
    });

    it("section summary should be 'None' when selectedDataSources is empty", () => {
      expect(dataSourceSectionCmp.vm.dataSourceSummary).toBe("None");
    });

    it("section summary should display the correct text when a single data source is selected", async () => {
      const selected: AssetIdAndKind[] = [
        {
          id: props.dataSources[1].meta.id,
          kind: props.dataSources[1].meta.kind,
        },
      ];
      await dataSourceSectionCmp.setProps({ selectedDataSources: selected });
      expect(dataSourceSectionCmp.vm.dataSourceSummary).toBe(props.dataSources[1].meta.name);
    });

    it("section summary should display the correct text when multiple data sources are selected", async () => {
      const selected: AssetIdAndKind[] = [
        {
          id: props.dataSources[1].meta.id,
          kind: props.dataSources[1].meta.kind,
        },
        {
          id: props.dataSources[0].meta.id,
          kind: props.dataSources[0].meta.kind,
        },
      ];
      await dataSourceSectionCmp.setProps({ selectedDataSources: selected });
      expect(dataSourceSectionCmp.vm.dataSourceSummary).toBe(mockDataSourceSectionData.dataSources[1].meta.name + " +1");
    });
  });

  describe("sectionOptions", () => {
    const props: IDataSourceSectionProps = {
      entityType: EWorkloadFormType.Workspace,
      loading: false,
      dataSources: mockDataSourceSectionData.dataSources,
      defaultOpened: false,
      selectedDataSources: [],
      clusterId: clusterId,
    };

    it("should display the default title in not provided in sectionOptions", () => {
      const dataSourceSectionCmp = getComponent(props);
      expect(dataSourceSectionCmp.html()).toContain("Select the data sources your workload needs to access");
    });

    it("should display the custom title when is is provided in sectionOptions", () => {
      const dataSourceSectionCmp = getComponent({
        ...props,
        sectionOptions: {
          cardsTitle: "Some title",
        },
      });
      expect(dataSourceSectionCmp.html()).toContain("Some title");
    });
  });

  describe("allowedCreatedTypes", () => {
    const props: IDataSourceSectionProps = {
      entityType: EWorkloadFormType.Workspace,
      loading: false,
      dataSources: mockDataSourceSectionData.dataSources,
      defaultOpened: false,
      selectedDataSources: [],
      clusterId: clusterId,
    };

    it("should allow all default data sources when no policy rules for workspace", () => {
      const dataSourceSectionCmp = getComponent(props);
      const result = dataSourceSectionCmp.vm.allowedCreatedTypes;
      expect(result).toEqual(
        new Set([
          AssetKind.Pvc,
          AssetKind.Nfs,
          AssetKind.S3,
          AssetKind.Git,
          AssetKind.HostPath,
          AssetKind.SecretVolume,
          AssetKind.ConfigMap,
        ]),
      );
    });

    it("should remove restricted data sources based on policy rules", () => {
      const dataSourceSectionCmp = getComponent({
        ...props,
        policyRules: {
          rules: {
            assets: {
              datasources: {
                pvc: { instances: { canAdd: false } },
                nfs: { instances: { canAdd: true } },
                s3: { instances: { canAdd: false } },
              },
            },
          },
        },
      });

      const result = dataSourceSectionCmp.vm.allowedCreatedTypes;
      expect(result).toEqual(
        new Set([AssetKind.Nfs, AssetKind.Git, AssetKind.HostPath, AssetKind.SecretVolume, AssetKind.ConfigMap]),
      );
    });

    it("should remove S3 and SecretVolume for Inference workloads", () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);

      const dataSourceSectionCmp = getComponent({
        ...props,
        entityType: EWorkloadFormType.Inference,
      });

      const result = dataSourceSectionCmp.vm.allowedCreatedTypes;

      expect(result).toEqual(
        new Set([AssetKind.Pvc, AssetKind.Nfs, AssetKind.Git, AssetKind.HostPath, AssetKind.ConfigMap]),
      );
    });

    it("should remove Nfs and HostPath when cluster does not support them for Inference workloads", () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);

      const dataSourceSectionCmp = getComponent({
        ...props,
        entityType: EWorkloadFormType.Inference,
      });

      const result = dataSourceSectionCmp.vm.allowedCreatedTypes;

      expect(result).toEqual(new Set([AssetKind.Pvc, AssetKind.Git, AssetKind.ConfigMap]));
    });
  });
});
