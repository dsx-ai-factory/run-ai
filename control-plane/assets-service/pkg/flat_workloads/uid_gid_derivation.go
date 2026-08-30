package flat_workloads

import "github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"

const (
	UidGidFromTheImage = api.UidGidSource("fromTheImage")
	CustomUidGid       = api.UidGidSource("custom")
	// UidGidFromIdpToken one is similar to custom, except that it indicates to the frontend to
	// attempt to load uid/gid from the idp token instead of asking the user to provide it
	UidGidFromIdpToken = api.UidGidSource("fromIdpToken")
)

func DetermineUidAndGid(specificEnv *api.SpecificRunParams, envAssetSpec api.EnvironmentAssetSpec) (*int64, *int64, *string, *api.Error) {
	var runAsUid *int64 = nil
	var runAsGid *int64 = nil
	var supplementalGroups *string = nil

	// if specific env is not provided - take the (optional) values from the environment asset spec
	if specificEnv == nil {
		runAsUid = envAssetSpec.RunAsUid
		runAsGid = envAssetSpec.RunAsGid
		supplementalGroups = envAssetSpec.SupplementalGroups
		return runAsUid, runAsGid, supplementalGroups, nil
	}

	if specificEnv.RunAsUid == nil {
		// if not value given in the specific env - take the (optional) default value from the environment asset spec
		runAsUid = envAssetSpec.RunAsUid
	} else {
		if *specificEnv.RunAsUid >= 0 {
			runAsUid = specificEnv.RunAsUid
		}
	}

	if specificEnv.RunAsGid == nil {
		// if not value given in the specific env - take the (optional) default value from the environment asset spec
		runAsGid = envAssetSpec.RunAsGid
	} else {
		if *specificEnv.RunAsGid >= 0 {
			runAsGid = specificEnv.RunAsGid
		}
	}

	if specificEnv.SupplementalGroups == nil {
		// if not value given in the specific env - take the (optional) default value from the environment asset spec
		supplementalGroups = envAssetSpec.SupplementalGroups
	} else {
		if *specificEnv.SupplementalGroups != "" {
			supplementalGroups = specificEnv.SupplementalGroups
		}
	}

	return runAsUid, runAsGid, supplementalGroups, nil
}
