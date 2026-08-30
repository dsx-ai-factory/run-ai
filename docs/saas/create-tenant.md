# Create and Access the Control Plane Tenant

To access NVIDIA Run:ai SaaS, create a control plane tenant through the [NVIDIA GPU Cloud (NGC](https://org.ngc.nvidia.com)) catalog. The tenant is your organization's dedicated, NVIDIA-hosted Run:ai environment. The control plane provides the central management layer for NVIDIA Run:ai, handling multi-cluster management, resource and access management, as well as workload submission and monitoring. NVIDIA hosts and maintains the control plane. Once the tenant is provisioned, you can install the NVIDIA Run:ai cluster and connect it to the control plane, making it ready to run training, inference, and other workloads.

<Callout intent="info">
To create a control plane tenant, your NGC account must have an active entitlement for NVIDIA Run:ai SaaS.
</Callout>

## Prerequisites

An active **NVIDIA Run:ai SaaS** entitlement in your NGC organization. To activate your entitlement, see [Activating an NGC Product from an NVIDIA Commercial Entitlement Certificate](https://docs.nvidia.com/ngc/latest/ngc-user-guide.html#activating-ngc-product-from-an-nvidia-commercial-entitlement-certificate).

## Provision the Control Plane Tenant

<Callout intent="info">
To delegate provisioning to another user, go to [org.ngc.nvidia.com/users](https://org.ngc.nvidia.com/users) and assign them the **NVIDIA Run:ai SaaS Admin** role. They can then go to [runai.ngc.nvidia.com](https://runai.ngc.nvidia.com/) to perform the provisioning steps.
</Callout>

1. Log in to [NGC](https://org.ngc.nvidia.com) and navigate to **Catalog** -> **NVIDIA Run:ai SaaS**.

2. In the **Create Control Plane Tenant** dialog, fill in the following:
   * **Tenant Admin Email** - Pre-filled with your NGC account email. You can assign additional admins after provisioning is complete.
   * **Tenant Alias** - A unique subdomain name for your tenant URL (for example, `myorg` results in `myorg.nv.run.ai`).

3. Click **Create Control Plane Tenant**. NGC provisions your tenant. This may take a few minutes.

Once provisioning is complete, the following details are displayed:

* **Tenant Name** - An auto-generated identifier for the tenant.
* **Tenant Alias** - The alias you provided.
* **Tenant Domain** - The URL of your NVIDIA Run:ai control plane, based on your alias (for example, `myorg.nv.run.ai`).
* **Created** - The date and time the tenant was created.

<Callout intent="info">
This is the only time your **Tenant Alias** and **Tenant Domain** are displayed. If you want to access your tenant directly using the Tenant Domain URL, make a note of it now. Alternatively, you can always access your tenant by navigating to [runai.ngc.nvidia.com](https://runai.ngc.nvidia.com), which automatically redirects you.
</Callout>

An email is sent to the Tenant Admin Email address with a link to activate your account.

## Access the Control Plane Tenant

1. Open the activation email and click **SIGN IN TO NVIDIA RUN:AI**.

2. On the password setup page, enter a **New Password**, confirm it, and click **Reset password**.
3. Once your account is updated, click **Back to Application**.
4. On the NVIDIA Run:ai sign-in page, enter your email and new password, then click **Sign In**.
5. Review the **Run:ai End User License Agreement** and click **I Agree**.

## Next Steps

After signing in, the onboarding wizard opens automatically. Before proceeding, ensure your environment meets all [system](../installation-and-configuration/installation/install-using-helm/system-requirements.md) and [network](../installation-and-configuration/installation/install-using-helm/network-requirements.md) requirements. Then follow the wizard to [install](../installation-and-configuration/installation/install-using-helm/helm-install.md) the NVIDIA Run:ai cluster.
