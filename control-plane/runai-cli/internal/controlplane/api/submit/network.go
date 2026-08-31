/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/network/externalurl"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/network/port"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/spf13/pflag"
)

type Network struct {
	Ports        []string
	ExternalUrls []string
}

func (n *Network) GeneratePorts() ([]openapi.Port, error) {
	var (
		ports []openapi.Port
	)

	list, err := port.NewPortList(n.Ports)
	if err != nil {
		return nil, err
	}
	ports = append(ports, list...)
	return ports, nil
}

func (n *Network) GenerateExternalUrls() ([]openapi.ExposedUrl, error) {
	var (
		eUrlList []openapi.ExposedUrl
	)

	list, err := externalurl.NewExternalUrlList(n.ExternalUrls)
	if err != nil {
		return nil, err
	}
	eUrlList = append(eUrlList, list...)
	return eUrlList, nil
}

func NetworkFlagSet(flagsNetwork *pflag.FlagSet, specNetwork *Network) {
	flagsNetwork.StringArrayVar(&specNetwork.Ports, consts.AttributeNamePorts, []string{}, consts.AttributeUsagePorts)
	flagsNetwork.StringArrayVar(&specNetwork.ExternalUrls, consts.AttributeNameExternalUrls, []string{}, consts.AttributeUsageExternalUrls)
}
