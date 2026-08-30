package flat_workloads

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
)

// ComplementEnvironmentWithSpecificEnv overrides asset fields with values from the specific environment.
// the fields that are "meshed" are
// - command/args
// - uid/gid/supplemental groups
// - environment variables
func ComplementEnvironmentWithSpecificEnv(environmentAsset *dbmodel.EnvironmentAssetRec, specificEnv *api.SpecificRunParams) *api.Error {
	if specificEnv == nil {
		return nil
	}

	//
	//   command and args from the specific env overrides the default command and args from the asset
	//
	if specificEnv.Command != nil {
		environmentAsset.Command = specificEnv.Command
	}
	if specificEnv.Args != nil {
		environmentAsset.Args = specificEnv.Args
	}

	if specificEnv.EnvironmentVariables != nil {
		environmentAsset.EnvironmentVariables = specificEnv.EnvironmentVariables
	}

	//
	//  unify uid/gid/supplemental
	//
	runAsUid, runAsGid, supplementalGroups, err := DetermineUidAndGid(specificEnv, environmentAsset.EnvironmentAssetSpec)
	if err != nil {
		return err
	}
	environmentAsset.RunAsUid = runAsUid
	environmentAsset.RunAsGid = runAsGid
	environmentAsset.SupplementalGroups = supplementalGroups

	return nil
}
