// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/kahing/goofys/api/common"
)

const lazyUnmountMessage = "lazy unmount completed\n"

func TestSIGTERMExitsAfterLazyUnmountLinux(t *testing.T) {
	// 1. Run the scenario in a child process because the handler calls os.Exit.
	cmd := exec.Command(os.Args[0], "-test.run=^TestBusyMountProcessLinux$", "-test.timeout=5s")
	cmd.Env = append(os.Environ(), "GOOFYS_TEST_SHUTDOWN=1")
	output, err := cmd.CombinedOutput()

	// 2. Check that the process exited successfully.
	if err != nil {
		t.Fatalf("process exited without lazily unmounting: %s", output)
	}

	// 3. Check that lazy unmount happened before exit.
	if !bytes.Contains(output, []byte(lazyUnmountMessage)) {
		t.Fatalf("process exited without lazily unmounting: %s", output)
	}
}

// TestBusyMountProcessLinux is the child process for the test above.
func TestBusyMountProcessLinux(t *testing.T) {
	if os.Getenv("GOOFYS_TEST_SHUTDOWN") != "1" {
		return
	}

	// 1. Simulate a busy mount: normal unmount fails, lazy unmount succeeds.
	unmount = func(_ string, flags int) error {
		switch flags {
		case 0:
			return syscall.EBUSY
		case syscall.MNT_DETACH:
			_, err := os.Stdout.WriteString(lazyUnmountMessage)
			return err
		}
		return syscall.EINVAL
	}

	// 2. Register the handler and send SIGTERM.
	registerSIGINTHandler(nil, &common.FlagStorage{MountPoint: "/mnt/s3"})
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}

	// 3. Simulate the FUSE server waiting. The handler must exit the process.
	select {}
}
