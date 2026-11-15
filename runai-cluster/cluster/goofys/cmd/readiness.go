// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var readMountInfo = func() ([]byte, error) { return os.ReadFile("/proc/self/mountinfo") }

type readinessMount struct {
	parent string
	path   string
	fsType string
}

func checkGoofysMount(mountInfo, path string) error {
	// mountinfo escapes these characters in pathnames.
	path = strings.NewReplacer("\\", `\134`, " ", `\040`, "\t", `\011`, "\n", `\012`).Replace(filepath.Clean(path))
	mounts := make(map[string]readinessMount)
	for line := range strings.Lines(mountInfo) {
		fields := strings.Split(strings.TrimSuffix(line, "\n"), " ")
		separator := slices.Index(fields, "-")
		if separator < 6 || len(fields) < separator+4 {
			return fmt.Errorf("invalid mountinfo record")
		}
		mounts[fields[0]] = readinessMount{parent: fields[1], path: fields[4], fsType: fields[separator+1]}
	}
	covered := make(map[string]bool)
	for id, mount := range mounts {
		if parent, ok := mounts[mount.parent]; ok && id != mount.parent && parent.path == mount.path {
			covered[mount.parent] = true
		}
	}
	var fsType string
	for id, mount := range mounts {
		if mount.path != path || !mountVisible(id, mounts, covered) {
			continue
		}
		if fsType != "" {
			return fmt.Errorf("ambiguous mounts at %s", path)
		}
		fsType = mount.fsType
	}
	if fsType != "fuse.goofys" {
		return fmt.Errorf("no current fuse.goofys mount at %s", path)
	}
	return nil
}

func mountVisible(id string, mounts map[string]readinessMount, covered map[string]bool) bool {
	if covered[id] {
		return false
	}
	for {
		mount := mounts[id]
		parent, ok := mounts[mount.parent]
		if !ok || id == mount.parent {
			return true
		}
		// An overmounted ancestor also hides mounts beneath it.
		if parent.path != mount.path && covered[mount.parent] {
			return false
		}
		id = mount.parent
	}
}
