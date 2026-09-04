# TUI Package - Advanced Screen-Based Architecture

This package implements a screen-based terminal user interface (TUI) for the Run:AI installer using the Bubble Tea framework. 

## Benefits and Design Principles

### Architectural Benefits

1. **Separation of Concerns**: Each screen manages its own logic, rendering, and state
2. **Maintainability**: Individual screens can be modified without affecting the overall system  
3. **Testability**: Comprehensive unit testing for each screen and component
4. **Extensibility**: New screens and features can be added by implementing standard interfaces
5. **Code Organization**: Related functionality is logically grouped and encapsulated
6. **Reusability**: Common components are shared across screens to reduce duplication

### Design Principles

- **User-Centric Design**: Clear navigation, helpful feedback, and intuitive workflows
- **Progressive Enhancement**: Advanced features don't interfere with basic functionality
- **Error Recovery**: Graceful error handling with clear guidance for resolution
- **Performance**: Efficient rendering and state management for responsive user experience
- **Accessibility**: Keyboard-only navigation with clear visual indicators

## Development and Testing

### Adding New Screens

1. **Implement the Screen Interface**: Create a new screen struct implementing all required methods
2. **Add to Flow**: Update screen navigation logic to include the new screen
3. **Create Tests**: Write comprehensive unit tests following existing patterns
4. **Update Documentation**: Add screen documentation and usage examples

### Testing Guidelines

- **Unit Tests**: Each screen has dedicated test files with comprehensive coverage
- **Integration Tests**: Screen transitions and data flow validation
- **UI Tests**: Rendering and interaction testing where applicable
- **Mock Dependencies**: Use mocks for external dependencies and services

### Common Patterns

- **State Management**: Use RunAIInstallData for sharing state between screens
- **Error Handling**: Consistent error message formatting and user guidance
- **Progress Indication**: Visual progress feedback for long-running operations
- **Validation**: Input validation with immediate feedback and clear error messages


## Architecture Overview

The TUI is built using a screen-based architecture where each screen implements a common interface, enabling seamless navigation and state management. Each screen is responsible for its own rendering, user input handling, and business logic, while maintaining consistency through shared components and utilities.

## Package Structure

### Core Architecture
```
pkg/tui/
├── runner.go                      # Main TUI program runner and initialization
├── inputs/                        # Input handling system
│   └── inputs.go                  # Text input component with horizontal scrolling
├── styles/                        # UI styling and theming
│   └── app_styles.go              # Application-wide styles (colors, borders, adaptive themes)
└── screens/                       # Screen implementations
    ├── assets.go                  # Screen registry and flow definitions
    ├── main_screen.go             # Main entry screen with frame management
    ├── prerequisites_screen.go    # Prerequisites selection
    ├── configuration_screen.go    # System configuration input
    ├── input_validation_screen.go # Progressive input validation
    ├── prerequisite_install_screen.go # Prerequisite installation
    ├── runai_installation_screen.go   # Run:AI component installation
    ├── common/                    # Shared screen components
    │   ├── types.go               # Screen and ScreenFrame interfaces
    │   ├── components.go          # Shared component utilities
    │   ├── installation_screen.go # Generic installation progress screen
    │   ├── installation_flow_manager.go # Installation flow rendering
    │   ├── options_screen.go      # Base options list model
    │   ├── options_multiple_selector.go # Multi-select component
    │   └── options_single_select.go     # Single-select component
    └── screen_component/          # Reusable UI components
        ├── types.go               # Component interfaces (ScreenFrame)
        ├── nav_buttons_frame.go   # Navigation buttons frame
        ├── nav_buttons_component.go # Button rendering
        ├── errors_frame.go        # Error messages display frame
        └── main_frame.go          # Main content frame
```

## Screen Interface and Implementation

All screens implement the `Screen` interface defined in `screens/common/types.go`:

