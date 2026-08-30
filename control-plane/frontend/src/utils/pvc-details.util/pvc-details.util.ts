import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import type { StorageClassSpec } from "@/swagger-models/k8s-objects-tracker-client";
import type { Tenant } from "@/swagger-models/tenants-manager-client";
import type { IDetailsPanelItem, IDetailsPanelSectionDef, TDetailsPanelItemGroup } from "@/models/details-panel.model";
import type {
  IPvcGeneralSectionData,
  IPvcRequestSectionData,
  IPvcStorageClassSectionData,
} from "@/models/pvc-details.model";
import { matchesItem } from "@/utils/details-panel.utils";
import { tableDisplayFormatters } from "@/utils/table-display-formatters.util";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

export const PVC_DETAILS_SECTIONS = {
  GENERAL: { id: "general", title: "General information" },
  REQUEST: { id: "request", title: "Request" },
  STORAGE_CLASS: { id: "storage-class", title: "Storage class" },
} as const;

const EMPTY_VALUE = "-";

export interface IBuildPvcSectionsInput {
  pvc: DatasourceListResponseEntry;
  storageClassSpec: StorageClassSpec | null;
  tenant: Tenant;
  orgUnitList: OrgUnit[];
}

export function buildPvcSections(input: IBuildPvcSectionsInput): IDetailsPanelSectionDef[] {
  return [buildGeneralSection(input), buildRequestSection(input), buildStorageClassSection(input)].filter(
    (s): s is IDetailsPanelSectionDef => s !== undefined,
  );
}

function matchAnyItem(items: IDetailsPanelItem[], term: string): boolean {
  return items.some((it) => matchesItem(it, term));
}

function buildGeneralSection(input: IBuildPvcSectionsInput): IDetailsPanelSectionDef | undefined {
  const { pvc, tenant, orgUnitList } = input;
  const groups: TDetailsPanelItemGroup[] = [
    [
      { label: "Description", value: pvc.meta.description || EMPTY_VALUE },
      { label: "Scope", value: tableDisplayFormatters.getScopeName(pvc, tenant, orgUnitList) || EMPTY_VALUE },
    ],
    [
      { label: "Creation time", value: formatDate(pvc.meta.createdAt) },
      { label: "Created by", value: pvc.meta.createdBy || EMPTY_VALUE },
    ],
  ];
  const data: IPvcGeneralSectionData = { groups };
  return {
    ...PVC_DETAILS_SECTIONS.GENERAL,
    customMatch: (term: string) => matchAnyItem(groups.flat(), term),
    data,
  };
}

function buildRequestSection(input: IBuildPvcSectionsInput): IDetailsPanelSectionDef | undefined {
  const claim = input.pvc.spec.pvc?.claimInfo;
  const readOnly = input.pvc.spec.pvc?.readOnly;
  const restrictions = readOnly === null || readOnly === undefined ? EMPTY_VALUE : readOnly ? "Read-only" : "Read-write";
  const items: IDetailsPanelItem[] = [
    { label: "Access modes", value: formatAccessModes(claim?.accessModes) },
    { label: "Storage request", value: claim?.size || EMPTY_VALUE },
    { label: "Volume mode", value: claim?.volumeMode || EMPTY_VALUE },
    { label: "Restrictions", value: restrictions },
  ];
  if (items.every((it) => it.value === EMPTY_VALUE)) return undefined;
  const data: IPvcRequestSectionData = { items };
  return {
    ...PVC_DETAILS_SECTIONS.REQUEST,
    customMatch: (term: string) => matchAnyItem(items, term),
    data,
  };
}

function buildStorageClassSection(input: IBuildPvcSectionsInput): IDetailsPanelSectionDef | undefined {
  const { pvc, storageClassSpec } = input;
  const items: IDetailsPanelItem[] = [
    { label: "Name", value: pvc.spec.pvc?.claimInfo?.storageClass || EMPTY_VALUE },
    { label: "Provisioner type", value: storageClassSpec?.provisioner || EMPTY_VALUE },
    { label: "Allow volume expansion", value: formatBoolean(storageClassSpec?.allow_volume_expansion) },
    { label: "Is default storage class", value: formatBoolean(storageClassSpec?.is_default) },
  ];
  if (items.every((it) => it.value === EMPTY_VALUE)) return undefined;
  const data: IPvcStorageClassSectionData = { items };
  return {
    ...PVC_DETAILS_SECTIONS.STORAGE_CLASS,
    customMatch: (term: string) => matchAnyItem(items, term),
    data,
  };
}

function formatAccessModes(modes: unknown): string {
  if (!modes || typeof modes !== "object") return EMPTY_VALUE;
  const enabled = Object.entries(modes as Record<string, boolean>)
    .filter(([, value]) => value)
    .map(([key]) => key);
  return enabled.length ? enabled.join(", ") : EMPTY_VALUE;
}

function formatDate(iso?: string | null): string {
  if (!iso) return EMPTY_VALUE;
  return dateUtil.dateFormat(new Date(iso), DATE_FORMATS.STANDARD_DATE);
}

function formatBoolean(value?: boolean | null): string {
  if (value === null || value === undefined) return EMPTY_VALUE;
  return value ? "Yes" : "No";
}
