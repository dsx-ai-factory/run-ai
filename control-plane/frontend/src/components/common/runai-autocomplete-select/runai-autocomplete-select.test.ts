/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiAutocompleteSelect } from "@/components/common";
import { createTestingPinia } from "@pinia/testing";
import { vi } from "vitest";
import { setActivePinia } from "pinia";
import { expect, test } from "vitest";

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("runai-autocomplete-select.vue", () => {
  let runaiAutoCompleteSelectCpms: VueWrapper<InstanceType<typeof RunaiAutocompleteSelect>>;

  beforeEach(() => {
    runaiAutoCompleteSelectCpms = mount(RunaiAutocompleteSelect, {
      global: {
        plugins: [Quasar],
      },
      props: {
        label: "label",
        placeholder: "placeholder",
        noResultLabel: "noResultLabel",
        searchHint: "searchHint",
        asyncOptionsPromise: () =>
          Promise.resolve([
            { value: "opt1", label: "opt1" },
            { value: "opt2", label: "opt2" },
            { value: "opt3", label: "opt3" },
            { value: "opt4", label: "opt4" },
            { value: "other", label: "other" },
          ]),
      },
    });
  });

  test("component is loading properly", async () => {
    expect(runaiAutoCompleteSelectCpms.html()).toContain("placeholder");
  });

  test("displays the input label", () => {
    const label = "Select an option:";
    runaiAutoCompleteSelectCpms = mount(RunaiAutocompleteSelect, {
      global: {
        plugins: [Quasar],
      },
      props: {
        label,
        noResultLabel: "No results found",
        searchHint: "Start typing to search",
        asyncOptionsPromise: async () => [],
      },
    });

    const inputLabel = runaiAutoCompleteSelectCpms.find(".q-placeholder");
    expect(inputLabel.html()).toContain(label);
  });
});
