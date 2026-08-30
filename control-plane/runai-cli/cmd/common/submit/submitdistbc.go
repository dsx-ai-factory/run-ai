// Package submit - delete this file once deprecated submit-dist is EOL
package submit

import (
	"os"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/spf13/cobra"
)

func PreDistBc(cmd *cobra.Command, args []string, flags spec.BCFlags) error {
	if flags.WorkloadName != "" {
		var temp []string
		// if name is not empty, add it to the args list after the framework
		args = append([]string{flags.WorkloadName}, args...)
		for i := range os.Args {
			temp = append(temp, os.Args[i])
			if os.Args[i] == cmd.Use {
				temp = append(temp, flags.WorkloadName)
			}
		}
		os.Args = temp
	}

	return Pre(cmd, args)
}
