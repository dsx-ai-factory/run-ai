// @ts-nocheck
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  camelStyleToUnderscore,
  previewStrings,
  convertToFlatArray,
  arrayToObjectMap,
  omit,
  pick,
  haveSameElements,
  openNewWindow,
} from "./common.util";

describe("common.util", () => {
  describe("camelStyleToUnderscore", () => {
    it("Should convert camel case to underscore", async () => {
      expect(camelStyleToUnderscore("camelCase")).toEqual("camel_case");
      expect(camelStyleToUnderscore("camelcase")).toEqual("camelcase");
      expect(camelStyleToUnderscore("camel_case")).toEqual("camel_case");
    });
  });

  describe("previewStrings", () => {
    it("return '-' when array is an empty or undefined", async () => {
      expect(previewStrings([])).toEqual("-");
      expect(previewStrings()).toEqual("-");
    });

    test("it should not return 0 when + more is 0", () => {
      expect(previewStrings(["running", "creating"], 2)).not.includes("0");
    });

    it("return one word when array have only one item", async () => {
      expect(previewStrings(["aaa"])).toEqual("aaa");
    });

    it("return one word and '+ XXX more' when array have more than the default min words to display", () => {
      expect(previewStrings(["aaa", "bbb"])).toEqual("aaa + 1 more");
      expect(previewStrings(["aaa", "bbb", "ccc"])).toEqual("aaa + 2 more");
    });

    it("return min words and '+ XXX more' as the rest of the length", () => {
      expect(previewStrings(["aaa", "bbb", "ccc"], 2)).toEqual("aaa, bbb + 1 more");
      expect(previewStrings(["aaa", "bbb", "ccc", "ddd"], 2)).toEqual("aaa, bbb + 2 more");
    });

    it("return all words without the '+ XXX more' if there are less words in the array than the min to display", () => {
      expect(previewStrings(["aaa", "bbb", "ccc", "ddd"], 5)).toEqual("aaa, bbb, ccc, ddd");
    });
  });

  describe("convertToFlatArray", () => {
    it("Should convert object to array of string", async () => {
      const data = {
        name: "John Doe",
        age: 30,
        hobbies: ["reading", "writing"],
        address: {
          street: "123 Main St",
          city: "Anytown",
          state: "CA",
        },
      };

      const result = convertToFlatArray(data);
      expect(result).toEqual(["John Doe", "30", "reading", "writing", "123 Main St", "Anytown", "CA"]);
    });

    it("Should keep array as array", async () => {
      const data = ["John Doe", "30", "reading", "writing", "123 Main St", { adress: "some street" }];
      const result = convertToFlatArray(data);
      expect(result).toEqual(["John Doe", "30", "reading", "writing", "123 Main St", "some street"]);
    });

    it("Should convert string to array of strings", async () => {
      const data = "John Doe";
      const result = convertToFlatArray(data);
      expect(result).toEqual(["John Doe"]);
    });

    it("Should convert number to array of strings", async () => {
      const data = 1;
      const result = convertToFlatArray(data);
      expect(result).toEqual(["1"]);
    });

    it("Should convert boolean to array of strings", async () => {
      const result = convertToFlatArray(true);
      expect(result).toEqual(["true"]);

      const result2 = convertToFlatArray(false);
      expect(result2).toEqual(["false"]);
    });
  });

  describe("arrayToObjectMap", () => {
    it("return object map by 'id' when sending prop name as string", () => {
      const arr = [
        {
          id: "649c2603fdfcb25859d36194",
          name: "Wolf Goodwin",
          gender: "male",
          company: "DOGTOWN",
          email: "wolfgoodwin@dogtown.com",
        },
        {
          id: "649c260330229d5f17921c06",
          name: "Bailey Kent",
          gender: "male",
          company: "ENJOLA",
          email: "baileykent@enjola.com",
        },
      ];

      const res = arrayToObjectMap(arr, "id");
      expect(res[arr[0].id]).toBe(arr[0]);
    });

    it("return object map by 'name' when sending prop name as string", () => {
      const arr = [
        {
          id: "649c2603fdfcb25859d36194",
          name: "Wolf Goodwin",
          gender: "male",
          company: "DOGTOWN",
          email: "wolfgoodwin@dogtown.com",
        },
        {
          id: "649c260330229d5f17921c06",
          name: "Bailey Kent",
          gender: "male",
          company: "ENJOLA",
          email: "baileykent@enjola.com",
        },
      ];

      const res = arrayToObjectMap(arr, "name");
      expect(res[arr[0].name]).toBe(arr[0]);
    });

    it("return object map by 'id' when sending prop name as function", () => {
      const arr = [
        {
          meta: {
            id: "649c2603fdfcb25859d36194",
            name: "Wolf Goodwin",
          },
          spec: {
            gender: "male",
            company: "DOGTOWN",
            email: "wolfgoodwin@dogtown.com",
          },
        },
        {
          meta: {
            id: "649c260330229d5f17921c06",
            name: "Bailey Kent",
          },
          spec: {
            gender: "male",
            company: "ENJOLA",
            email: "baileykent@enjola.com",
          },
        },
      ];

      const res = arrayToObjectMap(arr, (item) => item.meta.id);
      expect(res[arr[0].meta.id]).toBe(arr[0]);
    });
  });

  describe("pick", () => {
    const person = {
      name: "John",
      age: 30,
      address: "123 Main St",
    };

    it("picks a single key from the object", () => {
      const pickedPerson = pick(person, "age");
      expect(pickedPerson).toEqual({ age: 30 });
    });

    it("picks multiple keys from the object", () => {
      const pickedPerson = pick(person, "age", "address");
      expect(pickedPerson).toEqual({ age: 30, address: "123 Main St" });
    });

    it("returns a new object without modifying the original", () => {
      const pickedPerson = pick(person, "age");
      expect(pickedPerson).toEqual({ age: 30 });
      expect(person).toEqual({ name: "John", age: 30, address: "123 Main St" });
    });

    it("returns an empty object when a key does not exist", () => {
      const pickedPerson = pick(person, "home");
      expect(Object.keys(pickedPerson).length).toBe(0);
    });
  });

  describe("omit", () => {
    const person = {
      name: "John",
      age: 30,
      address: "123 Main St",
    };

    it("omits a single key from the object", () => {
      const omittedPerson = omit(person, ["age"]);
      expect(omittedPerson).toEqual({ name: "John", address: "123 Main St" });
    });

    it("omits multiple keys from the object", () => {
      const omittedPerson = omit(person, ["age", "address"]);
      expect(omittedPerson).toEqual({ name: "John" });
    });

    it("returns a new object without modifying the original", () => {
      const omittedPerson = omit(person, ["age"]);
      expect(omittedPerson).toEqual({ name: "John", address: "123 Main St" });
      expect(person).toEqual({ name: "John", age: 30, address: "123 Main St" });
    });
  });

  describe("haveSameElements", () => {
    it("should return true when arrays have the same elements in different order", () => {
      const arr1 = [1, 2, 3, 4];
      const arr2 = [4, 3, 2, 1];
      expect(haveSameElements(arr1, arr2)).toBe(true);
    });

    it("should return false when arrays have different elements", () => {
      const arr1 = [1, 2, 3, 4];
      const arr2 = [1, 2, 3, 5];
      expect(haveSameElements(arr1, arr2)).toBe(false);
    });
  });

  describe("openNewWindow", () => {
    beforeEach(() => {
      vi.stubGlobal("open", vi.fn());
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it.each([
      ["https://example.com/tool", "https://example.com/tool"],
      ["HTTP://example.com/tool", "HTTP://example.com/tool"],
      ["/roles/123", "/roles/123"],
      ["//example.com/tool", "//example.com/tool"],
      ["example.com/tool", "http://example.com/tool"],
      ["localhost:8080/tool", "http://localhost:8080/tool"],
    ])("opens an allowed URL %s", (url, expectedUrl) => {
      openNewWindow(url);

      expect(window.open).toHaveBeenCalledWith(expectedUrl, "_blank", "noopener,noreferrer");
    });

    it.each([
      "javascript:alert(document.domain)",
      "data:text/html,<script>alert(1)</script>",
      "ftp://example.com/file",
      "httpsomething:payload",
      "https://[invalid",
      "   ",
    ])("rejects an unsafe or malformed URL %s", (url) => {
      expect(openNewWindow(url)).toBeNull();
      expect(window.open).not.toHaveBeenCalled();
    });
  });
});
