/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
// Authored UI fixture. Never attach illustrative scheduler predictions to live records.
export const MOCK_SCHEDULING_IDS = [
  "preview-department-1",
  "preview-department-2",
  "preview-department-3",
  "preview-department-4",
  "preview-department-5",
];
export const MOCK_SCHEDULING_TEN_IDS = Array.from({ length: 10 }, (_, index) => `preview-department-${index + 1}`);
export const MOCK_GUARANTEED_OVERCOMMIT_QUOTAS = Object.fromEntries(
  MOCK_SCHEDULING_TEN_IDS.map((id, index) => [id, index === 0 ? 24 : 16]),
);
type MockSchedulingProfile = "zero-quota" | "guaranteed-overcommit";
const fixtures = [
  { ids: MOCK_SCHEDULING_IDS, allocation: [29, 29, 29, 29, 28] },
  { ids: MOCK_SCHEDULING_TEN_IDS, allocation: [15, 15, 15, 15, 14, 14, 14, 14, 14, 14] },
];
export function getMockSchedulingResult(
  ids: string[],
  capacity: number | undefined,
  demand: Record<string, number>,
  profile: MockSchedulingProfile = "zero-quota",
) {
  const fixture = fixtures.find(
    (fixture) => fixture.ids.length === ids.length && fixture.ids.every((id) => ids.includes(id)),
  );
  if (
    !fixture ||
    (profile === "guaranteed-overcommit" && fixture.ids.length !== 10) ||
    capacity !== 144 ||
    Object.entries(demand).some(([id, value]) => !fixture.ids.includes(id) || value !== 144)
  )
    return undefined;
  return Object.fromEntries(
    fixture.ids.map((id, index) => [
      id,
      {
        allocated: fixture.allocation[index],
        unmet: 144 - fixture.allocation[index],
        guaranteed: profile === "guaranteed-overcommit" ? fixture.allocation[index] : 0,
        overQuota: profile === "guaranteed-overcommit" ? 0 : fixture.allocation[index],
      },
    ]),
  );
}
