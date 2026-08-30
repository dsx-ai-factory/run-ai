import type { Ref } from "vue";
import type { QMenu } from "quasar";

export type MenuAnchor = QMenu["anchor"];

export type TourSide = "top" | "right" | "bottom" | "left";
export type TourAlignment = "start" | "center" | "end";
export type TourButton = "next" | "previous" | "close";

export interface TourPopover {
  title?: string;
  description?: string;
  side?: TourSide;
  align?: TourAlignment;
  showButtons?: TourButton[];
  showProgress?: boolean;
  nextBtnText?: string;
  prevBtnText?: string;
  doneBtnText?: string;
  progressText?: string;
  offset?: [number, number];
  onNextClick?: (element: Element | null, step: TourStep, context: TourContext) => void;
  onPrevClick?: (element: Element | null, step: TourStep, context: TourContext) => void;
  onCloseClick?: (element: Element | null, step: TourStep, context: TourContext) => void;
}

export interface BeforeShowStep {
  waitForElement?: string | string[];
  setup?: (step: TourStep, context: TourContext) => void | Promise<void>;
}

export interface NextOn {
  /** The DOM event that should trigger advancing to the next step. */
  event: "click" | "keydown" | "change" | "input" | "submit" | "focus" | "blur";
  /** An optional CSS selector for the element to attach the event listener to. If not provided, the step's main element is used. */
  selector?: string | string[];
  /** If the selector finds multiple elements, this flag determines whether to attach the listener to all of them. Defaults to false. */
  attachToAllElements?: boolean;
  /**
   * Whether to hide the tour when the action is performed, before moving to the next step.
   * This is ideal for steps that trigger navigation. The tour will remain hidden until the next step's `beforeShowStep` conditions are met.
   */
  hideOnAction?: boolean;
  /** An optional function that must return true for the advance to occur. Can be async. */
  eventCondition?: (event: Event, element: Element) => boolean | Promise<boolean>;
  /** An optional callback executed just before the tour advances. Takes precedence over `hideOnAction`. */
  onAdvance?: (event: Event, element: Element, step: TourStep, context: TourContext) => void | Promise<void>;
  /** Optional timeout in milliseconds for the eventCondition to wait before failing. Defaults to 10000ms (10 seconds). */
  timeout?: number;
  /** Optional interval in milliseconds between checks for the eventCondition. Defaults to 100ms. */
  checkInterval?: number;
}

export interface TourStep {
  element?: string | Element | (() => Element);
  /** Additional elements that should be accessible during this step */
  accessibleElements?: string[];
  popover?: TourPopover;
  disableActiveInteraction?: boolean;
  /** Hide the overlay for this step (no dark background, no highlight) */
  hideOverlay?: boolean;
  beforeShowStep?: BeforeShowStep;
  /** Configuration for automatically advancing to the next step when a specific event occurs. */
  nextOn?: NextOn;
  onHighlightStarted?: (element: Element | null, step: TourStep, context: TourContext) => void;
  onHighlighted?: (element: Element | null, step: TourStep, context: TourContext) => void;
  onDeselected?: (element: Element | null, step: TourStep, context: TourContext) => void;
}

export interface TourConfig {
  /** The array of steps that make up the tour. */
  steps?: TourStep[];
  /** Whether to animate the transition between steps. */
  animate?: boolean;
  /** The style of overlay to use: 'spotlight' (flashlight effect) or 'cutout' (traditional knockout). */
  overlayStyle?: "spotlight" | "cutout";
  /** The color of the overlay. */
  overlayColor?: string;
  /** The opacity of the overlay (0 to 1). */
  overlayOpacity?: number;
  /** Whether to use smooth scrolling when bringing elements into view. */
  smoothScroll?: boolean;
  /** The scroll behavior when bringing elements into view: 'center' (default), 'start', 'end', 'nearest', or 'smart'. */
  scrollBehavior?: "center" | "start" | "end" | "nearest" | "smart";
  /** Whether the user can close the tour by pressing Escape or clicking the close button. */
  allowClose?: boolean;
  /** Defines the behavior when the overlay is clicked. */
  overlayClickBehavior?: "close" | "nextStep" | "ignore";
  /** Padding around the highlighted element. */
  stagePadding?: number;
  /** The border radius of the highlighted element's cutout. */
  stageRadius?: number;
  /** Whether to disable interaction with the highlighted element. */
  disableActiveInteraction?: boolean;
  /** Whether to allow keyboard navigation (ArrowKeys, Escape). */
  allowKeyboardControl?: boolean;
  /** A CSS class to apply to the popover element. */
  popoverClass?: string;
  /** The offset of the popover from the highlighted element. */
  popoverOffset?: [number, number];
  /** The buttons to show in the popover. */
  showButtons?: TourButton[];
  /** Whether to show the progress indicator (e.g., "1/5"). */
  showProgress?: boolean;
  /** The template for the progress text. Use `{{current}}` and `{{total}}`. */
  progressText?: string;
  /** The text for the "Next" button. */
  nextBtnText?: string;
  /** The text for the "Previous" button. */
  prevBtnText?: string;
  /** The text for the "Done" button on the last step. */
  doneBtnText?: string;
  /** Whether to lock page scrolling during the tour to prevent users from scrolling away from highlighted elements. Defaults to true. */
  lockScrolling?: boolean;
  /** A callback executed just before a step is highlighted. */
  onHighlightStarted?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** A callback executed after a step has been highlighted. */
  onHighlighted?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** A callback executed after a step has been deselected. */
  onDeselected?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** A callback executed when the tour is destroyed. */
  onDestroyed?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** A callback for when the "Next" button is clicked. Overrides the default behavior. */
  onNextClick?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** A callback for when the "Previous" button is clicked. Overrides the default behavior. */
  onPrevClick?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** A callback for when the "Close" button is clicked. Overrides the default behavior. */
  onCloseClick?: (element: Element | null, step: TourStep, context: TourContext) => void;
  /** Whether to enable debug logging for scroll behavior and centering. */
  debug?: boolean;
  /** Custom viewport adjustments for centering calculations. */
  viewportAdjustments?: {
    headerHeight?: number;
    footerHeight?: number;
    popupWidth?: number;
    popupHeight?: number;
  };
}

