/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubeconfig

import (
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/auth"
	internalauth "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/siderolabs/go-kubeconfig"
	authv1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	ErrKubConfigNil        = errors.New("kubeconfig file is doesn't exist or empty")
	ErrPathIsEmpty         = errors.New("path is empty")
	ErrUsernameEmpty       = errors.New("username is empty")
	ErrCurrentContextEmpty = errors.New("kubeconfig current context is empty")
)

type Handler struct {
	merger      *kubeconfig.Merger
	path        string
	contextName string
}

func NewHandler(merger *kubeconfig.Merger, path, contextName string) (*Handler, error) {
	if merger == nil {
		return nil, ErrKubConfigNil
	}
	if merger.AuthInfos == nil {
		merger.AuthInfos = make(map[string]*api.AuthInfo)
	}

	if path == "" {
		return nil, ErrPathIsEmpty
	}

	if contextName == "" {
		return nil, ErrUsernameEmpty
	}

	return &Handler{
		merger:      merger,
		path:        path,
		contextName: contextName,
	}, nil
}

func (h *Handler) SetTokenExecCmd() {
	authInfo := api.NewAuthInfo()
	authInfo.Exec = &api.ExecConfig{
		Command:         consts.AppName,                                                                                                    // runai
		Args:            []string{auth.AuthUse, auth.GetTokenUse, fmt.Sprintf("-%s", auth.FlagShortOutput), internalauth.KubeconfigFormat}, // auth get-token -o kubeconfig
		APIVersion:      authv1.SchemeGroupVersion.String(),
		InteractiveMode: api.NeverExecInteractiveMode,
	}

	h.merger.AuthInfos[h.contextName] = authInfo
}

func (h *Handler) SetContext() error {
	if h.merger.CurrentContext == "" {
		return ErrCurrentContextEmpty
	}

	context, ok := h.merger.Contexts[h.merger.CurrentContext]
	if !ok {
		return fmt.Errorf("kubeconfig current context, has no context with name: %s", h.merger.CurrentContext)
	}

	context.AuthInfo = h.contextName
	return nil
}

func (h *Handler) Update() error {
	return h.merger.Write(h.path)
}

func (h *Handler) getMerger() *kubeconfig.Merger { //nolint:unused // for testing
	return h.merger
}
