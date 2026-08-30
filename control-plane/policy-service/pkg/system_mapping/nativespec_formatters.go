package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemNativeSpecFormatterFields builds the runai-nativeSpec formatter
// for enum-enforced fields of Run.ai native workload specs (WorkspaceSpecSpec,
// DistributedSpecSpec, MasterSpec).
func BuildSystemNativeSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	addNativeSpecField(m, "imagePullPolicy", `{
	  "imagePullPolicy": $ARGS.named.value
	}`, nil, nil,
		map[string][]string{"value": {"Always", "Never", "IfNotPresent"}})

	addNativeSpecField(m, "preemptibility", `{
	  "preemptibility": $ARGS.named.value
	}`, nil, nil,
		map[string][]string{"value": {"preemptible", "non-preemptible"}})

	addNativeSpecField(m, "restartPolicy", `{
	  "restartPolicy": $ARGS.named.value
	}`, nil, nil,
		map[string][]string{"value": {"Always", "Never", "OnFailure"}})

	addNativeSpecField(m, "compute-GpuRequestType", `{
	  "compute": {
	    "gpuRequestType": $ARGS.named.value
	  }
	}`, nil, nil,
		map[string][]string{"value": {"portion", "memory"}})

	addNativeSpecField(m, "security-SeccompProfileType", `{
	  "security": {
	    "seccompProfileType": $ARGS.named.value
	  }
	}`, nil, nil,
		map[string][]string{"value": {"RuntimeDefault", "Unconfined", "Localhost"}})

	addNativeSpecField(m, "security-UidGidSource", `{
	  "security": {
	    "uidGidSource": $ARGS.named.value
	  }
	}`, nil, nil,
		map[string][]string{"value": {"fromTheImage", "fromIdpToken", "custom"}})

	addNativeSpecField(m, "cleanPodPolicy", `{
	  "cleanPodPolicy": $ARGS.named.value
	}`, nil, nil,
		map[string][]string{"value": {"None", "All", "Running"}})

	addNativeSpecField(m, "distributedFramework", `{
	  "distributedFramework": $ARGS.named.value
	}`, nil, nil,
		map[string][]string{"value": {"MPI", "PyTorch", "TF", "XGBoost", "JAX"}})

	addNativeSpecField(m, "mpiLauncherCreationPolicy", `{
	  "mpiLauncherCreationPolicy": $ARGS.named.value
	}`, nil, nil,
		map[string][]string{"value": {"AtStartup", "WaitForWorkersReady"}})

	addItemizedNativeSpecFields(m)

	return m
}

