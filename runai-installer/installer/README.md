# Run:AI Installer

The Run:AI Installer is a comprehensive Terminal User Interface (TUI) built with Bubble Tea and Lip Gloss, designed to facilitate the installation of Run:AI platform on both cloud and on-premises environments. The installer provides a guided, screen-based workflow for deploying Run:AI components with real-time validation and progress tracking.

## Overview

This installer implements a modern, screen-based architecture where each installation step is handled by a dedicated screen component. The TUI guides users through the entire installation process, from initial configuration to component deployment, with comprehensive validation and error handling throughout.

## Features

- **Modern TUI Interface**: Beautiful, responsive terminal interface built with Bubble Tea and Lip Gloss
- **Screen-Based Architecture**: Modular design with dedicated screens for each installation phase
- **Comprehensive Installation Flow**: Supports CP, Cluster, Both, and Airgap SH installation types
- **Real-time Validation**: Progressive validation of user inputs with detailed feedback
- **Prerequisites Management**: Automated installation and validation of required components
- **Progress Tracking**: Visual progress indicators with elapsed time and timeout display
- **Error Handling**: Comprehensive error handling with clear user feedback and colored status indicators
- **Configuration Management**: Advanced configuration options for different deployment scenarios
- **Session Persistence**: Input values are automatically saved and restored between sessions
- **Kubernetes Context Awareness**: Displays current Kubernetes context in the title bar
- **Detailed Logging**: Comprehensive logging for all Kubernetes and Helm operations

## Installation Flow

The installer follows a comprehensive, screen-based flow:

1. **Installation Type Selection**: Choose between CP, Cluster, Both, or Airgap SH installation modes
2. **Deployment Environment**: Select Cloud or On-premise deployment configuration
3. **Prerequisites Selection**: Configure and install mandatory and optional prerequisites:
   - GPU Operator (NVIDIA)
   - Ingress Controller (NGINX)
   - Knative Serving
   - Network Operator (NVIDIA)
   - Prometheus
   - Training Operator (Kubeflow)
   - LWS - Distributed Inference (LeaderWorkerSet)
4. **Configuration Input**: Enter and validate system configurations:
   - NGC/JFROG credentials (with base64 validation)
   - FQDN (with DNS resolution validation)
   - Cluster name (Kubernetes naming conventions)
   - Custom ports (range validation: 30000-32767)
   - SSL certificates and TLS configurations
5. **Input Validation Phase**: Progressive real-time validation of all configurations
6. **Prerequisite Installation**: Automated installation of selected prerequisites with progress tracking
7. **Installation Progress**: Visual progress tracking with elapsed time and timeout display
8. **Completion**: Success confirmation (green) or failure details (red)

## Building

To build the installer:

```bash
make build
```

This will create binaries for multiple architectures in the `bin/` directory:
- `runai-installer-amd64` - For x86_64 systems
- `runai-installer-arm64` - For ARM64 systems

### Building for Specific Architectures

For ARM64 (Apple Silicon):
```bash
make build-darwin-arm
```

## Running

### Prerequisites

Before running the installer:
- Ensure `kubectl` is configured with a valid Kubernetes context
- The installer will validate the Kubernetes connection at startup

### Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--logs-dir` | Directory where installer logs should be written | Current directory |
| `--step-timeout` | Timeout for each installation step (e.g., '5m', '60s', '2m30s') | 7m |
| `--dev` | Enable development mode | false |
| `--gpu` | GPU type: `real` or `fake` | real |
| `--env` | Run:AI environment: `prod` or `test` | prod |
| `--cp-version` | Run:AI Control Plane version constraint | >0.0.0-0 |
| `--cluster-version` | Run:AI Cluster version constraint | >0.0.0-0 |

### Standard Usage

```bash
./bin/runai-installer-arm64 --logs-dir /path/to/logs
```

### With Custom Timeout

```bash
./bin/runai-installer-arm64 --logs-dir /path/to/logs --step-timeout 10m
```

### Development Mode

Development mode provides enhanced debugging and development capabilities:

```bash
./bin/runai-installer-arm64 --logs-dir /path/to/logs --dev
```

Development mode provides:
- Configuration info displayed in the title bar
- Enhanced logging and debugging information
- Ability to proceed past validation failures
- Additional debugging capabilities

You can also use the Makefile target for development:

```bash
make run
```

## Testing

Run the complete test suite:

```bash
make test
```

This includes:
- Unit tests for all components
- Screen-specific test suites
- Validation logic tests
- Integration tests

### Running Specific Tests

```bash
# Test TUI screens
go test -v ./pkg/tui/screens/...

# Test validation logic
go test -v ./pkg/inputs/validation/

# Test prerequisite steps
go test -v ./pkg/installations/prereq/steps/...

# Test with coverage
go test -cover -v ./pkg/...
```

## Project Structure

