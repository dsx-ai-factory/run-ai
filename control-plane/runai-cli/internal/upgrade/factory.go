/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"context"
	"errors"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/os"
)

const (
	msiFileName        = "runaiCLI.installer"
	UnixScriptFileName = "runaiCLIInstaller.sh"
)

type Factory struct {
	unixFlow    Upgrader
	windowsFlow Upgrader
}

type Upgrader interface {
	Upgrade(ctx context.Context, execShell string) (string, error)
}

func (f *Factory) CreateFlow(osType os.Type) (Upgrader, error) {
	switch osType {
	case os.Windows:
		return f.windowsFlow, nil
	case os.Linux, os.Darwin:
		return f.unixFlow, nil
	default:
		return nil, errors.New("unsupported OS type")
	}
}

func CreateFactory(installerClient *cli.Client, quite bool) (*Factory, error) {
	cacheFullPath := instance.GetVersionCachePath()
	osType := os.DetectOsType()
	arcType := os.DetectArch()
	if osType == os.Unknown || arcType == os.UnknownArch {
		return nil, errors.New(consts.UnsupportedOSArch)
	}

	// msiDownloader := NewBinaryDownloader(installerClient, osType, arcType) // re-enable once MSI is fixed
	return &Factory{
		unixFlow: NewBasicFlow(NewScriptDownloader(installerClient), NewUnixScriptInstaller(quite), cacheFullPath, UnixScriptFileName),
		//windowsFlow: NewBasicFlow(msiDownloader, BaseNewMsiInstaller(quite), fullPath, msiFileName), // replace with the line below once MSI is fixed
		windowsFlow: NewStaticMessage(consts.WindowsUpgradeMessage),
	}, nil
}