```go
type Screen interface {
    Init(runaiData *common.RunAIInstallData) tea.Cmd  // Initialize screen with data
    Update(msg tea.Msg) tea.Cmd                       // Handle Bubble Tea messages
    Render() string                                   // Render screen content
    CanProceed() bool                                 // Check if user can proceed
    NextScreen() Screen                               // Get next screen in flow
    GetHeader() string                                // Get screen header text
    UpdateSelection(runaiData *common.RunAIInstallData) // Update data with selections
    GetMessages() []string                            // Get error/info messages to display
    GetItemCount() int                                // Number of selectable items
    OnFocus()                                         // Called when screen gains focus
    OnBlur()                                          // Called when screen loses focus
    FocusFirst()                                      // Focus first selectable item
    FocusLast()                                       // Focus last selectable item
}
```

### Frame Interface

Screens are composed of frames (main content, navigation buttons, errors) implementing `ScreenFrame`:

```go
type ScreenFrame interface {
    Update(msg tea.Msg) tea.Cmd
    View() string
    OnFocus()
    OnBlur()
    GetItemCount() int
}
```

## Screen Flow and Navigation

The installer follows a structured screen flow:

1. **MainScreen**: Application entry point with Kubernetes context display
2. **PrerequisitesScreen**: Select and configure prerequisites (mandatory/optional)
3. **ConfigurationScreen**: Input system configuration and credentials
4. **InputValidationScreen**: Progressive validation of all inputs
5. **PrerequisiteInstallScreen**: Automated prerequisite installation with progress
6. **RunAIInstallationScreen**: Run:AI Control Plane and Cluster installation

Each screen manages its own state and determines the next screen in the flow. Navigation between frames (main content, navigation buttons, errors) is handled via arrow keys.

## Key Components and Features

### Common Components (`screens/common/` and `screens/screen_component/`)

- **Navigation Buttons**: Standardized Next/Back/Exit buttons with focus management
- **Options Selectors**: Single and multiple selection components with keyboard navigation
- **Base Options Screen**: Common functionality for option-based screens
- **Shared Types**: Screen and ScreenFrame interfaces
- **Error Frame**: Dynamic error message display with colored output
- **Main Frame**: Content area with selection highlighting

### Input System (`inputs/`)

- **Text Input Component**: Custom text input with horizontal scrolling for long values
- **Cursor Management**: Visible cursor with scroll indicators for overflow text
- **Paste Support**: Clipboard paste with cursor positioning at end of pasted text

### Styling System (`styles/`)

- **Consistent Theming**: Application-wide styling using Lip Gloss
- **Responsive Design**: Adaptive layouts for different terminal sizes
- **Visual Hierarchy**: Clear information architecture with proper spacing and colors

### Session Management (`pkg/common/`)

- **Installation Data**: Centralized data model for installation configuration
- **Options Parsing**: Command-line argument processing and validation
- **Session Persistence**: Input values saved to Kubernetes secret for recovery
- **Kubernetes Context**: Cluster connection validation and context display

## Screen-Specific Features

### Configuration Screen
- **Dynamic Input Fields**: Context-aware input fields based on selection
- **Real-time Validation**: Immediate feedback on input validity
- **Field Dependencies**: Conditional field display based on previous selections

### Input Validation Screen
- **Progressive Validation**: Step-by-step validation with visual progress
- **Detailed Feedback**: Comprehensive error messages and resolution guidance
- **Retry Logic**: Automatic retry mechanisms for transient failures

### Prerequisite Install Screen
- **Parallel Installation**: Efficient prerequisite installation with progress tracking
- **Dependency Management**: Intelligent prerequisite dependency resolution
- **Rollback Support**: Ability to rollback failed installations

### Installation Screen
- **Real-time Progress**: Live installation progress with elapsed time and timeout display
- **Component Tracking**: Individual step status with spinner animation
- **Colored Status**: Green for success, red for failures
- **Step-level Timeout**: Configurable timeout per installation step
- **Cached Status**: Efficient rendering without excessive I/O during status polling