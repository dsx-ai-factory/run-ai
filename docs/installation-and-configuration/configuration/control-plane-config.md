# Advanced Control Plane Configurations

Use this page to customize the NVIDIA Run:ai control plane beyond the default installation settings.

The settings on this page use NVIDIA Run:ai control plane Helm chart values. You can provide the same values through either installation path:

- Manual Helm - Set the values in a Helm values file or pass them as Helm flags.
- Guided Installation - Set the same values under `controlPlane.customHelmValues` in `config.yaml`. See [`controlPlane`](../installation/guided-installation/configuration-reference.md#control-plane-configuration).

## Helm Chart Values

The following Helm chart values can be used to customize the control plane for your environment. After changing a value, restart the relevant NVIDIA Run:ai pods so they can load the updated configuration.

| Key                                                                                                                                                                                                                                                       | Change                           | Description                                                                                                                                                                                                                                                                                 |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `global.ingress.ingressClass`                                                                                                                                                                                                                             | Ingress class                    | NVIDIA Run:ai uses NGINX as the default ingress controller. If your cluster has a different ingress controller, you can configure the ingress class to be created by NVIDIA Run:ai.                                                                                                         |
| `global.ingress.tlsSecretName`                                                                                                                                                                                                                            | TLS secret name                  | NVIDIA Run:ai requires the creation of a secret with [domain certificate](../system-requirements/networking-setup.md#control-plane-certificate). If the `runai-backend` namespace already had such a secret, you can set the secret name here |
| `<service-name>.podLabels`                                                                                                                                                                                                                                | Pod labels                       | Set NVIDIA Run:ai and 3rd party services' [Pod Labels ](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels)in a format of key/value pairs.                                                                                                                            |
| `<service-name>.tolerations`                                                                                                                                                                                                                              | Pod tolerations                  | Set NVIDIA Run:ai and third-party services' [Pod Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in list format. These tolerations apply only to the specified service.                                                                         |
| `global.tolerations`                                                                                                                                                                                                                                      | Pod tolerations                  | Set NVIDIA Run:ai and 3rd party services' [Pod Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) in list format. These tolerations are applied globally to all supported services.                                                                |
| `<service-name>.replicaCount`                                                                                                                                                                                                                             | Pod replicas                     | By default, NVIDIA Run:ai services are deployed with a single replica. Set this value to run the specified service with multiple pod replicas.                                                                                                                                              |
| `global.replicaCount`                                                                                                                                                                                                                                     | Pod replicas                     | By default, NVIDIA Run:ai services are deployed with a single replica. Set this value to apply a global replica count to all services.                                                                                                                                                      |
| `global.requireDefaultPodAntiAffinity`                                                                                                                                                                                                                    | Enable default pod anti-affinity | <p>When enabled, NVIDIA Run:ai applies a default pod anti-affinity rule that attempts to prevent pods belonging to the same service from being scheduled on the same node.<br/>Default: <code>true</code></p>                                                                                |
| <p><code>&#x3C;service-name></code><br/> <code>resources:</code><br/>  <code>limits:</code><br/>    <code>cpu: 500m</code><br/>    <code>memory: 512Mi</code><br/>  <code>requests:</code><br/>    <code>cpu: 250m</code><br/>    <code>memory: 256Mi</code></p> | Pod request and limits           | Set NVIDIA Run:ai and 3rd party services' resources                                                                                                                                                                                                                                         |
| `disableIstioSidecarInjection.enabled`                                                                                                                                                                                                                    | Disable Istio sidecar injection  | Disable the automatic injection of Istio sidecars across the entire NVIDIA Run:ai Control Plane services.                                                                                                                                                                                   |
| `global.affinity`                                                                                                                                                                                                                                         | System nodes                     | <p>Sets the system nodes where NVIDIA Run:ai system-level services are scheduled.<br/>Default: Prefer to schedule on nodes that are labeled with <code>node-role.kubernetes.io/runai-system</code></p>                                                                                       |
| `global.customCA.enabled`                                                                                                                                                                                                                                 | Certificate authority            | Enables the use of a custom Certificate Authority (CA) in your deployment. When set to `true`, the system is configured to trust a user-provided CA certificate for secure communication.                                                                                                   |

## Email Notifications Configuration <a href="#email-notifications-configuration" id="email-notifications-configuration"></a>

To enable and manage outbound email notifications for the NVIDIA Run:ai platform, configure the following values under the `notificationsService.config.sinks.runai-email` section in your Helm values. These settings are used both for workload-related notifications and for system messages from NVIDIA Run:ai. All parameters can be set in your `values.yaml` file during Helm deployment or via Helm upgrade.

| Parameter                                                            | Required | Default Value                     | Description                                                                                                                         |
| -------------------------------------------------------------------- | -------- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `notificationsService.config.sinks.runai-email.type`                 | required | email                             | Type of sink; must be set as **email**                                                                                              |
| `notificationsService.config.sinks.runai-email.smtp_host`            | required | Empty                             | SMTP server hostname                                                                                                                |
| `notificationsService.config.sinks.runai-email.smtp_port`            | optional | 587                               | SMTP server port                                                                                                                    |
| `notificationsService.config.sinks.runai-email.user`                 | optional | Empty                             | SMTP authentication username                                                                                                        |
| `notificationsService.config.sinks.runai-email.password`             | optional | Empty                             | SMTP authentication password                                                                                                        |
| `notificationsService.config.sinks.runai-email.smtp_tls_enabled`     | optional | true                              | Enable TLS for SMTP                                                                                                                 |
| `notificationsService.config.sinks.runai-email.from_display_name`    | optional | NVIDIA Run:ai                     | Display name for the sender email address                                                                                           |
| `notificationsService.config.sinks.runai-email.from`                 | optional | [test@run.ai](mailto:test@run.ai) | Sender email address, e.g. [notifications@my.domain](mailto:notifications@my.domain)                                                |
| `notificationsService.config.sinks.runai-email.direct_notifications` | optional | False                             | Send workload notifications directly to the submitter’s inbox instead of sending all notifications to a single global email address |
| `notificationsService.config.sinks.runai-email.auth_type`            | optional | auth\_login                       | Authentication type: auth\_login or auth\_plain                                                                                     |

### Example Configuration

```yaml
notificationsService:
  config:
    sinks:
      runai-email:
        type: email
        smtp_host: my.smtp.host
        smtp_port: 587
        user: smtp_user
        password: smtp_password
        smtp_tls_enabled: true
        from_display_name: NVIDIA Run:ai
        from: notifications@my.domain
        direct_notifications: true
        logo_url: https://s3.amazonaws.com/www.run.ai/nvidia_runai_logo_new.png
        auth_type: auth_login
```

