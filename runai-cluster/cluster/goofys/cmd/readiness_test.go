// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGoofysMount(t *testing.T) {
	for _, tc := range []struct {
		name      string
		path      string
		mountInfo string
		wantReady bool
	}{
		{name: "mounted", path: "/data", mountInfo: "12 1 0:1 / /data rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "unmounted", path: "/data"},
		{name: "parent only", path: "/data", mountInfo: "12 1 0:1 / / rw - fuse.goofys goofys rw\n"},
		{name: "child only", path: "/data", mountInfo: "12 1 0:1 / /data/child rw - fuse.goofys goofys rw\n"},
		{name: "different path", path: "/data", mountInfo: "12 1 0:1 / /data-other rw - fuse.goofys goofys rw\n"},
		{name: "other FUSE type", path: "/data", mountInfo: "12 1 0:1 / /data rw - fuse.s3fs s3fs rw\n"},
		{name: "optional fields", path: "/data", mountInfo: "12 1 0:1 / /data rw shared:7 master:2 - fuse.goofys goofys rw\n", wantReady: true},
		{name: "cleaned path", path: "/other/../data/", mountInfo: "12 1 0:1 / /data rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "space", path: "/data set", mountInfo: "12 1 0:1 / /data\\040set rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "unicode whitespace", path: "/data\u00a0set", mountInfo: "12 1 0:1 / /data\u00a0set rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "carriage return", path: "/data\rset", mountInfo: "12 1 0:1 / /data\rset rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "backslash", path: "/data\\set", mountInfo: "12 1 0:1 / /data\\134set rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "tab and newline", path: "/data\tset\n", mountInfo: "12 1 0:1 / /data\\011set\\012 rw - fuse.goofys goofys rw\n", wantReady: true},
		{name: "goofys over volume", path: "/data", mountInfo: "12 20 0:1 / /data rw - fuse.goofys goofys rw\n20 1 0:2 / /data rw - ext4 disk rw\n", wantReady: true},
		{name: "goofys covered", path: "/data", mountInfo: "12 20 0:2 / /data rw - tmpfs tmpfs rw\n20 1 0:1 / /data rw - fuse.goofys goofys rw\n"},
		{name: "ancestor covered",
			path:      "/data/nested",
			mountInfo: "20 1 0:2 / /data rw - ext4 disk rw\n12 20 0:1 / /data/nested rw - fuse.goofys goofys rw\n30 20 0:3 / /data rw - tmpfs tmpfs rw\n",
		},
		{name: "mount under visible ancestor",
			path:      "/data/nested",
			mountInfo: "20 1 0:2 / /data rw - ext4 disk rw\n12 30 0:1 / /data/nested rw - fuse.goofys goofys rw\n30 20 0:3 / /data rw - tmpfs tmpfs rw\n",
			wantReady: true,
		},
		{name: "ambiguous path", path: "/data", mountInfo: "12 1 0:1 / /data rw - fuse.goofys goofys rw\n20 1 0:2 / /data rw - tmpfs tmpfs rw\n"},
		{name: "malformed record", path: "/data", mountInfo: "invalid\n12 1 0:1 / /data rw - fuse.goofys goofys rw\n"},
		{name: "truncated record", path: "/data", mountInfo: "12 1 0:1 / /data rw - fuse.goofys\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkGoofysMount(tc.mountInfo, tc.path); (err == nil) != tc.wantReady {
				t.Fatalf("mount ready = %v, want %v: %v", err == nil, tc.wantReady, err)
			}
		})
	}
}

func TestCheckReadyRejectsStaleMarker(t *testing.T) {
	dir := t.TempDir()
	ready, mountInfo := filepath.Join(dir, "ready"), filepath.Join(dir, "mountinfo")
	if err := os.WriteFile(ready, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, mounted := range []bool{true, false} {
		data := ""
		if mounted {
			data = "12 1 0:1 / /data rw - fuse.goofys goofys rw\n"
		}
		if err := os.WriteFile(mountInfo, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := goofysCommand(ctx, "--check-ready", ready, "/data")
		cmd.Env = append(cmd.Env, "GOOFYS_TEST_MOUNTINFO="+mountInfo)
		err := cmd.Run()
		if ctx.Err() != nil {
			cancel()
			t.Fatal("readiness command timed out")
		}
		cancel()
		if (err == nil) != mounted {
			t.Fatalf("probe success = %v, mount present = %v", err == nil, mounted)
		}
		if _, err := os.Stat(ready); err != nil {
			t.Fatal("test must retain the marker after unmount")
		}
	}
	if err := os.Remove(mountInfo); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := goofysCommand(ctx, "--check-ready", ready, "/data")
	cmd.Env = append(cmd.Env, "GOOFYS_TEST_MOUNTINFO="+mountInfo)
	if err := cmd.Run(); err == nil {
		t.Fatal("probe succeeded when the current mounts could not be read")
	}
}
