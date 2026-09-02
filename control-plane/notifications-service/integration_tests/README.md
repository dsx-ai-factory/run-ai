# notifications-service Integration-Tests

Integration tests are aimed to run e2e flows

We employ Docker compose in order to launch 4 dockers:
1. SMTP service
2. Postgresql service (in-docker volatile data)
3. Notification service
4. Integration tests binaries (pre-compiled)

# Debugging
There are 2 modes to launch the integration tests:
1. "make ocker-compose-integration-tests-up" (CI mode)
2. "make ocker-compose-integration-tests-up-debug" (IDE Debug mode)

## CI mode
The first 1 will build and launch the dockers above. 

## IDE Debug mode
The second one will build and launch only the 3rd party Dockers.
Then, you can launch the Notification service from your IDE, and run the integration tests manually.

For "Notification service" - run the main with following environment variables:
    CONF_PATH=./integration_tests/config/config.advanced.local.yaml;CONF_DB_HOST=local;ENV_CONFIG_PROVIDER=advanced;ENCRYPT_DECRYPT_KEY=0123456789abcdefghwertol
or
    CONF_PATH=./integration_tests/config/config.file.yaml.local.yaml;CONF_DB_HOST=local;ENV_CONFIG_PROVIDER=file;ENCRYPT_DECRYPT_KEY=0123456789abcdefghwertol

For running the tests from IDE - should set following env vars
    NOTIFICATION_HOST=localhost;SMTP_HOST=localhost;

## Checking email inbox
Use http://localhost:8025/ to view emails
