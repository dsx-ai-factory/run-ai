/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package multi_error

import "sync"

type MultiError struct {
	errors []error
	mux    sync.RWMutex
}

func New() *MultiError {
	return &MultiError{}
}

func (m *MultiError) Error() string {
	m.mux.RLock()
	defer m.mux.RUnlock()

	errMsg := ""

	for _, err := range m.errors {
		errMsg += err.Error() + "\n"
	}

	return errMsg
}

func (m *MultiError) Len() int {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return len(m.errors)
}

func (m *MultiError) Empty() bool {
	return m.Len() == 0
}

func (m *MultiError) Append(err error) {
	if err == nil {
		return
	}
	m.mux.Lock()
	defer m.mux.Unlock()

	m.errors = append(m.errors, err)
}

func (m *MultiError) AppendAll(errs []error) {
	if errs == nil {
		return
	}
	m.mux.Lock()
	defer m.mux.Unlock()

	m.errors = append(m.errors, errs...)
}

func (m *MultiError) ErrorOrNil() error {
	if m.Empty() {
		return nil
	}
	return m
}

func (m *MultiError) Errors() []error {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return append([]error{}, m.errors...)
}
