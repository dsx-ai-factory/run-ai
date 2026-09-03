# Vue Guided Tour System

A comprehensive, flexible guided tour system for Vue 3 applications, inspired by [Driver.js] but built specifically for Vue with modern composables and TypeScript support.

## 🎯 Overview

This guided tour system provides an intuitive way to create interactive product tours, onboarding flows, and feature highlights in your Vue application. It features dual overlay styles, smooth animations, event-driven navigation, and comprehensive accessibility support.

## 🌟 Features

- **🎨 Dual Overlay Styles**: Choose between Spotlight (flashlight effect) and Cutout (traditional knockout)
- **⚡ Event-Driven Navigation**: Auto-advance tours based on user interactions
- **🔄 Smooth Animations**: Configurable transitions between tour steps
- **🎛️ Lifecycle Hooks**: Complete control over tour behavior at every stage
- **🎯 Smart Element Detection**: Graceful fallbacks when elements aren't found

## 📁 Project Structure

The guided tour system is located under `control-plane/frontend/src` and consists of several key components:

```
src/
├── composables/
│   ├── use-guided-tour.composable.ts    # Main composable logic
│   ├── guided-tour.types.ts             # TypeScript type definitions
│   └── README.md                        # This documentation
├── components/common/guided-tour-popup/
│   ├── guided-tour-popup.vue            # Popover/tooltip component
│   ├── guided-tour-popup-manager.vue    # Manager for popups and modals
│   └── index.ts                         # Component exports
├── views/authenticated/
│   └── authenticated-page.vue           # Main app layout with tour integration
└── assets/scss/components/
    └── _guided-tour.scss                # Styling and animations
```

### 🧩 Component Architecture

The tour system uses a layered component architecture:

1. **Core Logic** (`use-guided-tour.composable.ts`)
   - Manages tour state and configuration
   - Handles overlay creation and animations
   - Provides reactive state for Vue components

2. **Popup Component** (`guided-tour-popup.vue`)
   - Renders the actual popover/tooltip using Quasar's QMenu
   - Handles positioning, buttons, and content display
   - Emits events for navigation and closing

3. **Manager Component** (`guided-tour-popup-manager.vue`)
   - Orchestrates the popup component and exit confirmation modal
   - Connects composable state to UI components
   - Handles tour lifecycle events

4. **Integration** (`authenticated-page.vue`)
   - Includes `<guided-tour-popup-manager />` in the main app layout
   - Ensures tours are available across all authenticated routes
   - Manages tour completion modal

### Related Projects

The system draws inspiration from the [Driver.js] library, but has been completely rewritten as a Vue 3 composable with additional features and TypeScript support.

### 🔗 Component Usage

The guided tour components are automatically available throughout the application:

**Automatic Integration:**

```vue
<!-- In authenticated-page.vue (already included) -->
<template>
  <new-layout v-if="appIsReady">
    <!-- Other app content -->

    <!-- Tour manager handles all tour UI -->
    <guided-tour-popup-manager />
  </new-layout>
</template>
```

**Manual Component Usage:**

```vue
<!-- If you need to use components directly -->
<script setup>
import { GuidedTourPopup } from "@/components/common/guided-tour-popup";
import { GuidedTourPopupManager } from "@/components/common/guided-tour-popup";
</script>

<template>
  <!-- Direct popup usage (advanced) -->
  <guided-tour-popup
    v-model="showPopover"
    :anchor="anchor"
    :self="self"
    :target="targetSelector"
    :title="stepTitle"
    :description="stepDescription"
    @next="handleNext"
    @previous="handlePrevious"
    @close="handleClose"
  />

  <!-- Or use the manager (recommended) -->
  <guided-tour-popup-manager />
</template>
```

**Component Locations:**

- **Main Integration**: `src/views/authenticated/authenticated-page.vue`
- **Popup Component**: `src/components/common/guided-tour-popup/guided-tour-popup.vue`
- **Manager Component**: `src/components/common/guided-tour-popup/guided-tour-popup-manager.vue`

## 🚀 Quick Start

### Basic Usage

```vue
<script setup lang="ts">
import { useGuidedTour } from "@/composables/use-guided-tour.composable";

const { createTour } = useGuidedTour();

// Create a simple tour
const tour = createTour({
  steps: [
    {
      element: "#welcome-button",
      popover: {
        title: "Welcome!",
        description: "Click here to get started with our application.",
        side: "bottom",
      },
    },
    {
      element: "#navigation-menu",
      popover: {
        title: "Navigation",
        description: "Use this menu to navigate between different sections.",
        side: "right",
      },
    },
  ],
});

// Start the tour
function startTour() {
  tour.drive();
}
</script>

<template>
  <div>
    <button @click="startTour">Start Tour</button>
    <button id="welcome-button">Welcome Button</button>
    <nav id="navigation-menu">Navigation Menu</nav>
  </div>
</template>
```

