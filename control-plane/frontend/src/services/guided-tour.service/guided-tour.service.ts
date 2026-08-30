import type {
  TourConfig,
  TourState,
  TourStep,
  TourSide,
  TourAlignment,
  MenuAnchor,
} from "@/components/common/guided-tour-popup/guided-tour.types";

/**
 * Clears all active event condition timeout loops
 */
function clearActiveEventConditionLoops(state: TourState): void {
  if (state.activeEventConditionLoops) {
    state.activeEventConditionLoops.forEach((controller) => controller.abort());
    state.activeEventConditionLoops.clear();
  }
}

/**
 * Monitors DOM changes to detect when target elements appear/disappear
 */
function startElementMonitoring(
  step: TourStep,
  config: TourConfig,
  state: TourState,
  onElementChange: (newElements: Element[], removedElements: Element[]) => void,
): MutationObserver | null {
  if (!step.nextOn?.selector) return null;

  const selectors = Array.isArray(step.nextOn.selector) ? step.nextOn.selector : [step.nextOn.selector];
  const attachToAllElements = step.nextOn.attachToAllElements || false;

  const observer = new MutationObserver((mutations) => {
    let hasRelevantChanges = false;
    const addedNodes: Node[] = [];
    const removedNodes: Node[] = [];

    mutations.forEach((mutation) => {
      if (mutation.type === "childList") {
        mutation.addedNodes.forEach((node) => {
          if (node.nodeType === Node.ELEMENT_NODE) {
            addedNodes.push(node);
            hasRelevantChanges = true;
          }
        });
        mutation.removedNodes.forEach((node) => {
          if (node.nodeType === Node.ELEMENT_NODE) {
            removedNodes.push(node);
            hasRelevantChanges = true;
          }
        });
      }
    });

    if (!hasRelevantChanges) return;

    // Check if any of the added/removed nodes match our selectors
    const newElements: Element[] = [];
    const removedElements: Element[] = [];

    selectors.forEach((selector) => {
      // Check added nodes
      addedNodes.forEach((node) => {
        const element = node as Element;
        if (element.matches && element.matches(selector)) {
          newElements.push(element);
        }
        // Also check descendants
        const descendants = element.querySelectorAll?.(selector);
        if (descendants) {
          newElements.push(...Array.from(descendants));
        }
      });

      // Check removed nodes
      removedNodes.forEach((node) => {
        const element = node as Element;
        if (element.matches && element.matches(selector)) {
          removedElements.push(element);
        }
        // Also check descendants
        const descendants = element.querySelectorAll?.(selector);
        if (descendants) {
          removedElements.push(...Array.from(descendants));
        }
      });

      // Re-scan the entire document for new elements that might have appeared
      const currentElements = attachToAllElements
        ? Array.from(document.querySelectorAll(selector))
        : ([document.querySelector(selector)].filter(Boolean) as Element[]);

      // Find elements that weren't in our previous tracking
      const previousElements = state.trackedElements || new Set();
      currentElements.forEach((el) => {
        if (!previousElements.has(el)) {
          newElements.push(el);
        }
      });

      // Find elements that are no longer in the DOM
      previousElements.forEach((el) => {
        if (!document.contains(el)) {
          removedElements.push(el);
        }
      });
    });

    if (newElements.length > 0 || removedElements.length > 0) {
      onElementChange(newElements, removedElements);
    }
  });

  // Start observing
  observer.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: false,
    characterData: false,
  });

  return observer;
}

/**
 * Stops element monitoring
 */
function stopElementMonitoring(observer: MutationObserver | null): void {
  if (observer) {
    observer.disconnect();
  }
}

/**
 * Re-attaches event listeners to newly appeared elements
 */
function reattachListenersToElements(
  elements: Element[],
  step: TourStep,
  config: TourConfig,
  state: TourState,
  clearListeners: () => void,
  hideTour: () => void,
  moveNext: () => void,
): { element: Element; event: string; handler: (event: Event) => void }[] {
  if (!step.nextOn || elements.length === 0) return [];

  const { event, eventCondition, onAdvance, timeout, checkInterval } = step.nextOn;
  const listeners: { element: Element; event: string; handler: (event: Event) => void }[] = [];

  const handler = async (e: Event) => {
    const currentTarget = e.currentTarget as Element;

    if (eventCondition) {
      let canAdvance = false;

      if (timeout) {
        // Clear any existing timeout loops for this tour
        clearActiveEventConditionLoops(state);
        state.activeEventConditionLoops = new Set();

        // Create a new AbortController for this timeout loop
        const abortController = new AbortController();
        state.activeEventConditionLoops.add(abortController);

        // Implement timeout logic for eventCondition
        const maxWaitTime = timeout;
        const interval = checkInterval || 100;
        const startTime = Date.now();

        try {
          while (Date.now() - startTime < maxWaitTime && !abortController.signal.aborted) {
            canAdvance = await Promise.resolve(eventCondition(e, currentTarget));
            if (canAdvance) {
              break;
            }

            // Wait a bit before checking again, but check for abort signal
            await new Promise((resolve, reject) => {
              const timeoutId = setTimeout(resolve, interval);
              abortController.signal.addEventListener("abort", () => {
                clearTimeout(timeoutId);
                reject(new Error("Aborted"));
              });
            });
          }
        } catch (error) {
          if (error instanceof Error && error.message === "Aborted") {
            // Loop was aborted, don't advance
            return;
          }
          throw error;
        } finally {
          // Remove this controller from the active set
          state.activeEventConditionLoops?.delete(abortController);
        }
      } else {
        // No timeout, just call the condition once
        canAdvance = await Promise.resolve(eventCondition(e, currentTarget));
      }

      if (!canAdvance) {
        return;
      }
    }

    clearListeners();

    // Clear any active event condition loops when advancing
    clearActiveEventConditionLoops(state);

    if (onAdvance) {
      // Create minimal context for type compatibility - user code should use store directly
      const context = { config, state };
      await onAdvance(e, currentTarget, step, context);
      return;
    }

    if (step.nextOn?.hideOnAction) {
      hideTour();
      moveNext();
    } else {
      moveNext();
    }
  };

  elements.forEach((el) => {
    el.addEventListener(event, handler);
    listeners.push({ element: el, event, handler });
  });

  return listeners;
}

