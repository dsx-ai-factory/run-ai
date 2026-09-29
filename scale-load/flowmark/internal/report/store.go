/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("object not found")

type Store interface {
	Put(ctx context.Context, name string, data []byte) error
	Get(ctx context.Context, name string) ([]byte, error)
	String() string
}

type Local struct {
	dir string
}

func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("creating output directory %q: %w", dir, err)
	}
	return &Local{dir: dir}, nil
}

func (l *Local) Put(_ context.Context, name string, data []byte) error {
	target := filepath.Join(l.dir, filepath.Base(name))
	temp, err := os.CreateTemp(l.dir, filepath.Base(name)+".*.tmp")
	if err != nil {
		return err
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		_ = os.Remove(temp.Name())
		return err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(temp.Name())
		return err
	}
	if err := os.Rename(temp.Name(), target); err != nil {
		_ = os.Remove(temp.Name())
		return err
	}
	return nil
}

func (l *Local) Get(_ context.Context, name string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(l.dir, filepath.Base(name)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

func (l *Local) String() string {
	return "directory " + l.dir
}
