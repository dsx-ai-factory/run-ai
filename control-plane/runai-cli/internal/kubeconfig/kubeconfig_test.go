/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubeconfig

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gokubeconfig "github.com/siderolabs/go-kubeconfig"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	usernameTestName = "test-name"
)

var _ = Describe("Handler", Ordered, func() {
	var (
		kubePath string
	)

	BeforeAll(func() {
		var err error
		kubePath, err = GetHomePath()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Builder", func() {
		It("merger is nil", func() {
			_, err := NewHandler(nil, "", "")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrKubConfigNil))
		})

		It("path is empty", func() {
			_, err := NewHandler(&gokubeconfig.Merger{}, "", "")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrPathIsEmpty))
		})

		It("name is emtpy", func() {
			_, err := NewHandler(&gokubeconfig.Merger{}, kubePath, "")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrUsernameEmpty))
		})
	})

	Context("SetTokenExecCmd", func() {
		It("the exec config is correct", func() {
			merger := &gokubeconfig.Merger{}
			handler, err := NewHandler(merger, kubePath, usernameTestName)
			Expect(err).ToNot(HaveOccurred())

			merger = handler.getMerger()
			authInfo, ok := merger.AuthInfos[usernameTestName]
			Expect(ok).To(BeFalse())
			Expect(authInfo).To(BeNil())

			handler.SetTokenExecCmd()

			merger = handler.getMerger()
			authInfo, ok = merger.AuthInfos[usernameTestName]
			Expect(ok).To(BeTrue())

			Expect(authInfo.Token).To(Equal(""))

			Expect(authInfo.Exec).NotTo(BeNil())
			Expect(string(authInfo.Exec.InteractiveMode)).To(Equal("Never"))
			Expect(authInfo.Exec.Command).To(Equal("runai"))
			Expect(authInfo.Exec.APIVersion).To(Equal("client.authentication.k8s.io/v1"))
			Expect(authInfo.Exec.Args).To(ContainElements("auth", "get-token", "-o", "kubeconfig"))
		})
	})

	Context("SetContext", func() {
		It("the runai-cil user does not exist in the kubeconfig", func() {
			handler, err := NewHandler(&gokubeconfig.Merger{}, kubePath, usernameTestName)
			Expect(err).ToNot(HaveOccurred())
			err = handler.SetContext()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrCurrentContextEmpty))
		})

		It("current context does not exist is auth info map", func() {
			merger := &gokubeconfig.Merger{}
			merger.CurrentContext = "invalid-context"

			handler, err := NewHandler(merger, kubePath, usernameTestName)
			Expect(err).ToNot(HaveOccurred())
			err = handler.SetContext()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("kubeconfig current context, has no context with name: invalid-context"))
		})

		It("current context exist is auth info map", func() {
			merger := &gokubeconfig.Merger{}
			merger.Contexts = make(map[string]*api.Context)
			merger.Contexts["valid-context"] = &api.Context{}
			merger.CurrentContext = "valid-context"

			handler, err := NewHandler(merger, kubePath, usernameTestName)
			Expect(err).ToNot(HaveOccurred())
			err = handler.SetContext()
			Expect(err).ToNot(HaveOccurred())
			merger = handler.getMerger()
			Expect(merger.CurrentContext).To(Equal("valid-context"))
			context, ok := merger.Contexts["valid-context"]
			Expect(ok).To(BeTrue())
			Expect(context).ToNot(BeNil())
			Expect(context.AuthInfo).To(Equal(usernameTestName))
		})
	})

	Context("Update", func() {
		It("path invalid", func() {
			handler, err := NewHandler(&gokubeconfig.Merger{}, kubePath+"invalid_path/", usernameTestName)
			Expect(err).ToNot(HaveOccurred())
			err = handler.Update()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid_path/: is a directory"))
		})
	})
})

func GetHomePath() (string, error) {
	homePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if !strings.Contains(homePath, "/runai-cli/") {
		return "", fmt.Errorf("couldn't locate CLI folder")
	}
	homePath = strings.SplitAfter(homePath, "runai-cli/")[0]
	return homePath, err
}