```
runai-installer/
├── main.go                           # Application entry point
├── pkg/
│   ├── common/                       # Common utilities and data structures
│   │   ├── install_data.go           # Installation data structures
│   │   ├── options.go                # Command-line options parsing
│   │   ├── session.go                # Session management (K8s secret persistence)
│   │   └── types.go                  # Common type definitions
│   ├── inputs/                       # Input definitions and validation
│   │   ├── config_input.go           # Input field type definitions
│   │   └── validation/               # Input validation logic
│   │       └── validator.go          # Validation functions
│   ├── helm/                         # Helm operations
│   │   ├── helm_actions.go           # Helm install/upgrade implementations
│   │   └── helm_clients.go           # Helm client management
│   ├── k8s/                          # Kubernetes client operations
│   │   └── k8s_clients.go            # K8s client initialization and utilities
│   ├── logging/                      # Logging utilities
│   │   └── logging.go                # Structured logging with file output
│   ├── installations/                # Installation logic
│   │   ├── common/                   # Common installation utilities
│   │   │   ├── base_step.go          # Base step implementation with timeout
│   │   │   ├── install_step.go       # InstallStep interface definition
│   │   │   ├── install_flow.go       # Installation flow management
│   │   │   ├── clients.go            # Run:AI API clients management
│   │   │   └── fqdn_actions.go       # FQDN-related operations
│   │   ├── pre_req/                  # Prerequisites management
│   │   │   ├── pre_req_install_step.go # Prerequisite step base
│   │   │   └── steps/                # Prerequisite implementations
│   │   │       ├── types.go          # PrerequisiteStep interface
│   │   │       ├── pre_requisites_on_cluster.go # Prerequisites registry
│   │   │       ├── gpu_operator.go   # NVIDIA GPU Operator
│   │   │       ├── ingress.go        # NGINX Ingress Controller
│   │   │       ├── knative.go        # Knative Serving
│   │   │       ├── network_operator.go # NVIDIA Network Operator
│   │   │       ├── prometheus.go     # Prometheus Stack
│   │   │       ├── training_operator.go # Kubeflow Training Operator
│   │   │       ├── lws.go            # LeaderWorkerSet (LWS)
│   │   │       └── crd_checker.go    # CRD existence checker
│   │   └── runai_install/            # Run:AI installation logic
│   │       ├── component.go          # Component interface
│   │       ├── runai_install_step.go # Run:AI step base
│   │       ├── components/           # Run:AI component definitions
│   │       │   ├── cp.go             # Control Plane component
│   │       │   └── cluster.go        # Cluster component
│   │       ├── steps/                # Run:AI installation steps
│   │       │   ├── cp_chart.go       # Control Plane chart installation
│   │       │   ├── cp_install_repo_secret.go # Registry credentials
│   │       │   ├── install_cluster.go # Cluster creation and installation
│   │       │   ├── namespace_stage.go # Namespace creation
│   │       │   ├── tls_secret_stage.go # TLS secret creation
│   │       │   └── ca_secret_stage.go # CA secret creation
│   │       └── cert/                 # Certificate management
│   │           ├── cert_factory.go   # Certificate factory
│   │           ├── cert_handler.go   # Certificate operations
│   │           └── self_cert.go      # Self-signed certificate generation
│   └── tui/                          # Terminal User Interface
│       ├── runner.go                 # TUI program runner and initialization
│       ├── styles/                   # UI styling definitions
│       │   └── app_styles.go         # Colors, borders, adaptive themes
│       ├── inputs/                   # TUI input components
│       │   └── inputs.go             # Text input with horizontal scrolling
│       └── screens/                  # Screen implementations
│           ├── assets.go             # Screen registry and flow definitions
│           ├── main_screen.go        # Main entry screen with frame management
│           ├── prerequisites_screen.go # Prerequisites selection
│           ├── configuration_screen.go # Configuration input
│           ├── input_validation_screen.go # Input validation with progress
│           ├── prerequisite_install_screen.go # Prerequisite installation
│           ├── runai_installation_screen.go # Run:AI installation
│           ├── common/               # Shared screen components
│           │   ├── types.go          # Screen and ScreenFrame interfaces
│           │   ├── installation_screen.go # Generic installation screen
│           │   ├── installation_flow_manager.go # Flow rendering
│           │   ├── options_screen.go # Options list model
│           │   ├── options_multiple_selector.go # Multi-select
│           │   └── options_single_select.go # Single-select
│           └── screen_component/     # Reusable UI components
│               ├── types.go          # Component interfaces
│               ├── nav_buttons_frame.go # Navigation buttons frame
│               ├── nav_buttons_component.go # Button rendering
│               ├── errors_frame.go   # Error messages frame
│               └── main_frame.go     # Main content frame
├── bin/                              # Build output directory
├── Makefile                          # Build and development commands
├── Dockerfile                        # Container build configuration
└── README.md                         # This file
```

