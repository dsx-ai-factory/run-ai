/**
 * Utility functions for guided tours
 */

/**
 * Waits for a condition to be met with a timeout
 * @param condition - Function that returns true when condition is met
 * @param options - Configuration options for the timeout
 * @returns Promise<boolean> - True if condition was met, false if timeout reached
 */
export async function waitForCondition(
  condition: () => boolean | Promise<boolean>,
  options: {
    timeout?: number;
    interval?: number;
    description?: string;
  } = {},
): Promise<boolean> {
  const {
    timeout = 10000, // Default 10 seconds
    interval = 100, // Default check every 100ms
    description = "condition",
  } = options;

  const startTime = Date.now();

  while (Date.now() - startTime < timeout) {
    try {
      const result = await condition();
      if (result) {
        return true;
      }
    } catch (error) {
      console.warn(`[Guided Tour] Error checking ${description}:`, error);
    }

    // Wait before checking again
    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  console.warn(`[Guided Tour] Timeout waiting for ${description} after ${timeout}ms`);
  return false;
}

/**
 * Waits for an element to exist in the DOM
 * @param selector - CSS selector for the element
 * @param options - Timeout configuration
 * @returns Promise<boolean> - True if element found, false if timeout reached
 */
export async function waitForElement(
  selector: string,
  options: {
    timeout?: number;
    interval?: number;
  } = {},
): Promise<boolean> {
  return waitForCondition(() => !!document.querySelector(selector), {
    ...options,
    description: `element with selector "${selector}"`,
  });
}

/**
 * Waits for a function to return a truthy value
 * @param fn - Function to check
 * @param options - Timeout configuration
 * @returns Promise<boolean> - True if function returns truthy value, false if timeout reached
 */
export async function waitForFunction<T>(
  fn: () => T | Promise<T>,
  options: {
    timeout?: number;
    interval?: number;
    description?: string;
  } = {},
): Promise<boolean> {
  return waitForCondition(
    async () => {
      const result = await fn();
      return !!result;
    },
    {
      ...options,
      description: options.description || "function to return truthy value",
    },
  );
}
