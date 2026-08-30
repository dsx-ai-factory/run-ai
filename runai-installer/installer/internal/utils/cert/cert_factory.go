package cert

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
)

func GetCertHandler(sslCertPathParam, sslKeyPathParam, hostParam configinputs.InstallInputType, runaiData *model.RunAIInstallData) CertHandler {
	sslCertPath := runaiData.GetInputValue(sslCertPathParam)
	sslKeyPath := runaiData.GetInputValue(sslKeyPathParam)
	if (sslCertPath == "" || sslKeyPath == "") && runaiData.Config.DevMode {
		return NewGenerateSelfSignedCert(runaiData.GetInputValue(hostParam))
	}
	return NewLoadCertFromPath(sslCertPath, sslKeyPath)
}
