/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrity

import (
	"encoding/json"
	"testing"
)

// Producers describes which (secret, key) pairs the chart promises to provide,
// plus the set of secret names the chart designates as user-managed (opaque).
type Producers struct {
	Keys     map[string]map[string]struct{} // secret name -> set of declared keys
	External map[string]struct{}            // secret names whose contents are user-supplied
}

func newProducers() *Producers {
	return &Producers{
		Keys:     map[string]map[string]struct{}{},
		External: map[string]struct{}{},
	}
}

func (p *Producers) addKey(name, key string) {
	if name == "" || key == "" {
		return
	}
	if _, ok := p.Keys[name]; !ok {
		p.Keys[name] = map[string]struct{}{}
	}
	p.Keys[name][key] = struct{}{}
}

func (p *Producers) hasKey(name, key string) bool {
	keys, ok := p.Keys[name]
	if !ok {
		return false
	}
	_, ok = keys[key]
	return ok
}

func (p *Producers) hasSecret(name string) bool {
	_, ok := p.Keys[name]
	return ok
}

func (p *Producers) isExternal(name string) bool {
	_, ok := p.External[name]
	return ok
}

// secretGeneratorConfig mirrors the JSON written by the secret-generator-params-cm
// helper template. Only the fields we need are unmarshalled.
type secretGeneratorConfig struct {
	Secrets []struct {
		Name               string `json:"name"`
		ExistingSecretName string `json:"existingSecretName"`
		Keys               []struct {
			Name         string `json:"name"`
			ChartValue   string `json:"chartValue"`
			ExternalOnly bool   `json:"externalOnly"`
		} `json:"keys"`
	} `json:"secrets"`
}

func collectProducers(t *testing.T, m *Manifests) *Producers {
	t.Helper()
	p := newProducers()

	for _, s := range m.Secrets {
		for k := range s.Data {
			p.addKey(s.Name, k)
		}
		for k := range s.StringData {
			p.addKey(s.Name, k)
		}
	}

	for _, cm := range m.ConfigMaps {
		if cm.Labels["app.kubernetes.io/name"] != "secret-generator-config" {
			continue
		}
		raw, ok := cm.Data["secrets.json"]
		if !ok {
			t.Fatalf("ConfigMap %q: missing data.secrets.json", cm.Name)
		}
		var cfg secretGeneratorConfig
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			t.Fatalf("ConfigMap %q: parse secrets.json: %v", cm.Name, err)
		}
		for _, item := range cfg.Secrets {
			if isExternalEntry(item.Name, item.ExistingSecretName) {
				p.External[item.ExistingSecretName] = struct{}{}
				continue
			}
			for _, k := range item.Keys {
				if !willProduce(k.ExternalOnly, k.ChartValue) {
					continue
				}
				p.addKey(item.Name, k.Name)
			}
		}
	}

	return p
}

// An entry is external when it explicitly redirects to a different secret name.
// A self-reference (existingSecretName == name) is the chart's "no override"
// shape and is treated as the generator-managed Secret itself.
func isExternalEntry(name, existing string) bool {
	return existing != "" && existing != name
}

// externalOnly keys require chartValue (or legacy-secret migration, which we
// can't verify statically). Without one, the generator errors at install time,
// so the key is not guaranteed to exist.
func willProduce(externalOnly bool, chartValue string) bool {
	if !externalOnly {
		return true
	}
	return chartValue != ""
}
