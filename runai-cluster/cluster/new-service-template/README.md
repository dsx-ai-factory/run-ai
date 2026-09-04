# new-service-template generator using cookiecutter

This is a cookiecutter template for a new service in the runai-operator repository.
It will automatically generate most of what you need when writing a new service for the cluster.
The full documentation of what it does can be found here:
https://runai.atlassian.net/wiki/spaces/EN/pages/2703524287/Tutorial+-+New+Service+Creation

## Usage
`cd runai-cluster/cluster`

`cookiecutter new-service-template`

 -> Enter the name of the service in the formats requested in the prompt.
 Additional params:
* `add_project_based_rolebinding`: if you need extra permissions only in projects' namespaces
(for example, "update" permissions on pods - we only want to have this permission in runai projects' namespaces,
and not as a cluster-scoped permission).
* `engine_instead_of_cluster`: if you are writing a service under "engine" instead of "cluster".
* `deploy_remote_instead_of_local_cluster`: Set to `true` if your operand should be deployed to the remote (tenant) cluster in a dual-cluster setup.
This is typically needed for:
  - Operands with webhooks
  - DaemonSets that need to run on GPU nodes
  - Components that depend on hardware resources (GPUs, specific nodes, hardware metrics exporters)
Default is `false`, which deploys to the local (management) cluster. Most Deployment-based controllers should use `false`.

After generating, please follow the instructions described in the link above.
