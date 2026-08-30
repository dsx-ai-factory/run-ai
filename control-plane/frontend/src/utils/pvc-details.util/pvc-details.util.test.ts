import { describe, it, expect, vi } from "vitest";
import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import type { StorageClassSpec } from "@/swagger-models/k8s-objects-tracker-client";
import type { Tenant } from "@/swagger-models/tenants-manager-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import type {
  IPvcGeneralSectionData,
  IPvcRequestSectionData,
  IPvcStorageClassSectionData,
} from "@/models/pvc-details.model";

import { buildPvcSections, PVC_DETAILS_SECTIONS, type IBuildPvcSectionsInput } from "./pvc-details.util";

vi.mock("@/utils/table-display-formatters.util", () => ({
  tableDisplayFormatters: {
    getScopeName: () => "Project Alpha",
  },
}));

vi.mock("@/utils/date.util", () => ({
  dateUtil: {
    dateFormat: (date: Date) => `formatted:${date.toISOString()}`,
  },
}));

const EMPTY_VALUE = "-";

type PvcOverrides = {
  meta?: Partial<DatasourceListResponseEntry["meta"]>;
  spec?: Partial<DatasourceListResponseEntry["spec"]>;
};

function makePvc(overrides: PvcOverrides = {}): DatasourceListResponseEntry {
  return {
    meta: {
      description: "A test PVC",
      createdAt: "2026-05-18T10:00:00Z",
      createdBy: "alice@example.com",
      ...(overrides.meta ?? {}),
    },
    spec: {
      pvc: {
        readOnly: false,
        claimInfo: {
          size: "5Gi",
          storageClass: "standard",
          accessModes: { readWriteOnce: true, readOnlyMany: false, readWriteMany: false },
          volumeMode: "Filesystem",
        },
        ...(overrides.spec?.pvc ?? {}),
      },
      ...(overrides.spec ?? {}),
    },
    ...overrides,
  } as DatasourceListResponseEntry;
}

const tenant = {} as Tenant;
const orgUnitList: OrgUnit[] = [];

function makeInput(overrides: Partial<IBuildPvcSectionsInput> = {}): IBuildPvcSectionsInput {
  return {
    pvc: overrides.pvc ?? makePvc(),
    storageClassSpec:
      overrides.storageClassSpec !== undefined
        ? overrides.storageClassSpec
        : ({
            provisioner: "kubernetes.io/aws-ebs",
            allow_volume_expansion: true,
            is_default: false,
          } as StorageClassSpec),
    tenant: overrides.tenant ?? tenant,
    orgUnitList: overrides.orgUnitList ?? orgUnitList,
  };
}

function findItem(items: { label: string; value: string | string[] }[], label: string): string | string[] | undefined {
  return items.find((it) => it.label === label)?.value;
}

