package step

import "github.com/run-ai/runai/runai-installer/clients/runai/cluster"

func ClusterAuthValue(installInfo *cluster.ClusterInstallationInfoResponse) string {
	if installInfo.BootstrapToken != "" {
		return "controlPlane.bootstrapToken=" + installInfo.BootstrapToken
	}
	return "controlPlane.clientSecret=" + installInfo.ClientSecret
}