// addItemizedNativeSpecFields registers the "create a new array item" fields
// for every itemized field of the native spec. Each field carries a SetExpr
// that builds one array element from $ARGS.named.*, an ArrayKey naming the
// merge key column, and ArgsInfo/ArgsEnum describing the non-string / enum
// columns. Unlike the per-column "<array>-<Column>" leaves (which use item
// selectors to target a column across selected rows), these have no item
// selector and instead expose the full set of item columns so the UI can
// render a form to add a new item.
func addItemizedNativeSpecFields(m map[string]api.FormatterField) {
	intT := api.MAPPINGFIELDTYPE_INTEGER
	boolT := api.MAPPINGFIELDTYPE_BOOLEAN
	quantity := api.MAPPINGFIELDTYPE_QUANTITY
	strings := api.MAPPINGFIELDTYPE_STRINGS

	////////////////////////////////////////////////////////
	// labels (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "labels", `{
	  "labels": [
	    { "name": $ARGS.named.name, "value": $ARGS.named.value }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// annotations (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "annotations", `{
	  "annotations": [
	    { "name": $ARGS.named.name, "value": $ARGS.named.value }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// environmentVariables (keyed by name). All source kinds live in one
	// template; unset sources are pruned, so a rule that sets only "value"
	// yields a literal env var, one that sets configMapName/configMapKey
	// yields a configMap-sourced env var, etc.
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "environmentVariables", `{
	  "environmentVariables": [
	    {
	      "name": $ARGS.named.name,
	      "value": $ARGS.named.value,
	      "description": $ARGS.named.description,
	      "configMap": { "name": $ARGS.named.configMapName, "key": $ARGS.named.configMapKey },
	      "secret": { "name": $ARGS.named.secretName, "key": $ARGS.named.secretKey },
	      "podFieldRef": { "path": $ARGS.named.podFieldRefPath },
	      "userCredential": { "name": $ARGS.named.userCredentialName, "key": $ARGS.named.userCredentialKey }
	    }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// imagePullSecrets (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "imagePullSecrets", `{
	  "imagePullSecrets": [
	    { "name": $ARGS.named.name, "userCredential": $ARGS.named.userCredential }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"userCredential": boolT}, nil)

	////////////////////////////////////////////////////////
	// tolerations (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "tolerations", `{
	  "tolerations": [
	    {
	      "name": $ARGS.named.name,
	      "key": $ARGS.named.key,
	      "value": $ARGS.named.value,
	      "effect": $ARGS.named.effect,
	      "operator": $ARGS.named.operator,
	      "seconds": $ARGS.named.seconds
	    }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"seconds": intT}, nil)

	////////////////////////////////////////////////////////
	// ports (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "ports", `{
	  "ports": [
	    {
	      "name": $ARGS.named.name,
	      "container": $ARGS.named.container,
	      "external": $ARGS.named.external,
	      "serviceType": $ARGS.named.serviceType,
	      "toolName": $ARGS.named.toolName,
	      "toolType": $ARGS.named.toolType
	    }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"container": intT, "external": intT}, nil)

	////////////////////////////////////////////////////////
	// exposedUrls (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "exposedUrls", `{
	  "exposedUrls": [
	    {
	      "name": $ARGS.named.name,
	      "container": $ARGS.named.container,
	      "url": $ARGS.named.url,
	      "authorizationType": $ARGS.named.authorizationType,
	      "authorizedUsers": $ARGS.named.authorizedUsers,
	      "authorizedGroups": $ARGS.named.authorizedGroups,
	      "toolName": $ARGS.named.toolName,
	      "toolType": $ARGS.named.toolType
	    }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"container": intT, "authorizedUsers": strings, "authorizedGroups": strings},
		map[string][]string{"authorizationType": {"authenticatedUsers", "authorizedUsers", "authorizedGroups"}})

	////////////////////////////////////////////////////////
	// relatedUrls (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "relatedUrls", `{
	  "relatedUrls": [
	    { "name": $ARGS.named.name, "url": $ARGS.named.url, "type": $ARGS.named.type }
	  ]
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// compute.extendedResources (keyed by resource)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "compute-ExtendedResources", `{
	  "compute": {
	    "extendedResources": [
	      { "resource": $ARGS.named.resource, "quantity": $ARGS.named.quantity }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "resource"}, nil, nil)

	addItemizedStorageFields(m, intT, boolT, quantity)
}