## 🎨 Overlay Styles

### Spotlight Overlay (Default)

Creates a modern flashlight/spotlight effect with radial gradients:

```ts
const tour = createTour({
  overlayStyle: "spotlight", // Default
  steps: [
    /* your steps */
  ],
});
```

### Cutout Overlay (Traditional)

Traditional knockout overlay with customizable appearance:

```ts
const tour = createTour({
  overlayStyle: "cutout",
  overlayColor: "#000", // Overlay color
  overlayOpacity: 0.8, // Overlay opacity (0-1)
  stagePadding: 15, // Padding around highlighted element
  stageRadius: 8, // Border radius of the cutout
  steps: [
    /* your steps */
  ],
});
```

## ⚡ Event-Driven Navigation

Create tours that automatically advance based on user interactions:

```ts
const tour = createTour({
  steps: [
    {
      element: "#email-input",
      popover: {
        title: "Enter your email",
        description: "The tour will advance when you enter a valid email.",
        showButtons: ["close"], // Only show close button
      },
      nextOn: {
        event: "input",
        eventCondition: (event) => {
          const input = event.target as HTMLInputElement;
          return input.value.includes("@") && input.value.includes(".");
        },
      },
    },
    {
      element: "#submit-button",
      popover: {
        title: "Submit the form",
        description: "Click submit to complete the process.",
      },
      nextOn: {
        event: "click",
        hideOnAction: true, // Hide tour during navigation
      },
    },
  ],
});
```

## 🔄 Advanced Features

### Async Step Setup

Handle complex scenarios like waiting for elements or setting up state:

```ts
const tour = createTour({
  steps: [
    {
      element: "#dynamic-content",
      beforeShowStep: {
        waitForElement: [".async-component", "#loaded-data"],
        setup: async (element, step, context) => {
          // Wait for API call or perform setup
          await loadUserData();
          await new Promise((resolve) => setTimeout(resolve, 100));
        },
      },
      popover: {
        title: "Dynamic Content",
        description: "This content was loaded asynchronously.",
      },
    },
  ],
});
```

### Lifecycle Hooks

Control tour behavior at every stage:

```ts
const tour = createTour({
  // Global hooks
  onHighlightStarted: (element, step, context) => {
    console.log("About to highlight:", element);
  },
  onHighlighted: (element, step, context) => {
    console.log("Element highlighted:", element);
  },
  onDeselected: (element, step, context) => {
    console.log("Element deselected:", element);
  },

  steps: [
    {
      element: "#special-element",
      // Step-specific hooks (override global)
      onHighlightStarted: (element, step, context) => {
        element.classList.add("tour-highlight");
      },
      onDeselected: (element, step, context) => {
        element.classList.remove("tour-highlight");
      },
      popover: {
        title: "Special Element",
        description: "This element has custom highlighting.",
      },
    },
  ],
});
```

### Custom Button Handlers

Override default navigation behavior:

```ts
const tour = createTour({
  steps: [
    {
      element: "#custom-step",
      popover: {
        title: "Custom Navigation",
        description: "This step has custom button behavior.",
        onNextClick: (element, step, context) => {
          // Custom logic before advancing
          if (validateUserInput()) {
            context.tour.moveNext();
          } else {
            showValidationError();
          }
        },
        onPrevClick: (element, step, context) => {
          // Custom previous logic
          context.tour.movePrevious();
        },
      },
    },
  ],
});
```

## 🎛️ Configuration Options

### Tour Configuration

