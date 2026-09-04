/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cliflags

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagGroupName string

type FlagGroup struct {
	Name    FlagGroupName
	FlagSet *pflag.FlagSet
}

type FlagsByGroups struct {
	cmd           *cobra.Command
	groupsByOrder *[]FlagGroup
}

func NewFlagsByGroups(cmd *cobra.Command) FlagsByGroups {
	return FlagsByGroups{
		cmd:           cmd,
		groupsByOrder: &([]FlagGroup{}),
	}
}

func (fg *FlagGroup) usage() string {
	return fmt.Sprint(Bold(fg.Name), "\n", fg.FlagSet.FlagUsagesWrapped(1))
}

func NewFlagGroup(name FlagGroupName) FlagGroup {
	return FlagGroup{
		Name:    name,
		FlagSet: pflag.NewFlagSet(string(name), 0),
	}
}

func (fbg *FlagsByGroups) GetOrAddFlagSet(groupName FlagGroupName) (fs *pflag.FlagSet) {
	for _, item := range *fbg.groupsByOrder {
		if item.Name == groupName {
			fs = item.FlagSet
		}
	}
	if fs == nil {
		newFlagGroup := NewFlagGroup(groupName)
		*fbg.groupsByOrder = append(*fbg.groupsByOrder, newFlagGroup)
		fs = newFlagGroup.FlagSet
	}
	return fs
}

func (fbg *FlagsByGroups) UpdateFlagsByGroupsToCmd(completionFnsToRegister map[string]completion.CompletionFn) {
	updateFlagsByGroupsToCmd(fbg.cmd, completionFnsToRegister, *fbg.groupsByOrder...)
}

func updateFlagsByGroupsToCmd(cmd *cobra.Command, completionFnsToRegister map[string]completion.CompletionFn, fgs ...FlagGroup) {
	for _, fg := range fgs {
		cmd.Flags().AddFlagSet(fg.FlagSet)
	}

	for flagName, fn := range completionFnsToRegister {
		_ = cmd.RegisterFlagCompletionFunc(flagName, fn)
	}

	usage := flagsGroupUsage(fgs...)

	cmd.SetUsageTemplate(
		`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
` + usage +
			`Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{"\n"}}`)
}

func flagsGroupUsage(fgs ...FlagGroup) string {
	usage := []string{}
	for _, fg := range fgs {
		usage = append(usage, "\n")
		usage = append(usage, strings.TrimSpace(fg.usage()))
		usage = append(usage, "\n")
	}
	usageStr := strings.Join(usage, "")
	return usageStr + "\n"
}

func Bold(text interface{}) string {
	return fmt.Sprint("\033[1m", text, "\033[0m")
}
