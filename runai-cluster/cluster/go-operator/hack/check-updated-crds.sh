#! /bin/bash

crd_upgrader_path="cluster/runai-crd-upgrader/crds"
runai_chart_path="cluster/runai-chart/charts/crds/crds"
generated_deepcopy="zz_generated.*.go"

script_full_path=$(dirname "$0")

# try to copy CRDs
$script_full_path/copy-crds.sh
exit_status=0

# check if one of the CRDs has changes
if git status --porcelain | grep -e "$crd_upgrader_path" -e "$runai_chart_path" | grep -q "^[ M]"; then
    echo "Changes:"
    git --no-pager diff --no-relative --name-only HEAD~1 -- | grep -e "$crd_upgrader_path" -e "$runai_chart_path" | while read file; do echo "Changes in $file:";  git --no-pager diff HEAD~1 -- "../../$file"; done
    echo "There are changes in CRDs - Failing, either add the changes to git or remove them"
    ((exit_status++))
fi

if git status --porcelain | grep -e "$generated_deepcopy" | grep -q "^[ M]"; then
    echo "Changes:"
    git --no-pager diff --no-relative --name-only HEAD~1 -- | grep -e "$generated_deepcopy" | while read file; do echo "Changes in $file:";  git --no-pager diff HEAD~1 -- "../../$file"; done
    echo "There are changes in types files - Failing, either add the changes to git or remove them"
    ((exit_status++))
fi

if [ "$exit_status" == "0" ]; then
    echo "There are no changes in CRDs - OK"
    exit 0
else
    exit $exit_status
fi