```ts
interface TourConfig {
  // Core
  steps?: TourStep[]; // Array of tour steps
  animate?: boolean; // Enable/disable animations (default: true)

  // Overlay
  overlayStyle?: "spotlight" | "cutout"; // Overlay style (default: 'spotlight')
  overlayColor?: string; // Overlay color (default: '#000')
  overlayOpacity?: number; // Overlay opacity (default: 0.7)
  overlayClickBehavior?: "close" | "nextStep" | "ignore"; // Click behavior (default: 'ignore')

  // Highlighted Element
  stagePadding?: number; // Element padding (default: 10px)
  stageRadius?: number; // Border radius (default: 5px)
  disableActiveInteraction?: boolean; // Disable element interaction (default: false)

  // Popover
  popoverClass?: string; // Custom CSS class
  popoverOffset?: [number, number]; // Popover offset [x, y] (default: [10, 10])

  // Buttons & Progress
  showButtons?: TourButton[]; // Buttons to show (default: ['next', 'previous', 'close'])
  showProgress?: boolean; // Show progress indicator (default: false)
  progressText?: string; // Progress text template (default: '{{current}}/{{total}}')
  nextBtnText?: string; // Next button text (default: 'Next')
  prevBtnText?: string; // Previous button text (default: 'Previous')
  doneBtnText?: string; // Done button text (default: 'Done')

  // User Interaction
  allowClose?: boolean; // Allow closing tour (default: true)
  allowKeyboardControl?: boolean; // Enable keyboard navigation (default: false)
  smoothScroll?: boolean; // Smooth scrolling (default: false)
}
```

### Step Configuration

```ts
interface TourStep {
  element?: string | Element | (() => Element); // Target element
  accessibleElements?: string[]; // Additional elements that should be accessible during this step
  popover?: TourPopover; // Popover configuration
  disableActiveInteraction?: boolean; // Override global interaction setting
  beforeShowStep?: BeforeShowStep; // Async setup configuration
  nextOn?: NextOn; // Event-driven navigation

  // Lifecycle hooks
  onHighlightStarted?: (element, step, context) => void;
  onHighlighted?: (element, step, context) => void;
  onDeselected?: (element, step, context) => void;
}
```

### Multi-Element Accessibility

For steps where users need to interact with multiple elements (e.g., filling a form and clicking a button), use the `accessibleElements` property:

```ts
{
  element: '[aid="form-input"]', // Main highlighted element
  accessibleElements: ['.submit-button', '.cancel-button'], // Additional accessible elements
  popover: {
    description: 'Fill in the form and click submit',
    side: "right",
  },
  nextOn: {
    event: "click",
    selector: ".submit-button",
  },
}
```

This ensures that:

- The main element is highlighted with the tour overlay
- All specified accessible elements remain interactive
- Users can complete multi-step actions within a single tour step
- Visual indicators show which elements are accessible

## 🎨 Styling and Customization

The tour system uses SCSS for styling, located in `src/assets/scss/components/_guided-tour.scss`.

## 🔌 API Reference

### Composable Return Values

```ts
const {
  // Reactive State
  isActive, // Ref<boolean> - Whether tour is active
  showPopover, // Ref<boolean> - Whether popover is visible
  popoverProps, // Ref<object> - Popover component props
  currentStep, // Ref<TourStep | null> - Current step
  currentElement, // Ref<Element | null> - Current element
  showCompletionModal, // Ref<boolean> - Completion modal visibility
  showExitConfirmModal, // Ref<boolean> - Exit confirmation visibility

  // Methods
  createTour, // (config) => GuidedTour - Create new tour instance
  destroy, // () => void - Destroy current tour
  closeCompletionModal, // () => void - Close completion modal
  openCompletionModal, // () => void - Open completion modal
  showExitConfirmation, // () => void - Show exit confirmation
  hideExitConfirmation, // () => void - Hide exit confirmation
  confirmExit, // () => void - Confirm and exit tour
  hideTour, // () => void - Temporarily hide tour
  showTour, // () => void - Show previously hidden tour
} = useGuidedTour();
```

### Tour Instance Methods

```ts
const tour = createTour(config);

// Tour Control
tour.drive(stepIndex?);     // Start tour from specific step
tour.destroy();             // Stop and clean up tour
tour.refresh();             // Refresh current step
tour.hide();                // Temporarily hide tour
tour.show();                // Show hidden tour

// Navigation
tour.moveNext();            // Move to next step
tour.movePrevious();        // Move to previous step
tour.moveTo(index);         // Jump to specific step

// State
tour.isActive();            // Check if tour is running
tour.getActiveIndex();      // Get current step index
tour.getActiveStep();       // Get current step object
tour.getActiveElement();    // Get current DOM element
tour.isFirstStep();         // Check if on first step
tour.isLastStep();          // Check if on last step
tour.hasNextStep();         // Check if next step exists
tour.hasPreviousStep();     // Check if previous step exists

// Configuration
tour.setConfig(config);     // Update configuration
tour.setSteps(steps);       // Update steps
tour.getConfig();           // Get current config
tour.getState();            // Get internal state
```