export interface TourContext {
  config: TourConfig;
  state: TourState;
  tour?: GuidedTour; // Optional - use store directly instead
}

export interface TourState {
  isInitialized: boolean;
  activeIndex: number | undefined;
  activeElement: Element | null;
  activeStep: TourStep | null;
  previousElement: Element | null;
  previousStep: TourStep | null;
  overlayElement: SVGSVGElement | null;
  isTransitioning: boolean;
  animationDuration: number;
  transitionCallback?: (() => void) | null;
  // Scroll locking properties
  originalScrollPosition?: { x: number; y: number };
  originalBodyOverflow?: string;
  scrollPreventionHandler?: (e: Event) => boolean;
  keydownPreventionHandler?: (e: Event) => void;
  elementScrollHandlers?: Map<Element, (e: Event) => void>;
  // Tour-specific event handlers
  tourEventHandlers?: Array<{ event: string; handler: (e: Event) => void; capture: boolean }>;
  // NextOn event listeners
  nextOnListeners?: Array<{ element: Element; event: string; handler: (e: Event) => void }>;
  // Active timeout loops for eventCondition functions
  activeEventConditionLoops?: Set<AbortController>;
  // Dynamic element monitoring
  trackedElements?: Set<Element>;
  elementMonitor?: MutationObserver | null;
}

export interface GuidedTour {
  isActive: () => boolean;
  drive: (stepIndex?: number) => void;
  setConfig: (config: TourConfig) => void;
  setSteps: (steps: TourStep[]) => void;
  getConfig: () => TourConfig;
  getState: () => TourState;
  getActiveIndex: () => number | undefined;
  isFirstStep: () => boolean;
  isLastStep: () => boolean;
  getActiveStep: () => TourStep | null;
  getActiveElement: () => Element | null;
  moveNext: () => void;
  movePrevious: () => void;
  moveTo: (index: number) => void;
  hasNextStep: () => boolean;
  hasPreviousStep: () => boolean;
  highlight: (step: TourStep) => void;
  destroy: () => void;
  refresh: () => void;
  hide: () => void;
  show: () => void;
}

export interface PopoverProps {
  modelValue: boolean;
  anchor: MenuAnchor;
  self: MenuAnchor;
  offset: [number, number];
  noParentEvent: boolean;
  target: string;
  step: TourStep;
  title?: string;
  description?: string;
  showProgress: boolean;
  progressText: string;
  showButtons: TourButton[];
  hasNext: boolean;
  hasPrev: boolean;
  nextBtnText?: string;
  prevBtnText?: string;
  customClass: string;
}

export interface UseGuidedTourReturn {
  readonly isActive: Ref<boolean>;
  readonly showPopover: Ref<boolean>;
  readonly popoverProps: Ref<Partial<PopoverProps>>;
  readonly currentStep: Ref<TourStep | null>;
  readonly currentElement: Ref<Element | null>;
  readonly showCompletionModal: Ref<boolean>;
  readonly showExitConfirmModal: Ref<boolean>;
  createTour: (options?: TourConfig) => GuidedTour;
  destroy: () => void;
  closeCompletionModal: () => void;
  openCompletionModal: () => void;
  showExitConfirmation: () => void;
  hideExitConfirmation: () => void;
  confirmExit: () => void;
  hideTour: () => void;
  showTour: () => void;
  onNext: () => void;
  onPrevious: () => void;
  onClose: () => void;
}
