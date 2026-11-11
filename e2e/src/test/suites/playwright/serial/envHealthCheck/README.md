# Environment Health Check Test Suite

## Structure

The environment health check suite has its own configuration and hooks:

- **Configuration**: `playwright-config/envHealthCheck-playwright.config.ts` - Dedicated Playwright configuration
- **Global Setup**: `envHealthCheck-playwright-setup.ts` - Basic hook setup (Kubernetes context only)
- **Global Teardown**: `envHealthCheck-playwright-teardown.ts` - Basic cleanup (force browser termination)
- **Tests**: `src/test/suites/playwright/envHealthCheck/` - Health check test files

## Hook Type: Basic Hook

This suite uses the **basic hook** approach (similar to the original Mocha version), which provides:
- ✅ Kubernetes context initialization
- ✅ Admin context switching  
- ✅ Browser cleanup on teardown
- ❌ No complex authentication setup
- ❌ No user creation or token management



## Test Categories

### Nodes Check
- Validates node health status
- Verifies GPU availability in the cluster

### Control Plane & Cluster
- Checks RunAI namespace pod health
- Validates RunAI-backend pods (self-hosted only)
- Verifies GPU operator pods
- Checks monitoring namespace pods (non-OCP platforms)
- Validates operator status (version 2.9+)

### RunAI CLI
- Downloads and validates RunAI CLI functionality

### UI Health Check
- Tests UI login functionality with browser automation

### API Health Check  
- Validates cluster connectivity
- Tests cluster API endpoints

## Running the Tests

To run the environment health check suite:

```bash
npx playwright test --config=playwright-config/envHealthCheck-playwright.config.ts
```

## Environment Variables

The suite respects standard RunAI e2e environment variables:
- `CLUSTER_NAME_PREFIX` - Determines if running on self-hosted vs SaaS
- `RETRIES` - Number of retry attempts for failed tests
- `CONTROL_PLANE_URL` - Control plane 
