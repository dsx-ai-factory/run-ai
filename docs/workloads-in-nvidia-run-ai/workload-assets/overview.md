# Workload Assets

NVIDIA Run:ai [workload](../workloads.md) assets are preconfigured building blocks that simplify the workload submission effort and remove the complexities of Kubernetes and networks for AI practitioners.

Workload assets enable organizations to:

* Create and reuse preconfigured setup for code, data, storage and resources to be used by AI practitioners to simplify the process of submitting workloads
* Share the preconfigured setup with a wide audience of AI practitioners with similar needs

<Info>

**Note**

* The creation of assets is possible only via API and the NVIDIA Run:ai UI.
* The submission of workloads using assets, is possible only via the NVIDIA Run:ai UI.

</Info>

## Workload Asset Types

There are four workload asset types used by the workload:

* [Environments](environments.md)\
  The container image, tools and connections for the workload
* [Data sources](overview.md)\
  The type of data, its origin and the target storage location such as PVCs or cloud storage buckets where datasets are stored
* [Compute resources](compute-resources.md)\
  The compute specification, including GPU and CPU compute and memory
* [Credentials](credentials.md)\
  The secrets to be used to access sensitive data, services, and applications such as docker registry or S3 buckets

## Asset Scope

When a workload asset is created, a [scope](../../platform-management/aiinitiatives/adapting-ai-initiatives.md#scopes-in-an-organization) is required. The scope defines who in the organization can view and/or use the asset.

<Info>

**Note**

When an asset is created via API, the scope can be the entire account. This is currently an experimental feature.

</Info>

## Who Can Create an Asset?

Any subject (user, service account, or SSO group) with a [role](../../infrastructure-setup/authentication/roles.md) that has permissions to **Create** an asset, can do so within their scope.

## Who Can Use an Asset?

Assets are used when submitting workloads. Any subject (user, service account or SSO group) with a [role](../../infrastructure-setup/authentication/roles.md) that has permissions to **Create** workloads, can also use assets.

## Who Can View an Asset?

Any subject (user, service account, or SSO group) with a [role](../../infrastructure-setup/authentication/roles.md) that has permission to **View** an asset, can do so within their scope.
