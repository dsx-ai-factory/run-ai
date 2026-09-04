.PHONY: all openapi-build openapi-build-only go-build go-mod workloads policy assets workloads-manager help

# Default target - Run all steps (including mod-tidy-vendor) for all services
all:
	python3 -u build_tool.py --steps mod-tidy-vendor openapi-build code-build

# Specific steps for all services
openapi-build:
	python3 -u build_tool.py --steps mod-tidy-vendor openapi-build

openapi-build-only:
	python3 -u build_tool.py --steps openapi-build

go-build:
	python3 -u build_tool.py --steps mod-tidy-vendor code-build

go-mod:
	python3 -u build_tool.py --steps mod-tidy-vendor

# Target specific services (runs all steps: mod-tidy-vendor, openapi-build, code-build)
workloads:
	python3 -u build_tool.py --services workloads --steps mod-tidy-vendor openapi-build code-build

policy:
	python3 -u build_tool.py --services policy-service --steps mod-tidy-vendor openapi-build code-build

assets:
	python3 -u build_tool.py --services assets-service --steps mod-tidy-vendor openapi-build code-build

workloads-manager:
	python3 -u build_tool.py --services workloads-manager --steps mod-tidy-vendor openapi-build code-build

help:
	@echo "Makefile for managing build steps via build_tool.py:"
	@echo ""
	@echo "  General Targets:"
	@echo "    make all                 - Run ALL defined steps (mod-tidy-vendor, openapi-build, code-build) for ALL services."
	@echo ""
	@echo "  Step-Specific Targets (for ALL services):"
	@echo "    make openapi-build       - Run 'openapi-build' step for ALL services."
	@echo "    make openapi-build-only  - Run 'openapi-build' step only (no mod-tidy-vendor) for ALL services."
	@echo "    make go-build            - Run 'code-build' step for ALL services."
	@echo "    make go-mod              - Run 'mod-tidy-vendor' step for ALL services."
	@echo ""
	@echo "  Service-Specific Targets (runs ALL steps for the service):"
	@echo "    make workloads           - Run ALL steps (mod-tidy-vendor, openapi-build, code-build) for 'workloads' service."
	@echo "    make policy              - Run ALL steps (mod-tidy-vendor, openapi-build, code-build) for 'policy-service' service."
	@echo "    make assets              - Run ALL steps (mod-tidy-vendor, openapi-build, code-build) for 'assets-service' service."
	@echo "    make workloads-manager   - Run ALL steps (mod-tidy-vendor, openapi-build, code-build) for 'workloads-manager' service."
	@echo ""
	@echo ""
	@echo "For more granular control, use the Python script directly:"
	@echo "  python3 -u build_tool.py --help" 