## Navigation and User Interface

The TUI supports comprehensive keyboard navigation:

### General Navigation
- **Arrow Keys (↑/↓)**: Navigate between options, menu items, and frames
- **Enter**: Select options, confirm actions, toggle checkboxes, or proceed
- **Space**: Toggle checkbox selections
- **Delete/Backspace**: Delete characters in input fields
- **Ctrl+C**: Quit the application at any time

### Frame Navigation
- **↓ at end of list**: Move focus to navigation buttons
- **↑ at start of list**: Move focus to navigation buttons  
- **↑ in nav buttons**: Move focus to last item in main frame (if selectable)
- **↓ in nav buttons**: Move focus to first item in main frame (if selectable)

### Screen-Specific Controls
- **Configuration Screen**: Type directly into input fields, use arrows to navigate between fields
- **Selection Screens**: Use arrow keys to highlight options, Enter/Space to select
- **Validation Screen**: Automatic progression with real-time status updates
- **Installation Screen**: View-only progress tracking with elapsed time display

### Button Actions
- **Next**: Proceed to the next screen (available when requirements are met)
- **Back**: Return to the previous screen
- **Exit**: Quit the installer immediately

## Timeout and Progress Tracking

Each installation step has a configurable timeout (default: 7 minutes). The UI displays:
- **Elapsed time**: How long the current step has been running
- **Timeout**: Maximum allowed time for the step
- **Status indicators**: Installing spinner, success (green ✓), or failure (red ✗)

Configure the timeout using the `--step-timeout` flag:
```bash
./runai-installer --logs-dir ./logs --step-timeout 10m
```

## Session Persistence

The installer automatically saves input values to a Kubernetes secret (`runai-installer-session` in the `default` namespace). This allows:
- Resuming installation after interruption
- Reusing configuration across multiple installation attempts
- Preserving sensitive data securely in the cluster

## Logging

All operations are logged to the specified logs directory:
- **Application logs**: Startup, shutdown, and general application events
- **Kubernetes operations**: All API calls to the cluster
- **Helm operations**: Chart installations and upgrades
- **kubectl commands**: External command executions with full output

Log file location: `<logs-dir>/runai-installer.log`

### Log Format
```
time="2025-12-14 10:30:00.123" level=info msg="[K8s] Creating namespace 'runai-backend'..."
time="2025-12-14 10:30:01.456" level=info msg="[Helm] Starting install for release 'control-plane'..."
```

## Prerequisites

### Mandatory Prerequisites
- **Ingress Controller**: NGINX ingress controller for traffic routing

### Optional Prerequisites
- **GPU Operator**: NVIDIA GPU Operator for GPU workload support
- **Knative Serving**: Serverless workload support
- **Network Operator**: NVIDIA Network Operator for advanced networking
- **Prometheus**: Monitoring and metrics collection
- **Training Operator**: Kubeflow Training Operator for ML training jobs
- **LWS**: LeaderWorkerSet for distributed inference workloads

Each prerequisite includes:
- Automatic detection of existing installations
- Version compatibility checking
- Automated installation with progress tracking
- Health checks post-installation

## Troubleshooting

### Common Issues

1. **Kubernetes Connection Failed**: Ensure `kubectl` is configured and the cluster is accessible
2. **Timeout Errors**: Increase `--step-timeout` for slow networks or large deployments
3. **Permission Errors**: Verify cluster-admin or equivalent permissions
4. **Network Issues**: Ensure access to container registries (NGC, JFROG)

### Checking Logs

```bash
# View recent logs
tail -f /path/to/logs/runai-installer.log

# Search for errors
grep -i error /path/to/logs/runai-installer.log

# Search for specific component
grep "\[Helm\]" /path/to/logs/runai-installer.log
```

### Development Mode Benefits

Development mode (`--dev` flag) provides:
- Configuration info in title bar (versions, timeout, GPU type)
- Enhanced logging output
- Ability to bypass certain validations for testing
- Additional debug information

## Contributing

When contributing to this project, please ensure:

1. **Code Quality**: All code follows Go best practices and project conventions
2. **Testing**: All tests pass - run `make test` before submitting
3. **Formatting**: Code is properly formatted - run `make purify`
4. **Interface Compliance**: New steps implement the `InstallStep` interface
5. **Logging**: Add appropriate logging for all Kubernetes and Helm operations
6. **Documentation**: Update relevant documentation for new features

### Adding New Prerequisites

1. Create a new file in `pkg/installations/prereq/steps/`
2. Implement the `PrerequisiteStep` interface
3. Add `GetStatus()` for installation detection
4. Add `Execute()` for installation logic
5. Register in `pkg/installations/prereq/pre_requisites_on_cluster.go`
6. Add tests following existing patterns

## License

This project is licensed under the Apache License, Version 2.0.