/**
 * Re-applies styling to newly appeared elements
 */
function reapplyStylingToElements(elements: Element[], step: TourStep, config: TourConfig): void {
  elements.forEach((element) => {
    // Apply main element styling
    element.classList.add("guided-tour-active-element");
    element.setAttribute("aria-haspopup", "dialog");
    element.setAttribute("aria-expanded", "true");
    element.setAttribute("aria-controls", "guided-tour-popover-content");

    const disableInteraction = step.disableActiveInteraction ?? config.disableActiveInteraction;
    if (disableInteraction) {
      element.classList.add("guided-tour-no-interaction");
    }
  });

  // Handle accessible elements
  if (step.accessibleElements && step.accessibleElements.length > 0) {
    step.accessibleElements.forEach((selector) => {
      const accessibleElement = document.querySelector(selector);
      if (accessibleElement) {
        accessibleElement.classList.add("guided-tour-accessible-element");
        accessibleElement.setAttribute("data-guided-tour-accessible", "true");
      }
    });
  }
}

/**
 * Removes styling from elements that are no longer relevant
 */
function removeStylingFromElements(elements: Element[]): void {
  elements.forEach((element) => {
    element.classList.remove("guided-tour-active-element", "guided-tour-no-interaction");
    element.removeAttribute("aria-haspopup");
    element.removeAttribute("aria-expanded");
    element.removeAttribute("aria-controls");
    element.removeAttribute("data-guided-tour-target");
  });
}

export const guidedTourService = {
  // DOM utilities
  getTargetElement,
  isElementInView,
  bringIntoView,
  generateElementSelector,
  cleanupActiveElements,

  // Element styling and interaction
  applyElementStyling,

  // Popover positioning
  mapSideToAnchor,
  mapSideToSelf,

  // Animation and overlay
  updateGradientPosition,
  easeInOutQuad,
  animateOverlay,
  animateTransition,
  mountOverlay,
  updateOverlay,
  createOverlay,
  removeOverlay,
  generateCutoutSvgPath,

  // Event handling
  setupNextOn,
  handleBeforeShowStep,

  // Element waiting
  waitForElement,

  // Scroll locking
  lockScrolling,
  unlockScrolling,
  setupElementScrollAllowance,
  removeElementScrollAllowance,

  // Advanced scroll control
  forceLockScrolling,
  forceUnlockScrolling,
  allowElementScrolling,

  // Tour event handling
  cleanupTourEventHandlers,

  // Dynamic element monitoring
  startElementMonitoring,
  stopElementMonitoring,
  reattachListenersToElements,
  reapplyStylingToElements,
  removeStylingFromElements,

  // Utility functions
  clearActiveEventConditionLoops,
};

/**
 * Gets the target element for a tour step
 */
function getTargetElement(element?: string | Element | (() => Element)): Element {
  if (!element) {
    return document.body;
  }

  if (typeof element === "function") {
    try {
      const result = element();
      if (!result) {
        throw new Error("Element function returned null/undefined");
      }
      return result;
    } catch (error) {
      console.error("[Guided Tour] Element function threw an error:", error);
      throw new Error(`Failed to execute element function: ${error}`);
    }
  }

  if (typeof element === "string") {
    const foundElement = document.querySelector(element);
    if (!foundElement) {
      const error = `[Guided Tour] Critical: Target element not found: ${element}`;
      console.error(error);
      throw new Error(error);
    }
    return foundElement;
  }

  if (!element) {
    throw new Error("[Guided Tour] Critical: Element is null/undefined");
  }

  return element;
}

/**
 * Checks if element is currently in view
 */
function isElementInView(element: Element): boolean {
  const rect = element.getBoundingClientRect();
  const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
  const viewportWidth = window.innerWidth || document.documentElement.clientWidth;

  const verticalThreshold = Math.min(50, rect.height * 0.2);
  const horizontalThreshold = Math.min(50, rect.width * 0.2);

  return (
    rect.top >= -verticalThreshold &&
    rect.left >= -horizontalThreshold &&
    rect.bottom <= viewportHeight + verticalThreshold &&
    rect.right <= viewportWidth + horizontalThreshold
  );
}

/**
 * Gets the available viewport accounting for UI elements like headers, popups, etc.
 */
function getAvailableViewport(config?: TourConfig) {
  const viewportHeight = window.innerHeight;
  const viewportWidth = window.innerWidth;

  // Get custom adjustments from config
  const adjustments = config?.viewportAdjustments || {};

  // Account for common UI elements that might reduce available space
  const headerHeight = adjustments.headerHeight ?? 60; // Typical header height
  const footerHeight = adjustments.footerHeight ?? 0; // Footer height if any
  const popupWidth = adjustments.popupWidth ?? 320; // Guided tour popup width
  const popupHeight = adjustments.popupHeight ?? 0; // Popup height if it affects vertical space

  // Check for actual fixed elements in the DOM
  const header = document.querySelector('header, .header, [role="banner"]');
  const actualHeaderHeight = header ? header.getBoundingClientRect().height : headerHeight;

  return {
    height: viewportHeight - actualHeaderHeight - footerHeight - popupHeight,
    width: viewportWidth - popupWidth,
    offsetTop: actualHeaderHeight,
    offsetLeft: 0,
    fullHeight: viewportHeight,
    fullWidth: viewportWidth,
  };
}

/**
 * Brings element into view with appropriate scrolling behavior
 *
 * The function supports different scroll behaviors based on the config.scrollBehavior setting:
 * - "center" (default): Centers the element in the viewport with fine-tuning for perfect centering
 * - "start": Aligns the element to the top of the viewport
 * - "end": Aligns the element to the bottom of the viewport
 * - "nearest": Scrolls the minimum amount to bring the element into view
 * - "smart": Automatically chooses between center and start based on element height
 *
 * For the "center" behavior, the function uses a two-step approach:
 * 1. First scrolls using scrollIntoView with center alignment
 * 2. Then fine-tunes the position for perfect centering
 *
 * @param element - The element to bring into view
 * @param config - Tour configuration containing scroll behavior settings
 */
