export const objectUtil = {
  isEmpty,
  invertObject,
};

export function isEmpty(object: object | undefined | null): boolean {
  return !object || Object.keys(object).length === 0;
}

export function invertObject(obj: Record<string, string>): Record<string, string> {
  const inverted: Record<string, string> = {};
  for (const key in obj) {
    // eslint-disable-next-line no-prototype-builtins
    if (obj.hasOwnProperty(key)) {
      inverted[obj[key]] = key;
    }
  }
  return inverted;
}
