# Auto Update CLI Mechanism

The NVIDIA Run:ai CLI auto-update feature ensures that you are always utilizing the latest functionalities and critical fixes without the need for manual intervention. This mechanism is designed to streamline your workflow, prevent automation breakage due to outdated versions, and keep your CLI environment consistently up-to-date.

## Why Use Auto-Update

Manual updates can be time-consuming and often lead to inconsistencies across different user environments. The auto-update feature addresses these challenges by:

* **Eliminating the need for manual upgrades** - Users no longer have to periodically check for new versions or manually execute update commands.
* **Preventing automation breakage** - Automated scripts and workflows remain robust as the CLI client is always compatible with the latest backend changes and features.
* **Ensuring access to the latest features and fixes** - Users benefit immediately from new capabilities, performance improvements, and bug resolutions as soon as they are released.

## Configuring Auto-Update

The auto-update functionality is controlled via a new flag, `--auto-update` integrated into the [config set](runai/runai_config_set.md) CLI command. This allows for flexible configuration based on your operational needs.

### Enabling Auto-Update

To enable automatic version updates, execute the following command:

```bash
runai config set --auto-update true
```

Upon successful execution, you will receive a confirmation message:

```bash
Auto-upgrade has been enabled.
```

### Disabling Auto-Update

To disable automatic version updates, use the same command with the false argument:

```bash
runai config set --auto-update false
```

When auto-update is disabled, the CLI will confirm with a message and inform you about the change in behavior:

```bash
Auto-upgrade has been disabled. You will now see a warning on each command execution.
```

## Feature Behavior

The auto-update feature modifies the CLI's behavior based on its enabled or disabled state when a new version is detected.

### Default Behavior

* For existing CLI installations, the `--auto-update` flag is set to FALSE by default.
* For new installations  >= 2.22, the `--auto-update` flag is set to TRUE by default, ensuring new users immediately benefit from the auto-update mechanism.

### When Auto-Update is Disabled

If auto-update is disabled and a new version of the CLI is detected, a warning message will be displayed when logging in to the CLI. This warning serves as a persistent reminder to update your CLI manually to leverage the latest features and fixes.

```bash
Authentication was successful
Warning: You are currently using version x.y.z. A newer version a.b.c is available.
To update, run:
runai upgrade 
```

If you attempt to manually update the CLI using [runai upgrade](runai/runai_upgrade.md) while auto-update is disabled, the system will prompt you to enable the auto-update feature for a more seamless experience:

```bash
To enable automatic version updates, run:
runai config set --auto-update enabled
```

### When Auto-Update is Enabled

When auto-update is enabled, the CLI will automatically initiate an upgrade process before the execution of any log in command, provided a new version is detected. This ensures that your environment is always current before you begin your session.

```bash
A new version (vX.Y.Z) is available.
Updating now...
Update complete. (a.b.c --> x.y.x)
```

This automatic process ensures minimal disruption while keeping your CLI at the cutting edge.

## Error Scenarios

The auto-update mechanism includes robust error handling to inform you of any issues during the update process.

### Upgrade Failed

If the automatic upgrade process encounters an error and fails, the CLI will display a message advising you to attempt a manual upgrade:

```bash
Auto-upgrade failed. Please try again later or use 'runai upgrade' for manual upgrade. If it fails, contact your admin
```

It is recommended to first try [runai upgrade](runai/runai_upgrade.md) to resolve the issue. If the problem persists, contacting your administrator or Run:ai support is advised. Future documentation will include a dedicated upgrade guide to assist with troubleshooting.

### Unable to Check for Updates

In scenarios where the CLI cannot connect to the update server (e.g., due to network connectivity issues), a message indicating the inability to check for updates will be displayed:

```
Unable to check for updates.
```

This message indicates a potential network problem or an issue with the update service, preventing the CLI from verifying the latest version. You should check your internet connection or network configuration.
