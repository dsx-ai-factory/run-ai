/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it } from "vitest";
import { urlService } from "@/services/url.service/url.service";
import type { IFilterBy } from "@/models/filter.model";
import { createMemoryHistory, createRouter } from "vue-router";

describe("url.service", () => {
  describe("isControlPlaneRouteUrl", () => {
    const controlPlaneUrl = "https://runai.example.com/workloads";
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/", component: {} },
        { path: "/workload-templates/:id?", component: {} },
        { path: "/login/callback", component: {} },
        { path: "/:catchAll(.*)", redirect: "/login/callback" },
      ],
    });

    it.each([
      ["control-plane root", "https://runai.example.com/"],
      ["registered route", "https://runai.example.com/workload-templates"],
      ["route with a parameter", "https://runai.example.com/workload-templates/template-1"],
      ["unregistered child path", "https://runai.example.com/workload-templates/custom/tool"],
      ["parent of a registered route", "https://runai.example.com/login"],
      ["route with query and fragment", "https://runai.example.com/login/callback?next=/tool#details"],
      ["different scheme and port", "http://runai.example.com:8080/workload-templates"],
      ["case-insensitive hostname", "https://RUNAI.EXAMPLE.COM/workload-templates"],
      ["encoded route", "https://runai.example.com/%77orkload-templates/tool"],
      ["uppercase route", "https://runai.example.com/WORKLOAD-TEMPLATES"],
      ["mixed-case route child", "https://runai.example.com/Workload-Templates/tool"],
      ["malformed suffix", "https://runai.example.com/workload-templates/%FF"],
      ["encoded prefix with malformed suffix", "https://runai.example.com/%77orkload-templates/%FF"],
      ["encoded separator with malformed suffix", "https://runai.example.com/%77orkload-templates%2F%FF"],
      ["normalized dot segments", "https://runai.example.com/project/../workload-templates"],
    ])("detects a collision for %s", (_, url) => {
      expect(urlService.isControlPlaneRouteUrl(url, router.getRoutes(), controlPlaneUrl)).toBe(true);
    });

    it.each([
      ["another hostname", "https://tools.example.com/workload-templates"],
      ["workload hostname", "https://tool.runai.example.com/"],
      ["unrelated path", "https://runai.example.com/project/tool"],
      ["similar prefix", "https://runai.example.com/workload-templates-custom"],
      ["uppercase similar prefix", "https://runai.example.com/WORKLOAD-TEMPLATES-CUSTOM"],
      ["unrelated path with malformed suffix", "https://runai.example.com/project/%FF"],
      ["route name only in query", "https://runai.example.com/tool?next=/workload-templates"],
      ["empty value", ""],
      ["invalid URL", "not a URL"],
    ])("preserves existing handling for %s", (_, url) => {
      expect(urlService.isControlPlaneRouteUrl(url, router.getRoutes(), controlPlaneUrl)).toBe(false);
    });

    it("protects routes added at runtime, including aliases", () => {
      const removeRoute = router.addRoute({ path: "/new-page", alias: "/new-alias", component: {} });
      for (const path of ["/new-page/child", "/new-alias/child"]) {
        expect(
          urlService.isControlPlaneRouteUrl(`https://runai.example.com${path}`, router.getRoutes(), controlPlaneUrl),
        ).toBe(true);
      }
      removeRoute();
    });
  });

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
