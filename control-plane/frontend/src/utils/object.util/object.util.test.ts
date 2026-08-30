import { isEmpty } from "./object.util";
import { invertObject } from "./object.util";

describe("object.util", () => {
  describe("isEmpty", () => {
    it("Should check if object empty", () => {
      expect(isEmpty({})).toBeTruthy();
      expect(isEmpty({ test: 1 })).toBeFalsy();
      expect(isEmpty(undefined)).toBeTruthy();
      expect(isEmpty(null)).toBeTruthy();
    });
  });
  describe("invertObject", () => {
    it("should invert an object with unique values", () => {
      const obj = { a: "1", b: "2", c: "3" };
      const expected = { "1": "a", "2": "b", "3": "c" };
      expect(invertObject(obj)).toEqual(expected);
    });

    it("should handle objects with duplicate values by keeping the last key", () => {
      const obj = { a: "1", b: "2", c: "1" };
      const expected = { "1": "c", "2": "b" };
      expect(invertObject(obj)).toEqual(expected);
    });

    it("should return an empty object when given an empty object", () => {
      const obj = {};
      const expected = {};
      expect(invertObject(obj)).toEqual(expected);
    });

    it("should handle an object with one key-value pair", () => {
      const obj = { a: "1" };
      const expected = { "1": "a" };
      expect(invertObject(obj)).toEqual(expected);
    });

    it("should not modify the original object", () => {
      const obj = { a: "1", b: "2" };
      const copy = { ...obj };
      invertObject(obj);
      expect(obj).toEqual(copy);
    });
  });
});
