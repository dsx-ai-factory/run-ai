package business

type ClusterInstallationData struct {
	StackName          string
	EksClusterName     string
	TenantName         string
	ControlPlaneUrl    string
	InstallerAppId     string
	InstallerAppSecret string
	ClusterId          string
	ClusterUrl         string
}
