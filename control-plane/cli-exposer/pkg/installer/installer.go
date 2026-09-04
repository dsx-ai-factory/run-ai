/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/consts"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"os"
	"path"
	"text/template"
)

var (
	installerBasePath         = "/app/installer/"
	installerUnixTemplatePath = "unix/installer_template.sh"
	binariesPath              = "/app/dist/"
)

type Settings struct {
	CpURL      string
	AuthURL    string
	Token      string
	SdkVersion string
	CpVersion  string
}

func GetInstallerCode(settings Settings) ([]byte, error) {
	unixTemplate := path.Join(installerBasePath, installerUnixTemplatePath)
	baseName := path.Base(unixTemplate)
	tmpl, err := template.New(baseName).ParseFiles(unixTemplate)
	if err != nil {
		logging.Errorf("%s", consts.ErrParsingInstallerTemplate+err.Error())
		return nil, errors.New(consts.ErrParsingInstallerTemplate)
	}

	var tplOutput bytes.Buffer
	err = tmpl.Execute(&tplOutput, settings)
	if err != nil {
		logging.Errorf("%s", consts.ErrExecInstallerTemplate+err.Error())
		return nil, errors.New(consts.ErrExecInstallerTemplate)
	}
	return tplOutput.Bytes(), nil
}

func ValidateBinaryPath(ctx context.Context, operationSystem, arch string) (string, error) {
	filePath := path.Join(binariesPath, operationSystem, arch, "runai")

	if _, err := os.Stat(filePath); err != nil {
		msg := fmt.Sprintf(consts.ErrBinaryNotExist, operationSystem, arch)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", msg)
		return "", errors.New(msg)
	}
	return filePath, nil
}
