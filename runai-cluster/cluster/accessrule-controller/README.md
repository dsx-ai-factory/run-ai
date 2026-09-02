# accessrule-controller 

The accessrule-controller is a controller for the run.ai.accessrule crds.
Reconciling each accessrule to reflect the correct RoleBindings in the cluster.
For example, if we have an accessrule that assigns a specific user the Role
to CRUD workloads in a specific project - 
the accessrule-controller will make sure there's a RoleBinding object, binding between the
Role to the user, in the project's namespace.
You can read more here -
https://runai.atlassian.net/wiki/spaces/EN/pages/2521006122/Design+-+Reflecting+RBAC+in+the+k8s+-+Authorization+Service .

## Testing

The tests for the accessrule-controller are not exactly unittests, but rather environment - using a k8s common practice of *testEnv*.
To run tests for the first time, run:

**make test-kind**

this will download all go packages needed for the testing, and will install a kind cluster.

If you want to repeatedly run the tests, and you already ran the first command so you have the kind cluster installed, run:

**make test-kind-dev** 