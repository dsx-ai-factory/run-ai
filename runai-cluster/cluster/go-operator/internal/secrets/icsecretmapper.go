package secrets

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
)

type SecretPath string

const (
	ClientName   SecretPath = "global.controlPlane.config.agent.client.name"
	ClientSecret SecretPath = "global.controlPlane.config.agent.client.secret"
	ClusterID    SecretPath = "global.controlPlane.clusterUUID"

	YamlType = "yaml"
)

type InstallationConfigurationMapper struct {
	v *viper.Viper
}

func NewInstallationConfigurationMapper(secret *v1.Secret) (*InstallationConfigurationMapper, error) {
	icm := &InstallationConfigurationMapper{}
	err := icm.init(secret)
	if err != nil {
		return nil, err
	}
	return icm, nil
}

func (icm *InstallationConfigurationMapper) init(secret *v1.Secret) error {
	return icm.loadSecret(secret)
}

func (icm *InstallationConfigurationMapper) GetKeyBySP(sp SecretPath) string {
	return icm.v.GetString(string(sp))
}

func (icm *InstallationConfigurationMapper) GetKeyByPath(sp string) string {
	return icm.v.GetString(sp)
}

func (icm *InstallationConfigurationMapper) loadSecret(secret *v1.Secret) error {
	v := viper.New()
	v.SetConfigType(YamlType)
	if secret == nil {
		return fmt.Errorf("secret is nil")
	}

	if secret.Data == nil {
		return fmt.Errorf("secret data is empty")
	}

	if secretData, ok := secret.Data[InstallationConfigurationSecretName]; ok {
		br := bytes.NewReader(secretData)
		err := v.ReadConfig(br)
		if err != nil {
			return err
		}
		icm.v = v
		return nil
	}
	return fmt.Errorf("data does not contain '%s'", InstallationConfigurationSecretName)
}
