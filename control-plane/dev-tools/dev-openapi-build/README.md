# Go Services Build Tool

Automated build tool for Go services in the control-plane.

## Quick Start

```bash
cd control-plane/dev-tools/dev-openapi-build

# Run all steps for all services (default)
make

# Run all steps for a specific service
make workloads
make policy
make assets

# Run a specific step for all services
make openapi-build
make go-build
make go-mod
```

## Available Commands

### Service Targets
- `make workloads` - Build workloads service
- `make policy` - Build policy-service
- `make assets` - Build assets-service

### Step Targets (all services)
- `make openapi-build` - Generate OpenAPI code
- `make go-build` - Build Go binaries
- `make go-mod` - Run go mod tidy/vendor

### Advanced Usage

For more control, use the Python script directly:

```bash
# Run specific steps for specific services
python3 build_tool.py --services workloads --steps openapi-build

# See all options
python3 build_tool.py --help
```

## Services

- **workloads** - Workloads service
- **policy-service** - Policy service  
- **assets-service** - Assets service

## Build Steps

1. **mod-tidy-vendor** - Updates Go dependencies
2. **openapi-build** - Generates code from OpenAPI specs
3. **code-build** - Compiles Go binaries 