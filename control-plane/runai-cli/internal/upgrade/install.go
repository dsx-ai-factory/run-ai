/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
)

const msiExecBin = "msiexec"

type MsiInstaller struct {
	msiExecBin string
	args       []string
}

func (msi *MsiInstaller) Install(ctx context.Context, path, _ string) (string, error) {
	msiExecPath, err := exec.LookPath(msi.msiExecBin)
	if err != nil {
		return "", errors.New(consts.MsiNotFound)
	}
	cmd := exec.CommandContext(ctx, msiExecPath, append(msi.args, path)...)

	var errBuffer bytes.Buffer
	cmd.Stderr = &errBuffer
	err = cmd.Run()

	if err != nil {
		return "", fmt.Errorf(consts.UpgradeMsiFailed, errBuffer.String())
	}
	return "", nil
}

func BaseNewMsiInstaller(quite bool) *MsiInstaller {
	return NewMsiInstaller(msiExecBin, quite)
}

func NewMsiInstaller(msiExecBin string, quite bool) *MsiInstaller {
	args := []string{"/i"}
	if quite {
		args = append(args, "/quiet")
	}

	return &MsiInstaller{
		msiExecBin: msiExecBin,
		args:       args,
	}
}

type UnixScriptInstaller struct {
	quite bool
}

func NewUnixScriptInstaller(quite bool) *UnixScriptInstaller {
	return &UnixScriptInstaller{
		quite: quite,
	}
}
func (usi *UnixScriptInstaller) Install(ctx context.Context, installScript, execShell string) (string, error) {
	cmd := exec.CommandContext(ctx, execShell, installScript, "--upgrade")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf(consts.UpgradeScriptFailed, exitErr.ProcessState, exitErr.Stderr)
		}
		return "", fmt.Errorf(consts.UpgradeScriptFailed, err.Error(), output)
	}

	if usi.quite {
		return "", nil
	}
	return string(output), nil
}
