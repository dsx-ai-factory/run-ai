<template>
  <chart-widget-wrapper
    class="org-unit-ranking-widget"
    aid="org-unit-ranking-widget"
    :options="wrapperOptions"
    :loading="loading"
    :error="error"
    slot-wrapper-class=""
    hide-actions
    is-grid
    size="auto"
    skeleton-height="290px"
  >
    <template #header="{ titleId }">
      <h3 :id="titleId" class="chart-widget-heading--visually-hidden">{{ wrapperOptions.title }}</h3>
      <widget-header-select-wrapper>
        <runai-select aid="ranking-entity-select" :options="entityOptions" v-model="selectedEntityOption" />
        <runai-select aid="ranking-metric-select" :options="metricOptions" v-model="selectedMetricOption" />
        <template #right>
          <widget-link-btn
            class="q-mr-md"
            aid="ranking-redirect-btn"
            v-if="wrapperOptions.linkText"
            :link-text="wrapperOptions.linkText"
            :entity-type="wrapperOptions.entityType"
            @click="redirectToEntityIndex"
          />
        </template>
      </widget-header-select-wrapper>
    </template>
    <runai-table
      :rows="rankingRecords"
      :columns="columns"
      :filter-by="filterBy"
      :loading="loading"
      disable-selection
      no-shadow
      :hide-pagination="empty"
      :rows-per-page-options="[5]"
      :bordered="false"
      aid="ranking-table"
      override-table-header
      @project-clicked="handleProjectClicked"
      @department-clicked="handleDepartmentClicked"
      @ranking-timeframe-changed="onRankingTimeframeChanged"
      @update-filters="updateFilters"
    >
      <template #no-data>
        <runai-table-no-data
          :filter-by="filterBy"
          :entity-name="entityName"
          :custom-message="emptyStateMessage"
          :show-create-btn="false"
          :show-filter-icon-and-button="false"
      /></template>
    </runai-table>
  </chart-widget-wrapper>
