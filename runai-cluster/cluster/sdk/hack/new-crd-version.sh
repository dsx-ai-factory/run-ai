#! /bin/bash

function help() {
    echo "usage: $0 <GROUPNAME> <VERSION> <KIND> [NAMESPACED]"
}

function add_adapter_functions() {
    cat <<EOF >> $GROUPVERSIONFILE

// SchemeGroupVersion is the group version used to register these objects.
var SchemeGroupVersion = GroupVersion

// Resource takes an unqualified resource and returns a Group-qualified GroupResource.
func Resource(resource string) schema.GroupResource {
    return SchemeGroupVersion.WithResource(resource).GroupResource()
}
EOF
}

if [ $# -lt 3 ]; then
    help
    exit 1
fi

OS=$(uname)

GROUP=`echo $1 | cut -d. -f1`
VERSION=$2
KIND=$3
if [ $# = 4 ] ; then
    NAMESPACED=true
else
    NAMESPACED=false
fi

GROUPVERSIONFILE="apis/$GROUP/$VERSION/groupversion_info.go"

if [ $GROUP == "run" ] ; then
    FULLGROUP="run.ai"
    operator-sdk create api --group $GROUP --version $VERSION --kind $KIND --resource=true --controller=false --namespaced=$NAMESPACED
    if [ $OS == "Linux" ] ; then
        sed -i s/run.run/run/g $GROUPVERSIONFILE
    else
        echo "boom: ${OS}"
        sed -i '' 's/run.run/run/g' $GROUPVERSIONFILE
    fi
else
    FULLGROUP="$GROUP.run.ai"
    operator-sdk create api --group $GROUP --version $VERSION --kind $KIND --resource=true --controller=false --namespaced=$NAMESPACED
fi


cat <<EOF > apis/$GROUP/$VERSION/doc.go
//+groupName=$FULLGROUP
package $VERSION
EOF

grep -q "SchemeGroupVersion" $GROUPVERSIONFILE
if [ $? -ne 0 ] ; then
    add_adapter_functions
fi
