/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it } from "vitest";
import { urlService } from "@/services/url.service/url.service";
import type { IFilterBy } from "@/models/filter.model";

describe("url.service", () => {
  function setLocation(url: string): void {
    Object.defineProperty(window, "location", {
      value: {
        href: url,
      },
    });
  }

  describe("getFiltersObj", () => {
    const baseUrl = "http://example.com";

    beforeEach(() => {
      global.window = Object.create(window);
    });

    it("should return an empty object", () => {
      const url = `${baseUrl}`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res).toEqual({});
    });

    it("should return object contains sortBy prop", () => {
      const url = `${baseUrl}?sortBy=name`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res.sortBy).toEqual("name");
    });

    it("should return object contains query prop", () => {
      const url = `${baseUrl}?query=name`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res.query).toEqual("name");
    });

    it("should return object contains descending true as boolean type if exist", () => {
      const url = `${baseUrl}?descending=true`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res.descending).toEqual(true);
    });

    it("should return object contains descending false as boolean type if exist", () => {
      const url = `${baseUrl}?descending=false`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res.descending).toEqual(false);
    });

    it("should return object contains page as number if exist", () => {
      const url = `${baseUrl}?page=2`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res.page).toEqual(2);
    });

    it("should return object contains rowsPerPage as number if exist", () => {
      const url = `${baseUrl}?rowsPerPage=20`;
      setLocation(url);

      const res = urlService.getFiltersObj(window.location);
      expect(res.rowsPerPage).toEqual(20);
    });
  });

  describe("toSearchParams", () => {
    it("should return an empty string", () => {
      const res = urlService.toSearchParams({});
      expect(res).toEqual("");
    });

    it("should return string contains sortBy", () => {
      const res = urlService.toSearchParams({ sortBy: "puki" });
      expect(res).toContain("sortBy=puki");
    });

    it("should return string contains sortBy and page", () => {
      const res = urlService.toSearchParams({ sortBy: "puki", page: 1 });
      expect(res).toContain("sortBy=puki");
      expect(res).toContain("page=1");
      expect(res).toContain("&");
    });

    it("should return string contains query", () => {
      const res = urlService.toSearchParams({ query: "puki" });
      expect(res).toContain("query=puki");
    });

    it("should return string contains descending", () => {
      const res1 = urlService.toSearchParams({ descending: false });
      expect(res1).toContain("descending=false");

      const res2 = urlService.toSearchParams({ descending: true });
      expect(res2).toContain("descending=true");
    });

    it("should return string contains rowsPerPage", () => {
      const res1 = urlService.toSearchParams({ rowsPerPage: 20 });
      expect(res1).toContain("rowsPerPage=20");
    });

    it("should stringify nested objects correctly", () => {
      const filterBy = {
        name: "John",
        age: 30,
        columnFilters: [
          { column: "city", value: "New York" },
          { column: "country", value: "USA" },
        ],
        advancedFilters: [
          { type: "FreeText", name: "type", title: "Type", term: "blabla", value: "=@" },
          { type: "FreeText", name: "name", title: "Workload", term: "ws", value: "=@" },
        ],
      };

      const expectedSearchParams =
        'name=John&age=30&columnFilters=[{"column":"city","value":"New+York"},{"column":"country","value":"USA"}]&advancedFilters=[{"type":"FreeText","name":"type","title":"Type","term":"blabla","value":"=@"},{"type":"FreeText","name":"name","title":"Workload","term":"ws","value":"=@"}]';
      const result = urlService.toSearchParams(filterBy as unknown as IFilterBy);

      expect(result).toEqual(expectedSearchParams);
    });
  });

  describe("updateQueryParams", () => {
    const href = "http://example.com/?param1=value1&param2=value2";

    beforeEach(() => {
      global.window = Object.create(window);
      setLocation(href);
      vi.spyOn(history, "replaceState").mockImplementation(() => {
        // do nothing
      });
    });

    afterEach(() => {
      vi.resetAllMocks();
    });

    afterAll(() => {
      vi.restoreAllMocks();
    });

    it("should update existing query parameters", () => {
      urlService.updateQueryParams({ param1: "newValue" });

      expect(history.replaceState).toHaveBeenCalledWith(null, "", "http://example.com/?param1=newValue&param2=value2");
    });

    it("should add new query parameters", () => {
      urlService.updateQueryParams({ param3: "value3" });

      expect(history.replaceState).toHaveBeenCalledWith(
        null,
        "",
        "http://example.com/?param1=value1&param2=value2&param3=value3",
      );
    });

    it("should remove query parameters when value is null", () => {
      urlService.updateQueryParams({ param1: null });

      expect(history.replaceState).toHaveBeenCalledWith(null, "", "http://example.com/?param2=value2");
    });

    it("should handle multiple query parameters at once", () => {
      urlService.updateQueryParams({ param1: "newValue1", param2: "newValue2", param3: "value3" });

      expect(history.replaceState).toHaveBeenCalledWith(
        null,
        "",
        "http://example.com/?param1=newValue1&param2=newValue2&param3=value3",
      );
    });

    it("should not modify URL if no query parameters are passed", () => {
      urlService.updateQueryParams({});

      expect(history.replaceState).toHaveBeenCalledWith(null, "", href);
    });
  });
});
