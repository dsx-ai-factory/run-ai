#! /bin/bash

destinations=(
    ../runai-chart/crds
    ../runai-crd-upgrader/crds
)

for destination in "${destinations[@]}"; do
    if [ ! -d "$destination" ]; then
        echo "Destination directory $destination does not exist"
        exit 1
    fi

    cp config/crd/bases/run.ai_runaiconfigs.yaml "$destination"/run_ai_v1_runaiconfig.yaml
    cp config/crd/bases/run.ai_clusters.yaml "$destination"/run.ai_clusters.yaml
    cp config/crd/bases/run.ai_remoteconfigs.yaml "$destination"/run.ai_remoteconfigs.yaml
    cp config/crd/bases/run.ai_remoteconfigsyncs.yaml "$destination"/run.ai_remoteconfigsyncs.yaml
    cp config/crd/bases/run.ai_managednodesconfigs.yaml "$destination"/run.ai_managednodesconfigs.yaml

    cp config/crd/bases/run.ai_nodepools.yaml "$destination"/run_ai_nodepool.yaml
    cp config/crd/bases/run.ai_accessrules.yaml "$destination"/run_ai_accessrule.yaml
    cp config/crd/bases/run.ai_externalworkloads.yaml "$destination"/run_ai_externalworkload.yaml
    cp config/crd/bases/run.ai_datavolumes.yaml "$destination"/run.ai_datavolumes.yaml

    # Enable the following line when we support the complete OLM flow
    # cp config/crd/bases/installation.run.ai_clusters.yaml "$destination"/installation_v1_cluster.yaml
done
