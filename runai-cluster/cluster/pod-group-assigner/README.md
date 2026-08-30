# pod-group-assigner
The pod-group-assigner has a controller for pod-groups;
it assigns the pod groups to (multi) node pools.

## Description
A runaijob (and other job types we support) can be submitted to multiple NodePools;
and each NP has its own runai-scheduler.

The pod-group-assigner will assign the PodGroup to all relevant NodePools,
one by one, until successfully scheduled.

It has a reconcile-loop on PodGroups, and also on NodePools.
For each pod-group, will convert the NodeAffinity of the pod group's pods,
and find the requested node-pool list (could be just ["default"]).

The assignment is as follows -
* changing "runai/node-pool" label of the Pod Group
* changing podGroup.Spec.Queue
* changing podGroup.Spec.MarkUnschedulable (if needed)
* changing podGroup.Spec.SchedulingBackoff (if needed)

see detailed design and flow chart in:
https://runai.atlassian.net/wiki/spaces/EN/pages/2436727142/28.11+New+3+API+Scheduler+--+PodGroupAssigner


## Testing

The tests for the pod-group-assigner are not exactly unittests, but rather environment - using a k8s common practice of *testEnv*.
To run tests for the first time, run:

**make test-kind**

this will download all go packages needed for the testing, and will install a kind cluster.

If you want to repeatedly run the tests, and you already ran the first command so you have the kind cluster installed, run:

**make test-kind-dev** 