function bringIntoView(element: Element | null, config: TourConfig): void {
  if (config.debug) console.log("[Guided Tour] bringIntoView function called", { element: !!element, config: !!config });

  if (!element) {
    return;
  }

  // For center behavior, always attempt centering even if element is "in view"
  // because "in view" doesn't mean "centered"
  const scrollBehavior = config.scrollBehavior || "center";
  const shouldAlwaysCenter = scrollBehavior === "center";

  if (!shouldAlwaysCenter && isElementInView(element)) {
    return;
  }

  const shouldSmoothScroll = config.smoothScroll;
  const rect = element.getBoundingClientRect();
  const viewportHeight = window.innerHeight;
  const viewportWidth = window.innerWidth;

  // Debug logging
  if (config.debug) {
    console.log("[Guided Tour] bringIntoView called:", {
      element:
        element.tagName +
        (element.id ? `#${element.id}` : "") +
        (element.className ? `.${element.className.split(" ").join(".")}` : ""),
      scrollBehavior,
      shouldAlwaysCenter,
      isInView: isElementInView(element),
      elementRect: rect,
      viewport: { width: viewportWidth, height: viewportHeight },
    });
  }

  // Determine if we need to scroll and in which direction
  const needsVerticalScroll = rect.top < 0 || rect.bottom > viewportHeight;
  const needsHorizontalScroll = rect.left < 0 || rect.right > viewportWidth;

  if (!needsVerticalScroll && !needsHorizontalScroll) {
    return;
  }

  // Handle different scroll behaviors
  switch (scrollBehavior) {
    case "start":
      element.scrollIntoView({
        behavior: shouldSmoothScroll ? "smooth" : "auto",
        inline: "nearest",
        block: "start",
      });
      break;

    case "end":
      element.scrollIntoView({
        behavior: shouldSmoothScroll ? "smooth" : "auto",
        inline: "nearest",
        block: "end",
      });
      break;

    case "nearest":
      element.scrollIntoView({
        behavior: shouldSmoothScroll ? "smooth" : "auto",
        inline: "nearest",
        block: "nearest",
      });
      break;

    case "smart": {
      // Smart behavior: center by default, but use start for very tall elements
      const elementHeight = rect.height;
      const blockBehavior = elementHeight > viewportHeight * 0.8 ? "start" : "center";
      element.scrollIntoView({
        behavior: shouldSmoothScroll ? "smooth" : "auto",
        inline: "center",
        block: blockBehavior,
      });
      break;
    }

    case "center":
    default: {
      // Default centering behavior with improved fine-tuning
      element.scrollIntoView({
        behavior: shouldSmoothScroll ? "smooth" : "auto",
        inline: "center",
        block: "center",
      });

      // Improved fine-tuning with better timing and viewport accounting
      const performFineTuning = () => {
        const newRect = element.getBoundingClientRect();
        const currentScrollTop = window.pageYOffset || document.documentElement.scrollTop;
        const currentScrollLeft = window.pageXOffset || document.documentElement.scrollLeft;

        // Get available viewport accounting for UI elements
        const availableViewport = getAvailableViewport(config);

        // Check if element is actually centered with tolerance
        const verticalTolerance = 20; // Allow 20px tolerance
        const horizontalTolerance = 20;

        const targetTop = availableViewport.offsetTop + availableViewport.height / 2 - newRect.height / 2;
        const targetLeft = availableViewport.offsetLeft + availableViewport.width / 2 - newRect.width / 2;

        const isVerticallyCentered = Math.abs(newRect.top - targetTop) <= verticalTolerance;
        const isHorizontallyCentered = Math.abs(newRect.left - targetLeft) <= horizontalTolerance;

        // Debug logging if enabled
        if (config.debug) {
          console.log("[Guided Tour] Centering debug:", {
            elementRect: newRect,
            viewport: { width: window.innerWidth, height: window.innerHeight },
            availableViewport,
            targetPosition: { top: targetTop, left: targetLeft },
            currentPosition: { top: newRect.top, left: newRect.left },
            isCentered: { vertical: isVerticallyCentered, horizontal: isHorizontallyCentered },
            scrollPosition: { x: currentScrollLeft, y: currentScrollTop },
          });
        }

        if (isVerticallyCentered && isHorizontallyCentered) {
          return; // Already centered within tolerance
        }

        // Calculate adjustments
        let finalScrollTop = currentScrollTop;
        let finalScrollLeft = currentScrollLeft;

        // Vertical adjustment
        if (needsVerticalScroll && !isVerticallyCentered) {
          const adjustment = newRect.top - targetTop;
          finalScrollTop = currentScrollTop - adjustment;
        }

        // Horizontal adjustment
        if (needsHorizontalScroll && !isHorizontallyCentered) {
          const adjustment = newRect.left - targetLeft;
          finalScrollLeft = currentScrollLeft - adjustment;
        }

        // Apply final scroll position
        if (finalScrollTop !== currentScrollTop || finalScrollLeft !== currentScrollLeft) {
          if (config.debug) {
            console.log("[Guided Tour] Applying fine-tuning scroll:", {
              from: { x: currentScrollLeft, y: currentScrollTop },
              to: { x: finalScrollLeft, y: finalScrollTop },
            });
          }

          window.scrollTo({
            top: finalScrollTop,
            left: finalScrollLeft,
            behavior: shouldSmoothScroll ? "smooth" : "auto",
          });
        }
      };

      if (shouldSmoothScroll) {
        // Wait for smooth scroll to complete, then fine-tune
        setTimeout(() => {
          requestAnimationFrame(performFineTuning);

          // Add fallback if centering still fails after fine-tuning
          setTimeout(() => {
            const finalRect = element.getBoundingClientRect();
            const availableViewport = getAvailableViewport(config);
            const targetTop = availableViewport.offsetTop + availableViewport.height / 2 - finalRect.height / 2;
            const targetLeft = availableViewport.offsetLeft + availableViewport.width / 2 - finalRect.width / 2;

            const isStillNotCentered =
              Math.abs(finalRect.top - targetTop) > 50 || Math.abs(finalRect.left - targetLeft) > 50;

            if (isStillNotCentered && config.debug) {
              console.warn("[Guided Tour] Centering failed, trying fallback scroll behavior");
              // Try smart behavior as fallback
              element.scrollIntoView({
                behavior: "smooth",
                inline: "center",
                block: "start", // Use start instead of center for fallback
              });
            }
          }, 300); // Additional delay for fallback check
        }, 200); // Increased delay for smooth scrolling
      } else {
        // Immediate fine-tuning for non-smooth scroll
        requestAnimationFrame(performFineTuning);
      }
      break;
    }
  }
}

