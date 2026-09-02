/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package args

import (
	"fmt"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/spf13/cobra"
)

const (
	DashStr = "-"

	GroupTogetherBySingleQuote = "'%s'"
	EmptyStr                   = ""
	SpaceStr                   = " "

	helpMsg                = "Please check --help examples for how to user args and commands"
	errMsgMissingDelimiter = "argument found but missing delimiter '--', args: %s\n" + helpMsg
	errMsgArgsPrior        = "found args prior to args marker (--): %s\n" + helpMsg
)

func GetNameFromArgs(args *[]string) string {
	if len(*args) > 0 {
		return (*args)[0]
	}
	return ""
}

func ValidateArgsAndExtractName(cmd *cobra.Command, args *[]string) (string, error) {
	var nameFound string
	if len(*args) == 0 {
		return nameFound, nil
	}
	osArgs := os.Args
	for i := 0; i < len(osArgs)-1; i++ {
		if osArgs[i] == cmd.Name() || contains(cmd.Aliases, osArgs[i]) {
			nameCheck := osArgs[i+1]
			if !strings.HasPrefix(nameCheck, DashStr) {
				*args = (*args)[1:]
				nameFound = nameCheck
			}
			break
		}
	}

	// ToDo: should be deleted after remove all the deprecated code
	// This is a temporary solution to support the deprecated submit command without validation
	if cmd.Short == consts.SubmitShortMsg {
		return nameFound, nil
	}

	return nameFound, validateArgsStructure(args, &osArgs)
}

func validateArgsStructure(args, osArgs *[]string) error {
	if len(*args) == 0 {
		return nil
	}

	var (
		argsMarker int
	)

	for argsMarker = range *osArgs {
		if (*osArgs)[argsMarker] == "--" {
			break
		}
	}

	if argsMarker+1 == len(*osArgs) {
		return fmt.Errorf(errMsgMissingDelimiter, *args)
	}

	argsAfterMarker := len(*osArgs) - (argsMarker + 1)
	if argsAfterMarker == len(*args) {
		return nil
	}
	dislocatedArgs := (*args)[0:(len(*args) - argsAfterMarker)]
	return fmt.Errorf(errMsgArgsPrior, dislocatedArgs)
}

func GetExecuteString(args *[]string) (string, bool) {
	if args == nil {
		return EmptyStr, false
	}
	for i := range *args {
		if strings.Contains((*args)[i], "'") {
			(*args)[i] = strings.ReplaceAll((*args)[i], "'", "\"")
		}
		if strings.Contains((*args)[i], SpaceStr) {
			(*args)[i] = fmt.Sprintf(GroupTogetherBySingleQuote, (*args)[i])
		}
	}
	if len(*args) > 0 {
		return strings.Join(*args, SpaceStr), true
	}
	return "", false
}

func BothAreSet(hasName bool, hasPrefix bool) error {
	if hasName && hasPrefix {
		return fmt.Errorf("both prefix and name are set, only one can be used")
	}
	return nil
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