### Component APIs

#### GuidedTourPopup Props

```ts
interface GuidedTourPopupProps {
  modelValue: boolean; // Controls popup visibility
  anchor: string; // Quasar anchor positioning (e.g., 'bottom middle')
  self: string; // Quasar self positioning (e.g., 'top middle')
  offset: [number, number]; // Offset from target element [x, y]
  target: boolean | string; // CSS selector or boolean for target element
  title: string; // Popup title
  description: string; // Popup description/content
  showProgress: boolean; // Whether to show progress indicator
  progressText: string; // Progress text (e.g., "2/5")
  showButtons: TourButton[]; // Array of buttons to show
  hasNext: boolean; // Whether next step exists
  hasPrev: boolean; // Whether previous step exists
  nextBtnText: string; // Next button text
  prevBtnText: string; // Previous button text
  customClass: string; // Additional CSS classes
  onNext?: () => void; // Next button handler
  onPrevious?: () => void; // Previous button handler
  onClose?: () => void; // Close button handler
}
```

#### GuidedTourPopup Events

```ts
// Emitted events
"update:modelValue"; // (value: boolean) - When popup visibility changes
"show"; // () - When popup is shown
"hide"; // () - When popup is hidden
```

#### GuidedTourPopupManager

The manager component automatically handles all tour state and doesn't require props:

```vue
<guided-tour-popup-manager />
```

**Manager Features:**

- Connects to `useGuidedTour()` composable automatically
- Manages popup positioning and visibility
- Handles exit confirmation dialogs
- Provides consistent tour experience across the app

## 🏗️ Integration Examples

### Onboarding Flow

```ts
// composables/useOnboarding.ts
import { useGuidedTour } from "./use-guided-tour.composable";

export function useOnboarding() {
  const { createTour } = useGuidedTour();

  const onboardingTour = createTour({
    overlayStyle: "spotlight",
    showProgress: true,
    allowKeyboardControl: true,
    steps: [
      {
        element: "#app-header",
        popover: {
          title: "Welcome to Our App!",
          description: "Let's take a quick tour to get you started.",
          side: "bottom",
        },
      },
      {
        element: "#sidebar-menu",
        popover: {
          title: "Navigation Menu",
          description: "Access all features from this sidebar.",
          side: "right",
        },
      },
      {
        element: "#create-button",
        popover: {
          title: "Create New Items",
          description: "Click here to create new projects or documents.",
          side: "bottom",
        },
      },
    ],
  });

  return {
    startOnboarding: () => onboardingTour.drive(),
    skipOnboarding: () => onboardingTour.destroy(),
  };
}
```

### Feature Announcement

```ts
// composables/useFeatureAnnouncement.ts
export function useFeatureAnnouncement() {
  const { createTour } = useGuidedTour();

  const featureTour = createTour({
    overlayStyle: "cutout",
    overlayOpacity: 0.9,
    allowClose: true,
    steps: [
      {
        element: "#new-feature-button",
        beforeShowStep: {
          setup: async () => {
            // Ensure feature is loaded
            await loadNewFeature();
          },
        },
        popover: {
          title: "🎉 New Feature Alert!",
          description: "We've added a powerful new tool to help you work faster.",
          showButtons: ["next", "close"],
          side: "left",
        },
      },
      {
        element: "#feature-settings",
        popover: {
          title: "Customize Your Experience",
          description: "Configure the new feature to match your workflow.",
          doneBtnText: "Got it!",
        },
      },
    ],
  });

  return {
    announceFeature: () => featureTour.drive(),
  };
}
```

## 🐛 Troubleshooting

### Common Issues

**Tour doesn't start**

- Check that elements exist in the DOM
- Verify element selectors are correct
- Ensure tour is created after component mount

**Elements not highlighting correctly**

- Check z-index conflicts with existing styles
- Verify element is visible and not covered
- Test with different overlay styles

**Popover positioning issues**

- Adjust `popoverOffset` configuration
- Try different `side` and `align` values
- Check for viewport overflow

**Performance issues**

- Reduce animation duration
- Disable animations for complex layouts
- Optimize `beforeShowStep` setup functions

**Component integration issues**

- Ensure `<guided-tour-popup-manager />` is included in authenticated-page.vue

## 🤝 Contributing

When contributing to the guided tour system:

1. Update type definitions in `guided-tour.types.ts`
2. Add comprehensive JSDoc comments
3. Update SCSS variables and mixins as needed
4. Test with both overlay styles