</template>
<script lang="ts">
import { computed, defineComponent, type PropType, ref, watch } from "vue";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { useRouter } from "vue-router";
import { EFilterOperator } from "@/models/filter.model";
import { EWidgetEntity } from "@/models/chart.model";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { useRefresh } from "@/composables/use-refresh.composable";
import { EIntervalLabels } from "@/models/interval.model";
import { RunaiSelect } from "@/components/common/runai-select";
import {
  EOrgUnitEntityRankingValue,
  EOrgUnitOverTimeColumnName,
  EOrgUnitOverTimeValue,
  EOrgUnitRankingMetricValue,
  type IOrgUnitEntitySelectOption,
  type IOrgUnitRankingMetricSelectOption,
  type IOrgUnitRankingTableFilterBy,
  NUMBER_OF_RANKING_ENTITIES_PER_PAGE,
  ORG_UNIT_RANKING_ENTITY_SELECT_OPTIONS,
  ORG_UNIT_RANKING_METRIC_SELECT_OPTIONS,
} from "@/models/org-unit.model";
import { WidgetLinkBtn } from "@/components/dashboard-v2/widgets/common/buttons/widget-link-btn";
import {
  departmentsRankingColumns,
  EOrgUnitProjectsRankingColumnName,
  EOrgUnitRankingColumnName,
  type IOrgUnitRankingTable,
  projectsRankingColumns,
} from "@/table-models/org-unit.table-model";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { ETableFilters, HIGHLIGHT_COLUMN_CONFIG } from "@/models/table.model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { filterService } from "@/services/filter.service/filter.service";
import type { IProjectTableFilterBy } from "@/models/project.model";
import type { IDepartmentTableFilterBy } from "@/models/department.model";
import { EProjectColumnLabel, projectIndexColumns } from "@/table-models/project.table-model";
import { departmentIndexColumns } from "@/table-models/department.table-model";
import { DepartmentFilterSortFields, ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { WidgetHeaderSelectWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/widget-header-select-wrapper";
const TOP_METRIC_FILTER_VALUE = 1;
const BOTTOM_METRIC_FILTER_VALUE = 0;
export default defineComponent({
  name: "org-unit-ranking-widget",
  components: {
    WidgetHeaderSelectWrapper,
    WidgetLinkBtn,
    RunaiSelect,
    RunaiTableNoData,
    ChartWidgetWrapper,
    RunaiTable,
  },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
    hasReadProjectsPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasReadDepartmentsPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  setup(props) {
    const router = useRouter();

    const loading = ref(true);
    const error = ref(false);
    const rankingRecords = ref<IOrgUnitRankingTable[]>([]);
    const metricOptions = ref<IOrgUnitRankingMetricSelectOption[]>(ORG_UNIT_RANKING_METRIC_SELECT_OPTIONS);
    const selectedEntityOption = ref<IOrgUnitEntitySelectOption>(ORG_UNIT_RANKING_ENTITY_SELECT_OPTIONS[0]);
    const selectedMetricOption = ref<IOrgUnitRankingMetricSelectOption>(ORG_UNIT_RANKING_METRIC_SELECT_OPTIONS[0]);

    const rankingTimeframe = ref<EOrgUnitOverTimeValue>(EOrgUnitOverTimeValue.TwentyFourHours);

    const filterBy = ref<IOrgUnitRankingTableFilterBy>({
      descending: true,
      sortBy: EOrgUnitRankingColumnName.AllocationValue,
      rankingTimeframe: rankingTimeframe.value,
      rowsPerPage: NUMBER_OF_RANKING_ENTITIES_PER_PAGE,
    });

    const departmentCount = ref(0);

    const isMultipleDepartments = computed(() => departmentCount.value > 1);
    const entityOptions = computed(() => {
      return ORG_UNIT_RANKING_ENTITY_SELECT_OPTIONS.filter((option) => {
        if (option.value === EOrgUnitEntityRankingValue.Project) {
          return props.hasReadProjectsPermission;
        }
        if (option.value === EOrgUnitEntityRankingValue.Department) {
          return props.hasReadDepartmentsPermission && isMultipleDepartments.value;
        }
        return false;
      });
    });

    // Ensure selectedEntityOption is valid when entityOptions change
    watch(
      entityOptions,
      (newOptions) => {
        if (newOptions.length > 0 && !newOptions.find((opt) => opt.value === selectedEntityOption.value.value)) {
          selectedEntityOption.value = newOptions[0];
        }
      },
      { immediate: true },
    );
    const getDepartmentCount = async () => {
      try {
        const filterBy = [filterUtil.getEqualsFilterString(DepartmentFilterSortFields.ClusterId, props.clusterId)];
        departmentCount.value = await orgUnitService.countDepartments(filterBy);
      } catch (err) {
        console.error(`Failed to get department count: ${err}`);
      }
    };

    const getEntityRanking = async () => {
      try {
        const { metricType, order } = orgUnitUtil.mapRankingMetricToMetricTypeAndOrder(
          selectedMetricOption.value.value,
          rankingTimeframe.value,
        );
        error.value = false;
        rankingRecords.value = await getEntityFunc.value(metricType, order, props.clusterId, props.nodePoolName);
      } catch (err) {
        error.value = true;
      } finally {
        loading.value = false;
      }
    };

    const redirectToEntityIndex = () => {
      selectAvgResourceColumn();
      const advancedFilters = getAdvancedFiltersQuery();
      const routeName = isProject.value ? PROJECT_ROUTE_NAMES.PROJECT_INDEX : DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX;
      router.push({ name: routeName, query: { advancedFilters } });
    };

    const selectAvgResourceColumn = () => {
      const tableFilterName = getTableFilterName();
      const filters = loadEntityFilters(tableFilterName);
      updateDisplayedColumns(filters);
      updateEntityFilters(filters);
      filterService.saveFilters(tableFilterName, filters);
    };

    const getTableFilterName = () => {
      return isProject.value ? ETableFilters.PROJECT : ETableFilters.DEPARTMENT;
    };

    const loadEntityFilters = (tableFilterName: ETableFilters) => {
      return filterService.loadFilters(window.location, tableFilterName, {}, true) as
        | IProjectTableFilterBy
        | IDepartmentTableFilterBy;
    };

    const updateDisplayedColumns = (filters: IProjectTableFilterBy | IDepartmentTableFilterBy) => {
      const defaultColumns = isProject.value ? projectIndexColumns : departmentIndexColumns;
      if (filters.displayedColumns?.length) {
        if (!filters.displayedColumns.includes(EOrgUnitOverTimeColumnName.AverageGpuUtilization)) {
          filters.displayedColumns.push(EOrgUnitOverTimeColumnName.AverageGpuUtilization);
        }
        if (!filters.displayedColumns.includes(EOrgUnitOverTimeColumnName.AverageGpuAllocation)) {
          filters.displayedColumns.push(EOrgUnitOverTimeColumnName.AverageGpuAllocation);
        }
      } else {
        filters.displayedColumns = [
          ...defaultColumns.filter((col) => !!col.display).map((col) => col.name),
          EOrgUnitOverTimeColumnName.AverageGpuUtilization,
          EOrgUnitOverTimeColumnName.AverageGpuAllocation,
        ];
      }
    };

    const getAdvancedFiltersQuery = () => {
      const minValue = isTopAverageMetric.value ? TOP_METRIC_FILTER_VALUE : BOTTOM_METRIC_FILTER_VALUE;
      const name = isAllocationMetric.value
        ? EOrgUnitOverTimeColumnName.AverageGpuAllocation
        : EOrgUnitOverTimeColumnName.AverageGpuUtilization;
      const label = isAllocationMetric.value
        ? EProjectColumnLabel.AverageGpuAllocation
        : EProjectColumnLabel.AverageGpuUtilization;

      return `[${filterUtil.getNumericFilterQuery(name, label, EFilterOperator.GreaterThanOrEqual, minValue)}]`;
    };
    const updateEntityFilters = (filters: IProjectTableFilterBy | IDepartmentTableFilterBy) => {
      filters.avgResourceTimeframe = rankingTimeframe.value;
      filters.descending = isTopAverageMetric.value;
      filters.sortBy = isAllocationMetric.value
        ? EOrgUnitOverTimeColumnName.AverageGpuAllocation
        : EOrgUnitOverTimeColumnName.AverageGpuUtilization;
    };

    const highlightColumn = () => {
      if (isAllocationMetric.value) {
        filterBy.value.sortBy = EOrgUnitRankingColumnName.AllocationValue;
      } else if (isMemoryUtilizationMetric.value) {
        filterBy.value.sortBy = EOrgUnitRankingColumnName.MemoryUtilizationValue;
      } else if (isGpuMemoryUtilizationMetric.value) {
        filterBy.value.sortBy = EOrgUnitRankingColumnName.MemoryUtilizationValue;
      } else {
        filterBy.value.sortBy = EOrgUnitRankingColumnName.UtilizationValue;
      }

      filterBy.value.descending = isTopAverageMetric.value;
    };

    const isMemoryUtilizationMetric = computed(() =>
      [
        EOrgUnitRankingMetricValue.TopAverageGpuMemoryUtilization,
        EOrgUnitRankingMetricValue.BottomAverageGpuMemoryUtilization,
      ].includes(selectedMetricOption.value.value),
    );

    const isGpuMemoryUtilizationMetric = computed(() =>
      [
        EOrgUnitRankingMetricValue.TopAverageGpuMemoryUtilization,
        EOrgUnitRankingMetricValue.BottomAverageGpuMemoryUtilization,
      ].includes(selectedMetricOption.value.value),
    );

    const columns = computed(() => {
      const entityColumns = isProject.value ? projectsRankingColumns : departmentsRankingColumns;

      let highlightColumnName: EOrgUnitRankingColumnName;
      if (isAllocationMetric.value) {
        highlightColumnName = EOrgUnitRankingColumnName.AllocationValue;
      } else if (isMemoryUtilizationMetric.value) {
        highlightColumnName = EOrgUnitRankingColumnName.MemoryUtilizationValue;
      } else {
        highlightColumnName = EOrgUnitRankingColumnName.UtilizationValue;
      }

      return entityColumns
        .filter((col) => isMultipleDepartments.value || col.name !== EOrgUnitProjectsRankingColumnName.ParentName)
        .map((col) => (col.name === highlightColumnName ? { ...col, ...HIGHLIGHT_COLUMN_CONFIG, sortable: true } : col));
    });

    const wrapperOptions = computed(() => {
      const linkText = isProject.value ? "All projects" : "All departments";
      const entityType = isProject.value ? EWidgetEntity.Project : EWidgetEntity.Department;
      return {
        title: "",
        linkText,
        entityType,
      };
    });

    const isProject = computed(() => selectedEntityOption.value.value === EOrgUnitEntityRankingValue.Project);
    const entityName = computed(() => (isProject.value ? "project" : "department"));
    const empty = computed(() => !loading.value && rankingRecords.value.length === 0);
    const emptyStateMessage = computed(() => (isProject.value ? "No projects to display" : "No departments to display"));
    const getEntityFunc = computed(() =>
      isProject.value ? orgUnitService.getProjectsRanking : orgUnitService.getDepartmentsRanking,
    );

    const isAllocationMetric = computed(() =>
      [
        EOrgUnitRankingMetricValue.TopAverageGpuAllocation,
        EOrgUnitRankingMetricValue.BottomAverageGpuAllocation,
      ].includes(selectedMetricOption.value.value),
    );

    const isTopAverageMetric = computed(() =>
      [
        EOrgUnitRankingMetricValue.TopAverageGpuAllocation,
        EOrgUnitRankingMetricValue.TopAverageGpuUtilization,
        EOrgUnitRankingMetricValue.TopAverageGpuMemoryUtilization,
      ].includes(selectedMetricOption.value.value),
    );

    watch(
      [selectedEntityOption, rankingTimeframe, selectedMetricOption, () => props.nodePoolName, () => props.clusterId],
      async () => {
        loading.value = true;
        error.value = false;
        highlightColumn();
        await getEntityRanking();
      },
    );

    const handleProjectClicked = (record: IOrgUnitRankingTable) => {
      router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
        query: {
          [ProjectFilterSortFields.Name]: record.name,
          [ProjectFilterSortFields.ClusterId]: props.clusterId,
        },
      });
    };

    const handleDepartmentClicked = (record: IOrgUnitRankingTable) => {
      router.push({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
        query: {
          [DepartmentFilterSortFields.Name]: record.name,
          [DepartmentFilterSortFields.ClusterId]: props.clusterId,
        },
      });
    };

    const onRankingTimeframeChanged = (timeframe: string) => {
      rankingTimeframe.value = timeframe as EOrgUnitOverTimeValue;
      filterBy.value.rankingTimeframe = timeframe;
    };

    const updateFilters = (filters: IOrgUnitRankingTableFilterBy) => {
      if (filters.descending !== filterBy.value.descending) return; // prevent sorting change from here
      filterBy.value = filters;
    };

    useRefresh(EIntervalLabels.OrgUnitRankingWidget, async () => {
      await getEntityRanking();
    });
    getDepartmentCount();
    return {
      loading,
      error,
      empty,
      rankingRecords,
      filterBy,
      wrapperOptions,
      redirectToEntityIndex,
      entityOptions,
      selectedEntityOption,
      metricOptions,
      selectedMetricOption,
      columns,
      emptyStateMessage,
      entityName,
      handleProjectClicked,
      handleDepartmentClicked,
      onRankingTimeframeChanged,
      updateFilters,
    };
  },
});
</script>

<style scoped lang="scss">
.org-unit-ranking-widget {
  .q-table th.sort-desc .q-table__sort-icon {
    margin-top: 5px;
  }
  :deep(.runai-table) {
    display: flex;
    flex-direction: column;

    .q-table__container {
      flex: 1;
      display: flex;
      flex-direction: column;
    }

    .q-table__middle {
      flex: 1;
    }
  }
}
</style>

<style lang="scss">
.org-unit-ranking-widget {
  .q-table th.sort-desc .q-table__sort-icon {
    margin-top: 5px;
  }
}
</style>
