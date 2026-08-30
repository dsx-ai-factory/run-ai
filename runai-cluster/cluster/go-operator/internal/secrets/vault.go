package secrets

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type Vault struct {
	name       string
	icm        *InstallationConfigurationMapper
	secretMap  map[string][]byte
	deployment *v1.Secret
}

func NewVault(name string, ic, deployment *v1.Secret) (*Vault, error) {
	if deployment == nil {
		return nil, fmt.Errorf("deployment secret is nil")
	}
	icm, err := NewInstallationConfigurationMapper(ic)
	if err != nil {
		return nil, err
	}
	return &Vault{name: name, icm: icm, secretMap: make(map[string][]byte), deployment: deployment}, nil
}

func (v *Vault) Name() string {
	return v.name
}

func (v *Vault) AddFromSecret(secretPath SecretPath, keyName string) {
	value := v.icm.GetKeyBySP(secretPath)
	if value == "" {
		return
	}

	v.AddKeyValueSecret(keyName, value)
}

func (v *Vault) AddFromSecretByPath(viperPath string, keyName string) {
	value := v.icm.GetKeyByPath(viperPath)
	if value == "" {
		return
	}

	v.AddKeyValueSecret(keyName, value)
}

func (v *Vault) AddKeyValueSecret(keyName string, secret string) {
	v.secretMap[keyName] = []byte(secret)
}

func (v *Vault) Generate() (*v1.Secret, error) {
	if len(v.secretMap) == 0 {
		return v.deployment, nil
	}
	var jsonMapper = make(map[string]string)
	for key, value := range v.secretMap {
		jsonMapper[key] = string(value)
	}

	data, err := json.Marshal(jsonMapper)
	if err != nil {
		return nil, err
	}
	if v.deployment.Data == nil {
		v.deployment.Data = make(map[string][]byte)
	}
	v.deployment.Data[v.name] = data

	return v.deployment, nil
}