// addItemizedStorageFields registers the "create a new array item" fields for
// the storage.* itemized arrays.
func addItemizedStorageFields(m map[string]api.FormatterField, intT, boolT, quantity api.MappingFieldType) {
	////////////////////////////////////////////////////////
	// storage.pvc (keyed by name). Both the "existing PVC" (claimName) and the
	// "new PVC" (claimInfo) shapes share one template; unset branches prune.
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-Pvc", `{
	  "storage": {
	    "pvc": [
	      {
	        "name": $ARGS.named.name,
	        "path": $ARGS.named.path,
	        "claimName": $ARGS.named.claimName,
	        "existingPvc": $ARGS.named.existingPvc,
	        "readOnly": $ARGS.named.readOnly,
	        "ephemeral": $ARGS.named.ephemeral,
	        "dataSharing": $ARGS.named.dataSharing,
	        "claimInfo": {
	          "size": $ARGS.named.size,
	          "storageClass": $ARGS.named.storageClass,
	          "volumeMode": $ARGS.named.volumeMode,
	          "accessModes": {
	            "readWriteOnce": $ARGS.named.readWriteOnce,
	            "readOnlyMany": $ARGS.named.readOnlyMany,
	            "readWriteMany": $ARGS.named.readWriteMany
	          },
	          "addedAttrValues": $ARGS.named.addedAttrValues
	        }
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{
			"existingPvc": boolT, "readOnly": boolT, "ephemeral": boolT, "dataSharing": boolT,
			"size": quantity, "readWriteOnce": boolT, "readOnlyMany": boolT, "readWriteMany": boolT,
			"addedAttrValues": api.MAPPINGFIELDTYPE_OBJECT,
		}, nil)

	////////////////////////////////////////////////////////
	// storage.hostPath (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-HostPath", `{
	  "storage": {
	    "hostPath": [
	      {
	        "name": $ARGS.named.name,
	        "mountPath": $ARGS.named.mountPath,
	        "path": $ARGS.named.path,
	        "readOnly": $ARGS.named.readOnly,
	        "mountPropagation": $ARGS.named.mountPropagation
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"readOnly": boolT}, nil)

	////////////////////////////////////////////////////////
	// storage.nfs (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-Nfs", `{
	  "storage": {
	    "nfs": [
	      {
	        "name": $ARGS.named.name,
	        "path": $ARGS.named.path,
	        "server": $ARGS.named.server,
	        "mountPath": $ARGS.named.mountPath,
	        "readOnly": $ARGS.named.readOnly
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"readOnly": boolT}, nil)

	////////////////////////////////////////////////////////
	// storage.git (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-Git", `{
	  "storage": {
	    "git": [
	      {
	        "name": $ARGS.named.name,
	        "repository": $ARGS.named.repository,
	        "branch": $ARGS.named.branch,
	        "revision": $ARGS.named.revision,
	        "path": $ARGS.named.path,
	        "passwordSecret": $ARGS.named.passwordSecret,
	        "secretKeyOfPassword": $ARGS.named.secretKeyOfPassword,
	        "secretKeyOfUser": $ARGS.named.secretKeyOfUser,
	        "secretRef": {
	          "name": $ARGS.named.secretRefName,
	          "authenticationMethod": $ARGS.named.authenticationMethod,
	          "secretKeyOfUser": $ARGS.named.secretRefSecretKeyOfUser,
	          "secretKeyOfPassword": $ARGS.named.secretRefSecretKeyOfPassword
	        }
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// storage.s3 (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-S3", `{
	  "storage": {
	    "s3": [
	      {
	        "name": $ARGS.named.name,
	        "bucket": $ARGS.named.bucket,
	        "path": $ARGS.named.path,
	        "url": $ARGS.named.url,
	        "accessKeySecret": $ARGS.named.accessKeySecret,
	        "awsSessionToken": $ARGS.named.awsSessionToken,
	        "secretKeyOfAccessKeyId": $ARGS.named.secretKeyOfAccessKeyId,
	        "secretKeyOfSecretKey": $ARGS.named.secretKeyOfSecretKey
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// storage.configMapVolume (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-ConfigMapVolume", `{
	  "storage": {
	    "configMapVolume": [
	      {
	        "name": $ARGS.named.name,
	        "configMap": $ARGS.named.configMap,
	        "mountPath": $ARGS.named.mountPath,
	        "subPath": $ARGS.named.subPath,
	        "defaultMode": $ARGS.named.defaultMode
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// storage.secretVolume (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-SecretVolume", `{
	  "storage": {
	    "secretVolume": [
	      {
	        "name": $ARGS.named.name,
	        "secret": $ARGS.named.secret,
	        "mountPath": $ARGS.named.mountPath,
	        "defaultMode": $ARGS.named.defaultMode
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"}, nil, nil)

	////////////////////////////////////////////////////////
	// storage.dataVolume (keyed by id)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-DataVolume", `{
	  "storage": {
	    "dataVolume": [
	      { "id": $ARGS.named.id, "mountPath": $ARGS.named.mountPath }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "id"}, nil, nil)

	////////////////////////////////////////////////////////
	// storage.emptyDirVolume (keyed by name)
	////////////////////////////////////////////////////////
	addNativeSpecFieldWithArgs(m, "storage-EmptyDirVolume", `{
	  "storage": {
	    "emptyDirVolume": [
	      {
	        "name": $ARGS.named.name,
	        "path": $ARGS.named.path,
	        "medium": $ARGS.named.medium,
	        "sizeLimit": $ARGS.named.sizeLimit
	      }
	    ]
	  }
	}`, nil, &api.FormatterArrayKey{ArgName: "name"},
		map[string]api.MappingFieldType{"sizeLimit": quantity}, nil)
}

func addNativeSpecField(
	dst map[string]api.FormatterField,
	name, setExpr string,
	getExpr *string,
	arrayKey *api.FormatterArrayKey,
	argsEnum map[string][]string,
) {
	addNativeSpecFieldWithArgs(dst, name, setExpr, getExpr, arrayKey, nil, argsEnum)
}

// addNativeSpecFieldWithArgs is the full form used by the itemized "create a
// new array item" fields: on top of the enum constraints (argsEnum) it declares
// the non-string column types via argsInfo. The key column is named by
// arrayKey.ArgName and does not need an argsInfo entry (it is a string).
func addNativeSpecFieldWithArgs(
	dst map[string]api.FormatterField,
	name, setExpr string,
	getExpr *string,
	arrayKey *api.FormatterArrayKey,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{SetExpr: &setExpr, GetExpr: getExpr, ArrayKey: arrayKey}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
