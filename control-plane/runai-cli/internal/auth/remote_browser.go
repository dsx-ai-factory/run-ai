/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

type RemoteBrowserHandler struct {
	// using reader so we can mock it in tests
	reader io.Reader
}

func NewRemoteBrowserHandler(reader io.Reader) *RemoteBrowserHandler {
	return &RemoteBrowserHandler{reader: reader}
}

func (rb *RemoteBrowserHandler) Handle(_ context.Context, authUrl *url.URL) (string, error) {
	var code string
	fmt.Printf(consts.RemoteBrowserOpenUrlMsg, authUrl.String()) //nolint:forbidigo // user error
	readerInput := bufio.NewReader(rb.reader)
	codeRaw, err := readerInput.ReadString('\n')
	// remove any new lines
	code = strings.TrimSpace(codeRaw)
	if err != nil {
		return "", fmt.Errorf("%s %w", consts.FailedCopyCodeFromUser, err)
	}
	if code == "" {
		return "", errors.New(consts.EmptyCodeError)
	}
	return code, nil
}
