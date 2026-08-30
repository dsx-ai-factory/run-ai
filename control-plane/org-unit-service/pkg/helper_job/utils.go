package helper_job

import "os"

const JobModeArg = "job-mode"

func ShouldRunAsJob() bool {
	for _, arg := range os.Args {
		if arg == JobModeArg {
			return true
		}
	}
	return false
}
