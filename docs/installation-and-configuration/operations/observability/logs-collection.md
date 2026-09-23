# Logs Collection

This guide provides instructions for IT administrators on collecting NVIDIA Run:ai logs for support, including prerequisites, CLI commands, and log file retrieval. It also covers enabling verbose logging for Prometheus and the NVIDIA Run:ai Scheduler.

## Collect Logs to Send to Support

To collect NVIDIA Run:ai logs, follow these steps:

### Prerequisites

* Ensure that you have administrator-level access to the Kubernetes cluster where NVIDIA Run:ai is installed.
* The NVIDIA Run:ai [CLI](../../../reference/cli/install-cli.md) must be installed.

### Step-by-step Instructions

1. Open a terminal on any machine where the NVIDIA Run:ai CLI is installed and configured with access to the Kubernetes cluster. The user running this command typically requires system administrator [permissions](../../../platform-management/authentication/roles.md), including access to pods and namespaces.
2.  Collect the Logs. Execute the following command to collect the logs. See the [CLI commands reference](../../../reference/cli/runai/runai_diagnostics.md) for more details:

    ```bash
    runai diagnostics collect-logs
    ```

This command gathers diagnostic logs from your Kubernetes cluster to facilitate troubleshooting or support requests with NVIDIA Run:ai Support. The following optional flags are available:

| Flag                  | Description                                                                                                                                                                         |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--output-dir <path>` | Directory where the log archive is saved. Defaults to the directory from which the command is run.                                                                                  |
| `--namespaces <list>` | Comma-separated list of namespaces to collect logs from. Defaults to `runai`, `runai-reservation`, `runai-backend`, `training-operator`, `knative`, `gpu-operator`, `nim-operator`. |
| `--no-previous`       | Excludes previous pod logs from the collection.                                                                                                                                     |

3. After the command completes, the CLI displays the path of the generated compressed log file. Send this file to NVIDIA Run:ai Support for troubleshooting.

<Note>

**Note**

The command collects diagnostic logs from Kubernetes namespaces in a NVIDIA Run:ai installation. By default, logs are collected from the following namespaces: `runai`, `runai-reservation`, `runai-backend`, `training-operator`, `knative`, `gpu-operator`, and `nim-operator`. Use the `--namespaces` flag to target specific namespaces.

</Note>

## Logs Verbosity

Increase log verbosity to capture more detailed information, providing deeper insights into system behavior and make it easier to identify and resolve issues.

### Prerequisites

Before you begin, ensure you have the following:

* Access to the Kubernetes cluster where NVIDIA Run:ai is installed
  * Including [necessary permissions](../../../platform-management/authentication/roles.md) to view and modify configurations.
* kubectl installed and configured:
  * The Kubernetes command-line tool, `kubectl`, must be installed and configured to interact with the cluster.
  * Sufficient privileges to edit configurations and view logs.
* Monitoring Disk Space
  * When enabling verbose logging, ensure adequate disk space to handle the increased log output, especially when enabling debug or high verbosity levels.

### Adding Verbosity

<Accordion title="Adding verbosity to Prometheus">

**Adding verbosity to Prometheus**

To increase the logging verbosity for Prometheus, follow these steps:

1.  Edit the `RunaiConfig` to adjust Prometheus log levels. Copy the following command to your terminal:

    ```bash
    kubectl edit runaiconfig runai -n runai
    ```
2.  In the configuration file that opens, add or modify the following section to set the log level to `debug`:

    ```bash
    spec:
      prometheus:
        spec:
          logLevel: debug
    ```
3.  Save the changes. To view the Prometheus logs with the new verbosity level, run:

    ```bash
    kubectl logs -n runai prometheus-runai-0 
    ```

    This command streams the last 100 lines of logs from Prometheus, providing detailed information useful for debu

</Accordion>

<Accordion title="Adding verbosity to the Scheduler">

**Adding verbosity to the Scheduler**

To enable extended logging for the NVIDIA Run:ai scheduler:

1.  Edit the `RunaiConfig` to adjust scheduler verbosity:

    ```bash
    kubectl edit runaiconfig runai -n runai
    ```
2.  Add or modify the following section under the scheduler settings:

    ```bash
    runai-scheduler:
      args:
        verbosity: 6
    ```

    This increases the verbosity level of the scheduler logs to provide more detailed output.

**Warning:** Enabling verbose logging can significantly increase disk space usage. Monitor your storage capacity and adjust the verbosity level as necessary.

</Accordion>
