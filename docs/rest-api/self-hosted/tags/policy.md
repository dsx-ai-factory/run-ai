Policies allow administrators to impose restrictions and set default values for researcher workloads.
Restrictions and default values can be placed on CPUs, GPUs, and other resources or entities.
For more information, see [Policies](/run-ai/self-hosted/documentation/platform-management/policies/policies-and-rules).

Use the Policy endpoints to list applied policies, and to create, retrieve, update, and delete policies for an organizational unit.

## Policies by Workload Type

Training, workspace, distributed, inference, distributed inference, and NVIDIA NIM service policies each have their own endpoints. For each type, you can retrieve the policy of an organizational unit, update part of it, overwrite it, or delete it.

## Policies by Policy ID

Endpoints marked experimental create a policy for a specific workload type. You can then list policies with filtering, sorting, and pagination, and retrieve, replace, patch, or delete a policy by its ID, including deleting individual rules from it. Retrieving a policy returns both its declared rules and its effective rules. You can also retrieve the effective policy for a given type and scope, which combines the policy of that scope, if one exists, with the effective policies of its parent scopes.

Experimental endpoints can also retrieve the field, spec selector, and item selector names and descriptions for a workload type, and the detailed breakdown of one or more field references in its mapping.
