/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"encoding/json"
	"io"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

type ndjsonSink struct {
	w      io.Writer
	closer io.Closer
}

func newNDJSONSink(w io.Writer, closer io.Closer) *ndjsonSink {
	return &ndjsonSink{w: w, closer: closer}
}

func (s *ndjsonSink) Handle(e orchestrator2.Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = s.w.Write(b)
	return err
}

func (s *ndjsonSink) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}
