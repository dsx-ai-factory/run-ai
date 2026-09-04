# RunAI Cluster SDK

## Adding new CRD version
* Run `hack/new-crd-version.sh <GROUPNAME> <VERSION> <KIND> [NAMESPACED]` to generate boilerplate for your new CRD or version of one.
* Group names should be `run` or `run.ai` (which will both be used as `run.ai`) or `scheduling`/`scheduling.run.ai`
* Try to only add fields between CRD versions, so that the automatic conversion will work. This is more important for CRDs such as PodGroups which are only created onces than to Projects/Departments that are being synced by the agent.
* You will have to manually move the `storageversion` tag to the new version of the CRD and add the `genclient` tag to generate listers/informers/clients.
* Add the new version to the `hack/update-client.sh` script to generate listers/informers/clients.
* run `make generate manifests generate-client` to generate the CRD new version, go clients, listers, informers etc (CICD will make you run it)
* You can use `hack/copy-crds.sh` to copy the output CRDs to the chart where they will be later moved to the be used in public.