describe("buildPvcSections", () => {
  it("returns all three sections when data is fully populated", () => {
    const sections = buildPvcSections(makeInput());
    expect(sections.map((s) => s.id)).toEqual([
      PVC_DETAILS_SECTIONS.GENERAL.id,
      PVC_DETAILS_SECTIONS.REQUEST.id,
      PVC_DETAILS_SECTIONS.STORAGE_CLASS.id,
    ]);
  });

  describe("general section", () => {
    it("falls back to EMPTY_VALUE when description / createdBy are missing", () => {
      const sections = buildPvcSections(makeInput({ pvc: makePvc({ meta: { createdAt: "2026-05-18T10:00:00Z" } }) }));
      const general = sections.find((s) => s.id === PVC_DETAILS_SECTIONS.GENERAL.id);
      const items = (general?.data as IPvcGeneralSectionData).groups.flat();
      expect(findItem(items, "Description")).toBe(EMPTY_VALUE);
      expect(findItem(items, "Created by")).toBe(EMPTY_VALUE);
    });

    it("formats creation time via dateUtil and resolves scope via tableDisplayFormatters", () => {
      const sections = buildPvcSections(makeInput());
      const general = sections.find((s) => s.id === PVC_DETAILS_SECTIONS.GENERAL.id);
      const items = (general?.data as IPvcGeneralSectionData).groups.flat();
      expect(findItem(items, "Scope")).toBe("Project Alpha");
      expect(findItem(items, "Creation time")).toMatch(/^formatted:2026-05-18T10:00:00/);
    });
  });

  describe("request section", () => {
    it("renders Read-only when readOnly is true", () => {
      const pvc = makePvc({ spec: { pvc: { readOnly: true } } });
      const sections = buildPvcSections(makeInput({ pvc }));
      const items = (sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id)?.data as IPvcRequestSectionData)
        .items;
      expect(findItem(items, "Restrictions")).toBe("Read-only");
    });

    it("renders Read-write when readOnly is false", () => {
      const sections = buildPvcSections(makeInput());
      const items = (sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id)?.data as IPvcRequestSectionData)
        .items;
      expect(findItem(items, "Restrictions")).toBe("Read-write");
    });

    it("renders EMPTY_VALUE for Restrictions when readOnly is undefined (regression for CodeRabbit fix)", () => {
      const pvc = makePvc({ spec: { pvc: {} } });
      const sections = buildPvcSections(makeInput({ pvc }));
      const request = sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id);
      expect(request).toBeUndefined();
    });

    it("hides the section entirely when claimInfo + readOnly are absent", () => {
      const pvc = makePvc({ spec: {} });
      const sections = buildPvcSections(makeInput({ pvc }));
      expect(sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id)).toBeUndefined();
    });

    it("joins multiple enabled access modes with a comma", () => {
      const pvc = makePvc({
        spec: {
          pvc: {
            claimInfo: { accessModes: { readWriteOnce: true, readOnlyMany: true, readWriteMany: false } },
          },
        },
      });
      const sections = buildPvcSections(makeInput({ pvc }));
      const items = (sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id)?.data as IPvcRequestSectionData)
        .items;
      expect(findItem(items, "Access modes")).toBe("readWriteOnce, readOnlyMany");
    });

    it("renders EMPTY_VALUE for Access modes when no mode is enabled (and keeps the section if other fields are set)", () => {
      const pvc = makePvc({
        spec: {
          pvc: {
            readOnly: false,
            claimInfo: {
              size: "10Gi",
              accessModes: { readWriteOnce: false, readOnlyMany: false, readWriteMany: false },
            },
          },
        },
      });
      const sections = buildPvcSections(makeInput({ pvc }));
      const items = (sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id)?.data as IPvcRequestSectionData)
        .items;
      expect(findItem(items, "Access modes")).toBe(EMPTY_VALUE);
    });
  });

  describe("storage class section", () => {
    it("renders Yes / No via formatBoolean and EMPTY_VALUE for nullish booleans", () => {
      const sections = buildPvcSections(makeInput());
      const items = (
        sections.find((s) => s.id === PVC_DETAILS_SECTIONS.STORAGE_CLASS.id)?.data as IPvcStorageClassSectionData
      ).items;
      expect(findItem(items, "Allow volume expansion")).toBe("Yes");
      expect(findItem(items, "Is default storage class")).toBe("No");
    });

    it("hides the section when no name and no storageClassSpec are provided", () => {
      const pvc = makePvc({ spec: { pvc: { claimInfo: {} } } });
      const sections = buildPvcSections(makeInput({ pvc, storageClassSpec: null }));
      expect(sections.find((s) => s.id === PVC_DETAILS_SECTIONS.STORAGE_CLASS.id)).toBeUndefined();
    });

    it("renders the section when only the storage class name is known", () => {
      const sections = buildPvcSections(makeInput({ storageClassSpec: null }));
      const storageClass = sections.find((s) => s.id === PVC_DETAILS_SECTIONS.STORAGE_CLASS.id);
      const items = (storageClass?.data as IPvcStorageClassSectionData).items;
      expect(findItem(items, "Name")).toBe("standard");
      expect(findItem(items, "Provisioner type")).toBe(EMPTY_VALUE);
    });
  });

  describe("customMatch", () => {
    it("matches against label and value substrings (case-insensitive) in the general section", () => {
      const sections = buildPvcSections(makeInput());
      const general = sections.find((s) => s.id === PVC_DETAILS_SECTIONS.GENERAL.id);
      expect(general?.customMatch("alice")).toBe(true);
      expect(general?.customMatch("project alpha")).toBe(true);
      expect(general?.customMatch("does-not-exist")).toBe(false);
    });

    it("matches against the request section's items", () => {
      const sections = buildPvcSections(makeInput());
      const request = sections.find((s) => s.id === PVC_DETAILS_SECTIONS.REQUEST.id);
      expect(request?.customMatch("readwriteonce")).toBe(true);
      expect(request?.customMatch("5gi")).toBe(true);
      expect(request?.customMatch("nope")).toBe(false);
    });
  });
});
