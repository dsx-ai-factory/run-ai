# Phoenix Playwright Test Suite

## Run All Tests

```bash
npx playwright test --config=playwright-config/phoenix-playwright.config.ts
```

## Run Specific Test File

```bash
# Apps tests
npx playwright test --config=playwright-config/phoenix-playwright.config.ts apps.test.ts

# Tenant tests
npx playwright test --config=playwright-config/phoenix-playwright.config.ts tenant.test.ts

# Tenant creation tests
npx playwright test --config=playwright-config/phoenix-playwright.config.ts tenantCreation.test.ts

# Tenant status codes tests
npx playwright test --config=playwright-config/phoenix-playwright.config.ts tenantStatusCodes.test.ts

# Delete tenant tests
npx playwright test --config=playwright-config/phoenix-playwright.config.ts deleteTenant.test.ts

# Database tests
npx playwright test --config=playwright-config/phoenix-playwright.config.ts testdb.test.ts
```

## Run Specific Test

```bash
# Run by test name (partial match)
npx playwright test --config=playwright-config/phoenix-playwright.config.ts -g "should get application default"

npx playwright test --config=playwright-config/phoenix-playwright.config.ts -g "should create tenant and validate name"

npx playwright test --config=playwright-config/phoenix-playwright.config.ts -g "should test db connection"
```

## Allure Reports

```bash
# Run tests (generates allure-results automatically)
npx playwright test --config=playwright-config/phoenix-playwright.config.ts

# Generate Allure report
npm run allure:generate

# Open Allure report in browser
npm run allure:serve

# Or generate and serve in one command
allure serve allure-results
```

## Additional Options

```bash
# Show test list without running
npx playwright test --config=playwright-config/phoenix-playwright.config.ts --list

# Run with debug output
npx playwright test --config=playwright-config/phoenix-playwright.config.ts --debug

# Run with headed browser (for UI tests)
npx playwright test --config=playwright-config/phoenix-playwright.config.ts --headed

# Generate HTML report
npx playwright show-report