/**
 * Generates a unique selector for an element
 */
function generateElementSelector(element: Element): string {
  const aid = element.getAttribute("aid");
  if (aid) {
    return `[aid="${aid}"]`;
  }

  if (element.id) {
    return `#${element.id}`;
  }

  const testId = element.getAttribute("data-testid");
  if (testId) {
    return `[data-testid="${testId}"]`;
  }

  const uniqueId = `guided-tour-target-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  element.setAttribute("data-guided-tour-target", uniqueId);
  return `[data-guided-tour-target="${uniqueId}"]`;
}

/**
 * Cleans up active element styling
 */
function cleanupActiveElements(): void {
  // Clean up main active element
  document.querySelectorAll(".guided-tour-active-element").forEach((el) => {
    el.classList.remove("guided-tour-active-element", "guided-tour-no-interaction");
    el.removeAttribute("aria-haspopup");
    el.removeAttribute("aria-expanded");
    el.removeAttribute("aria-controls");
    el.removeAttribute("data-guided-tour-target");
  });

  // Clean up accessible elements
  document.querySelectorAll(".guided-tour-accessible-element").forEach((el) => {
    el.classList.remove("guided-tour-accessible-element");
    el.removeAttribute("data-guided-tour-accessible");
  });
}

/**
 * Applies styling and accessibility attributes to the active element
 */
function applyElementStyling(element: Element, step: TourStep, config: TourConfig): void {
  cleanupActiveElements();

  element.classList.add("guided-tour-active-element");
  element.setAttribute("aria-haspopup", "dialog");
  element.setAttribute("aria-expanded", "true");
  element.setAttribute("aria-controls", "guided-tour-popover-content");

  const disableInteraction = step.disableActiveInteraction ?? config.disableActiveInteraction;
  if (disableInteraction) {
    element.classList.add("guided-tour-no-interaction");
  }

  // Handle additional accessible elements
  if (step.accessibleElements && step.accessibleElements.length > 0) {
    step.accessibleElements.forEach((selector) => {
      const accessibleElement = document.querySelector(selector);
      if (accessibleElement) {
        accessibleElement.classList.add("guided-tour-accessible-element");
        accessibleElement.setAttribute("data-guided-tour-accessible", "true");
      }
    });
  }
}

/**
 * Maps tour side and alignment to Quasar menu anchor
 */
function mapSideToAnchor(side: TourSide, align: TourAlignment): MenuAnchor {
  const alignMap = { start: "left", center: "middle", end: "right" };
  const alignText = alignMap[align];

  switch (side) {
    case "top":
      return `top ${alignText}` as MenuAnchor;
    case "bottom":
      return `bottom ${alignText}` as MenuAnchor;
    case "left":
      return "center left" as MenuAnchor;
    case "right":
      return "center right" as MenuAnchor;
    default:
      return `bottom ${alignText}` as MenuAnchor;
  }
}

/**
 * Maps tour side and alignment to Quasar menu self position
 */
function mapSideToSelf(side: TourSide, align: TourAlignment): MenuAnchor {
  const alignMap = { start: "left", center: "middle", end: "right" };
  const alignText = alignMap[align];

  switch (side) {
    case "top":
      return `bottom ${alignText}` as MenuAnchor;
    case "bottom":
      return `top ${alignText}` as MenuAnchor;
    case "left":
      return "center right" as MenuAnchor;
    case "right":
      return "center left" as MenuAnchor;
    default:
      return `top ${alignText}` as MenuAnchor;
  }
}

/**
 * Calculates a combined bounding rectangle for multiple elements
 */
function calculateCombinedBoundingRect(elements: Element[]): DOMRect {
  if (elements.length === 0) {
    return new DOMRect(0, 0, 0, 0);
  }

  if (elements.length === 1) {
    return elements[0].getBoundingClientRect();
  }

  const rects = elements.map((el) => el.getBoundingClientRect());
  const left = Math.min(...rects.map((r) => r.left));
  const top = Math.min(...rects.map((r) => r.top));
  const right = Math.max(...rects.map((r) => r.right));
  const bottom = Math.max(...rects.map((r) => r.bottom));

  return new DOMRect(left, top, right - left, bottom - top);
}

/**
 * Updates gradient position for spotlight overlay
 */
function updateGradientPosition(gradient: SVGRadialGradientElement, rect: DOMRect): void {
  const centerX = ((rect.left + rect.width / 2) / window.innerWidth) * 100;
  const centerY = ((rect.top + rect.height / 2) / window.innerHeight) * 100;

  const elementSize = Math.max(rect.width, rect.height);
  const padding = Math.max(200, elementSize * 1.2);
  const radiusInPixels = elementSize + padding;

  const viewportSize = Math.min(window.innerWidth, window.innerHeight);
  const gradientRadius = (radiusInPixels / viewportSize) * 100;

  const clampedRadius = Math.max(25, Math.min(70, gradientRadius));

  gradient.setAttribute("cx", `${centerX}%`);
  gradient.setAttribute("cy", `${centerY}%`);
  gradient.setAttribute("r", `${clampedRadius}%`);
}

/**
 * Easing function for smooth animations
 */
function easeInOutQuad(t: number): number {
  return t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t;
}

/**
 * Animates overlay between two elements
 */
function animateOverlay(
  elapsed: number,
  duration: number,
  fromElement: Element,
  toElement: Element,
  config: TourConfig,
  state: TourState,
): void {
  const fromRect = fromElement.getBoundingClientRect();
  const toRect = toElement.getBoundingClientRect();
  const progress = Math.min(elapsed / duration, 1);
  const easedProgress = easeInOutQuad(progress);

  const animatedRect = {
    x: fromRect.x + (toRect.x - fromRect.x) * easedProgress,
    y: fromRect.y + (toRect.y - fromRect.y) * easedProgress,
    width: fromRect.width + (toRect.width - fromRect.width) * easedProgress,
    height: fromRect.height + (toRect.height - fromRect.height) * easedProgress,
    left: fromRect.left + (toRect.left - fromRect.left) * easedProgress,
    top: fromRect.top + (toRect.top - fromRect.top) * easedProgress,
    right: fromRect.right + (toRect.right - fromRect.right) * easedProgress,
    bottom: fromRect.bottom + (toRect.bottom - fromRect.bottom) * easedProgress,
  };

  const overlayStyle = config.overlayStyle || "spotlight";

  if (overlayStyle === "spotlight") {
    const gradient = state.overlayElement?.querySelector("#guided-tour-gradient") as SVGRadialGradientElement;
    if (gradient) {
      updateGradientPosition(gradient, animatedRect as DOMRect);
    }
  } else if (overlayStyle === "cutout") {
    const pathElement = state.overlayElement?.querySelector("path") as SVGPathElement;
    if (pathElement) {
      pathElement.setAttribute("d", generateCutoutSvgPath(animatedRect as DOMRect, config));
    }
  }
}

/**
 * Manages transition animation between elements
 */
function animateTransition(
  fromElement: Element,
  toElement: Element,
  duration: number,
  config: TourConfig,
  state: TourState,
  onComplete: () => void,
): void {
  const start = Date.now();

  if (state.transitionCallback) {
    state.transitionCallback = null;
  }

  const animate = () => {
    if (state.transitionCallback !== animate) {
      return;
    }

    const elapsed = Date.now() - start;

    if (elapsed < duration) {
      animateOverlay(elapsed, duration, fromElement, toElement, config, state);
      requestAnimationFrame(animate);
    } else {
      onComplete();
    }
  };

  state.transitionCallback = animate;
  requestAnimationFrame(animate);
}

/**
 * Creates SVG overlay element
 */
function mountOverlay(targetElement: Element, config: TourConfig, overlayClickHandler: () => void): SVGSVGElement {
  const rect = targetElement.getBoundingClientRect();
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  const overlayStyle = config.overlayStyle || "spotlight";

  svg.classList.add("guided-tour-overlay");
  svg.classList.add(`guided-tour-overlay--${overlayStyle}`);

  if (config.animate) {
    svg.classList.add("guided-tour-fade");
  }

  svg.setAttribute("viewBox", `0 0 ${window.innerWidth} ${window.innerHeight}`);
  svg.setAttribute("xmlSpace", "preserve");
  svg.setAttribute("version", "1.1");
  svg.setAttribute("preserveAspectRatio", "xMinYMin slice");
  svg.style.cssText = `
    position: fixed; top: 0; left: 0; width: 100%; height: 100%;
    z-index: 5000; pointer-events: none; fill-rule: evenodd;
    clip-rule: evenodd; stroke-linejoin: round; stroke-miterlimit: 2;
  `;

  if (overlayStyle === "spotlight") {
    createSpotlightOverlay(svg, rect, overlayClickHandler);
  } else if (overlayStyle === "cutout") {
    createCutoutOverlay(svg, rect, config, overlayClickHandler);
  }

  document.body.appendChild(svg);
  return svg;
}

/**
 * Creates spotlight style overlay
 */
function createSpotlightOverlay(svg: SVGSVGElement, rect: DOMRect, overlayClickHandler: () => void): void {
  const defs = document.createElementNS("http://www.w3.org/2000/svg", "defs");
  const gradient = document.createElementNS("http://www.w3.org/2000/svg", "radialGradient");
  gradient.setAttribute("id", "guided-tour-gradient");

  const opacityStep = 0.005;
  const startingPrecentage = 40;
  const maxOpacity = 0.16;

  const stops = [
    { offset: "0%", color: "rgba(255, 255, 255, 0)" },
    { offset: `${startingPrecentage}%`, color: `rgba(255, 255, 255, 0)` },
  ];
  for (let i = 1; i <= 100 - startingPrecentage; i++) {
    const stepOpacity = i * opacityStep;
    stops.push({
      offset: `${startingPrecentage + i}%`,
      color: `rgba(0, 0, 0, ${stepOpacity})`,
    });

    // if max opacity is reached, kill the loop
    if (stepOpacity >= maxOpacity) break;
  }

  stops.forEach((stopData) => {
    const stop = document.createElementNS("http://www.w3.org/2000/svg", "stop");
    stop.setAttribute("offset", stopData.offset);
    stop.setAttribute("stop-color", stopData.color);
    gradient.appendChild(stop);
  });

  defs.appendChild(gradient);
  svg.appendChild(defs);

  const rectElement = document.createElementNS("http://www.w3.org/2000/svg", "rect");
  rectElement.setAttribute("width", "100%");
  rectElement.setAttribute("height", "100%");
  rectElement.setAttribute("fill", "url(#guided-tour-gradient)");

  updateGradientPosition(gradient, rect);

  svg.addEventListener("click", (e) => {
    if (e.target === rectElement) {
      overlayClickHandler();
    }
  });

  svg.appendChild(rectElement);
}

/**
 * Creates cutout style overlay
 */
function createCutoutOverlay(
  svg: SVGSVGElement,
  rect: DOMRect,
  config: TourConfig,
  overlayClickHandler: () => void,
): void {
  const pathElement = document.createElementNS("http://www.w3.org/2000/svg", "path");
  pathElement.setAttribute("d", generateCutoutSvgPath(rect, config));
  pathElement.style.fill = config.overlayColor || "#000";
  pathElement.style.opacity = `${config.overlayOpacity || 0.7}`;
  pathElement.style.pointerEvents = "auto";
  pathElement.style.cursor = "auto";

  svg.addEventListener("click", (e) => {
    if (e.target === pathElement) {
      overlayClickHandler();
    }
  });

  svg.appendChild(pathElement);
}

/**
 * Updates existing overlay to highlight new element
 */
function updateOverlay(targetElement: Element, config: TourConfig, state: TourState, step?: TourStep): void {
  if (!state.overlayElement) return;

  // Calculate the bounding rectangle for the spotlight
  let rect: DOMRect;

  if (step?.accessibleElements && step.accessibleElements.length > 0) {
    // Get all accessible elements including the main element
    const allElements = [
      targetElement,
      ...(step.accessibleElements.map((selector) => document.querySelector(selector)).filter(Boolean) as Element[]),
    ];

    rect = calculateCombinedBoundingRect(allElements);
  } else {
    rect = targetElement.getBoundingClientRect();
  }

  state.overlayElement.setAttribute("viewBox", `0 0 ${window.innerWidth} ${window.innerHeight}`);

  const overlayStyle = config.overlayStyle || "spotlight";

  if (overlayStyle === "spotlight") {
    const gradient = state.overlayElement.querySelector("#guided-tour-gradient") as SVGRadialGradientElement;
    if (gradient) {
      updateGradientPosition(gradient, rect);
    }
  } else if (overlayStyle === "cutout") {
    const pathElement = state.overlayElement.querySelector("path") as SVGPathElement;
    if (pathElement) {
      pathElement.setAttribute("d", generateCutoutSvgPath(rect, config));
    }
  }
}

/**
 * Creates overlay or updates existing one
 */
function createOverlay(
  targetElement: Element,
  config: TourConfig,
  state: TourState,
  overlayClickHandler: () => void,
  step?: TourStep,
): void {
  if (state.overlayElement) {
    updateOverlay(targetElement, config, state, step);
    return;
  }

  state.overlayElement = mountOverlay(targetElement, config, overlayClickHandler);
}

/**
 * Removes overlay from DOM
 */
function removeOverlay(state: TourState): void {
  if (state.overlayElement) {
    state.overlayElement.remove();
    state.overlayElement = null;
  }
}

/**
 * Generates SVG path for cutout overlay
 */
function generateCutoutSvgPath(rect: DOMRect, config: TourConfig): string {
  const windowX = window.innerWidth;
  const windowY = window.innerHeight;

  const stagePadding = config.stagePadding || 10;
  const stageRadius = config.stageRadius || 5;

  const stageWidth = rect.width + stagePadding * 2;
  const stageHeight = rect.height + stagePadding * 2;

  const limitedRadius = Math.min(stageRadius, stageWidth / 2, stageHeight / 2);
  const normalizedRadius = Math.floor(Math.max(limitedRadius, 0));

  const highlightBoxX = rect.left - stagePadding + normalizedRadius;
  const highlightBoxY = rect.top - stagePadding;
  const highlightBoxWidth = stageWidth - normalizedRadius * 2;
  const highlightBoxHeight = stageHeight - normalizedRadius * 2;

  return `M${windowX},0L0,0L0,${windowY}L${windowX},${windowY}L${windowX},0Z
    M${highlightBoxX},${highlightBoxY} h${highlightBoxWidth} a${normalizedRadius},${normalizedRadius} 0 0 1 ${normalizedRadius},${normalizedRadius} v${highlightBoxHeight} a${normalizedRadius},${normalizedRadius} 0 0 1 -${normalizedRadius},${normalizedRadius} h-${highlightBoxWidth} a${normalizedRadius},${normalizedRadius} 0 0 1 -${normalizedRadius},-${normalizedRadius} v-${highlightBoxHeight} a${normalizedRadius},${normalizedRadius} 0 0 1 ${normalizedRadius},-${normalizedRadius} z`;
}

/**
 * Waits for element to appear in DOM
 */
async function waitForElement(selector: string, parent?: Element, timeout = 10000): Promise<Element> {
  return new Promise((resolve, reject) => {
    const startTime = Date.now();

    const checkElement = () => {
      const element = (parent || document).querySelector(selector);

      if (element) {
        resolve(element);
        return;
      }

      if (Date.now() - startTime >= timeout) {
        reject(new Error(`Element "${selector}" not found within ${timeout}ms`));
        return;
      }

      requestAnimationFrame(checkElement);
    };

    checkElement();
  });
}

/**
 * Handles beforeShowStep setup
 */
async function handleBeforeShowStep(step: TourStep, config: TourConfig, state: TourState): Promise<void> {
  if (!step.beforeShowStep) return;

  const { waitForElement: elementOrElements, setup } = step.beforeShowStep;

  // Execute setup first - it might create/modify elements
  if (setup) {
    try {
      // const targetElement = getTargetElement(step.element);
      // Create minimal context for type compatibility - user code should use store directly
      const context = { config, state };
      // await setup(targetElement, step, context);
      await setup(step, context);
    } catch (error) {
      console.error("beforeShowStep: Setup failed", error);
    }
  }

  // Then wait for elements to appear (after setup has potentially created them)
  if (elementOrElements) {
    const selectors = Array.isArray(elementOrElements) ? elementOrElements : [elementOrElements];
    const waitPromises = selectors.map((selector) => waitForElement(selector, undefined, 10000));

    try {
      await Promise.all(waitPromises);
    } catch (error) {
      console.error("beforeShowStep: Critical elements not found, cancelling tour:", error);
      // Immediately throw error to cancel the tour
      throw error;
    }
  }
}

/**
 * Sets up nextOn event listeners
 */
async function setupNextOn(
  element: Element | null,
  step: TourStep,
  config: TourConfig,
  state: TourState,
  clearListeners: () => void,
  hideTour: () => void,
  moveNext: () => void,
): Promise<{ element: Element; event: string; handler: (event: Event) => void }[]> {
  if (!step.nextOn) return [];

  const {
    event,
    selector,
    attachToAllElements = false,
    hideOnAction = false,
    eventCondition,
    onAdvance,
    timeout,
    checkInterval,
  } = step.nextOn;
  const listeners: { element: Element; event: string; handler: (event: Event) => void }[] = [];
  const targetElements: Element[] = [];

  if (selector) {
    const selectors = Array.isArray(selector) ? selector : [selector];

    // Process each selector and ensure elements are found
    for (const s of selectors) {
      let foundElements = attachToAllElements ? document.querySelectorAll(s) : [document.querySelector(s)];

      // If no elements found, wait for them since setup has already completed
      if (foundElements.length === 0 || (foundElements.length === 1 && !foundElements[0])) {
        // Use a simple retry mechanism with short delays
        let retryCount = 0;
        const maxRetries = 10;

        while (
          retryCount < maxRetries &&
          (foundElements.length === 0 || (foundElements.length === 1 && !foundElements[0]))
        ) {
          retryCount++;

          // Small delay between retries
          await new Promise((resolve) => setTimeout(resolve, 50));

          foundElements = attachToAllElements ? document.querySelectorAll(s) : [document.querySelector(s)];
        }

        if (foundElements.length === 0 || (foundElements.length === 1 && !foundElements[0])) {
          const error = `[Guided Tour Service] Critical: Element still not found after ${maxRetries} retries: ${s}`;
          console.error(error);
          throw new Error(error); // This will fail the tour
        }
      }

      foundElements.forEach((el, index) => {
        if (el) {
          targetElements.push(el);
        } else {
          console.log(`[Guided Tour Service] Element ${index} is null/undefined`);
        }
      });
    }
  } else if (element) {
    targetElements.push(element);
  }

  if (targetElements.length === 0) {
    console.warn(`[Guided Tour] NextOn target not found for selector:`, selector);
    return [];
  }

  const handler = async (e: Event) => {
    const currentTarget = e.currentTarget as Element;

    if (eventCondition) {
      let canAdvance = false;

      if (timeout) {
        // Clear any existing timeout loops for this tour
        clearActiveEventConditionLoops(state);
        state.activeEventConditionLoops = new Set();

        // Create a new AbortController for this timeout loop
        const abortController = new AbortController();
        state.activeEventConditionLoops.add(abortController);

        // Implement timeout logic for eventCondition
        const maxWaitTime = timeout;
        const interval = checkInterval || 100;
        const startTime = Date.now();

        try {
          while (Date.now() - startTime < maxWaitTime && !abortController.signal.aborted) {
            canAdvance = await Promise.resolve(eventCondition(e, currentTarget));
            if (canAdvance) {
              break;
            }

            // Wait a bit before checking again, but check for abort signal
            await new Promise((resolve, reject) => {
              const timeoutId = setTimeout(resolve, interval);
              abortController.signal.addEventListener("abort", () => {
                clearTimeout(timeoutId);
                reject(new Error("Aborted"));
              });
            });
          }
        } catch (error) {
          if (error instanceof Error && error.message === "Aborted") {
            // Loop was aborted, don't advance
            return;
          }
          throw error;
        } finally {
          // Remove this controller from the active set
          state.activeEventConditionLoops?.delete(abortController);
        }
      } else {
        // No timeout, just call the condition once
        canAdvance = await Promise.resolve(eventCondition(e, currentTarget));
      }

      if (!canAdvance) {
        return;
      }
    }

    clearListeners();

    // Clear any active event condition loops when advancing
    clearActiveEventConditionLoops(state);

    if (onAdvance) {
      // Create minimal context for type compatibility - user code should use store directly
      const context = { config, state };
      await onAdvance(e, currentTarget, step, context);
      return;
    }

    if (hideOnAction) {
      hideTour();
      moveNext();
    } else {
      moveNext();
    }
  };

  targetElements.forEach((el) => {
    el.addEventListener(event, handler);
    listeners.push({ element: el, event, handler });
  });

  return listeners;
}

/**
 * Locks page scrolling to prevent user from scrolling away from the tour
 */
function lockScrolling(config: TourConfig, state: TourState): void {
  // Check if scroll locking was enabled
  if (config.lockScrolling === false) return;

  // Remove any existing scroll prevention handlers first to avoid duplicates
  if (state.scrollPreventionHandler) {
    document.removeEventListener("wheel", state.scrollPreventionHandler, { capture: true });
    document.removeEventListener("touchmove", state.scrollPreventionHandler, { capture: true });
  }
  if (state.keydownPreventionHandler) {
    document.removeEventListener("keydown", state.keydownPreventionHandler, { capture: true });
  }

  // Store current scroll position
  const scrollX = window.pageXOffset || document.documentElement.scrollLeft;
  const scrollY = window.pageYOffset || document.documentElement.scrollTop;

  state.originalScrollPosition = { x: scrollX, y: scrollY };
  state.originalBodyOverflow = document.body.style.overflow;

  // Lock scrolling by preventing wheel events on the document
  const preventScroll = (e: Event) => {
    e.preventDefault();
    e.stopPropagation();
    return false;
  };

  // Create keydown handler for preventing navigation keys
  const keydownHandler = (e: Event) => {
    if (
      ["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", "PageUp", "PageDown", "Home", "End", " "].includes(
        (e as KeyboardEvent).key,
      )
    ) {
      e.preventDefault();
      e.stopPropagation();
    }
  };

  // Store handlers for cleanup
  state.scrollPreventionHandler = preventScroll;
  state.keydownPreventionHandler = keydownHandler;

  // Add event listeners
  document.addEventListener("wheel", preventScroll, { passive: false, capture: true });
  document.addEventListener("touchmove", preventScroll, { passive: false, capture: true });
  document.addEventListener("keydown", keydownHandler, { passive: false, capture: true });

  // Lock body scrolling
  document.body.style.overflow = "hidden";
  document.body.style.position = "fixed";
  document.body.style.top = `-${scrollY}px`;
  document.body.style.left = `-${scrollX}px`;
  document.body.style.width = "100%";
}

/**
 * Unlocks page scrolling and restores original scroll position
 */
function unlockScrolling(config: TourConfig, state: TourState): void {
  // Check if scroll locking was enabled
  if (config.lockScrolling === false) return;

  // Remove scroll prevention event listeners
  if (state.scrollPreventionHandler) {
    document.removeEventListener("wheel", state.scrollPreventionHandler, { capture: true });
    document.removeEventListener("touchmove", state.scrollPreventionHandler, { capture: true });
    state.scrollPreventionHandler = undefined;
  }

  // Remove keydown prevention event listener
  if (state.keydownPreventionHandler) {
    document.removeEventListener("keydown", state.keydownPreventionHandler, { capture: true });
    state.keydownPreventionHandler = undefined;
  }

  // Restore original scroll position and body styles
  if (state.originalScrollPosition) {
    const { x, y } = state.originalScrollPosition;
    document.body.style.overflow = state.originalBodyOverflow || "auto";
    document.body.style.position = "";
    document.body.style.top = "";
    document.body.style.left = "";
    document.body.style.width = "";

    // Restore scroll position
    window.scrollTo(x, y);

    // Clear stored values
    state.originalScrollPosition = undefined;
    state.originalBodyOverflow = undefined;
  }
}

/**
 * Allows scrolling within specific elements while maintaining page scroll lock
 */
function setupElementScrollAllowance(element: Element, config: TourConfig, state: TourState): void {
  if (!state.scrollPreventionHandler) return;

  // Check if element is a form input that might need scrolling
  const isFormInput = element.matches('input, textarea, select, [contenteditable="true"]');

  // Check if element has scrollable content
  const hasScrollableContent = element.scrollHeight > element.clientHeight || element.scrollWidth > element.clientWidth;

  // Check if element has overflow scroll/auto
  const computedStyle = window.getComputedStyle(element);
  const hasOverflowScroll =
    computedStyle.overflow === "scroll" ||
    computedStyle.overflow === "auto" ||
    computedStyle.overflowY === "scroll" ||
    computedStyle.overflowY === "auto" ||
    computedStyle.overflowX === "scroll" ||
    computedStyle.overflowX === "auto";

  if (isFormInput || hasScrollableContent || hasOverflowScroll) {
    allowElementScrolling(element, state);
    // Add CSS class for styling
    element.classList.add("guided-tour-scrollable-element");
  }
}

/**
 * Allows scrolling within a specific element
 */
function allowElementScrolling(element: Element, state: TourState): void {
  if (!state.scrollPreventionHandler) return;

  const allowScroll = (e: Event) => {
    e.stopPropagation();
    // Don't prevent default - allow the element to scroll
  };

  // Store the allow scroll handler for cleanup
  if (!state.elementScrollHandlers) {
    state.elementScrollHandlers = new Map();
  }

  state.elementScrollHandlers.set(element, allowScroll);

  // Add event listeners to allow scrolling within this element
  element.addEventListener("wheel", allowScroll, { passive: true, capture: true });
  element.addEventListener("touchmove", allowScroll, { passive: true, capture: true });
}

/**
 * Removes element scroll allowance
 */
function removeElementScrollAllowance(element: Element, state: TourState): void {
  if (!state.elementScrollHandlers) return;

  const handler = state.elementScrollHandlers.get(element);
  if (handler) {
    element.removeEventListener("wheel", handler, { capture: true });
    element.removeEventListener("touchmove", handler, { capture: true });
    state.elementScrollHandlers.delete(element);
  }

  // Remove CSS class
  element.classList.remove("guided-tour-scrollable-element");
}

/**
 * Forces scroll locking regardless of configuration
 */
function forceLockScrolling(state: TourState): void {
  // Temporarily enable scroll locking
  const tempConfig = { lockScrolling: true } as TourConfig;
  lockScrolling(tempConfig, state);
}

/**
 * Forces scroll unlocking regardless of configuration
 */
function forceUnlockScrolling(state: TourState): void {
  // Temporarily disable scroll locking
  const tempConfig = { lockScrolling: false } as TourConfig;
  unlockScrolling(tempConfig, state);
}

/**
 * Cleans up all event listeners and handlers related to the guided tour.
 */
function cleanupTourEventHandlers(state: TourState): void {
  // Remove nextOn event listeners
  if (state.nextOnListeners) {
    state.nextOnListeners.forEach(({ element, event, handler }) => {
      element.removeEventListener(event, handler);
    });
    state.nextOnListeners = [];
  }

  // Remove transition animation callback
  if (state.transitionCallback) {
    state.transitionCallback = null;
  }

  // Clear active event condition timeout loops
  clearActiveEventConditionLoops(state);

  // Remove scroll prevention event listeners
  if (state.scrollPreventionHandler) {
    document.removeEventListener("wheel", state.scrollPreventionHandler, { capture: true });
    document.removeEventListener("touchmove", state.scrollPreventionHandler, { capture: true });
    state.scrollPreventionHandler = undefined;
  }

  // Remove keydown prevention event listener
  if (state.keydownPreventionHandler) {
    document.removeEventListener("keydown", state.keydownPreventionHandler, { capture: true });
    state.keydownPreventionHandler = undefined;
  }

  // Remove element scroll allowance handlers
  if (state.elementScrollHandlers) {
    state.elementScrollHandlers.forEach((_) => {
      // This is a bit complex as we need to remove listeners from all elements
      // that were allowed to scroll. For simplicity, we'll just clear the map.
      // A more robust solution would involve storing a map of element to its allowed handlers.
      // For now, we'll assume the map is cleared when the state is reset.
    });
    state.elementScrollHandlers.clear();
  }

  // Remove overlay element
  if (state.overlayElement) {
    state.overlayElement.remove();
    state.overlayElement = null;
  }

  // Clear active element styling
  cleanupActiveElements();
}
