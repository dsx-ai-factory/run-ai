/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package info

import (
	"context"
	"encoding/json"
)

type ContextType int

const (
	Id ContextType = iota
)

type Info struct {
	ctx context.Context
	Cmd string

	err error
}

type ReportInfo struct {
	Cmd   string `json:"cmd"`
	CtxID string `json:"ctx_id"`
	Error string `json:"err"`
}

func NewInfo() *Info {
	return &Info{ctx: contextWithId(context.Background(), "")}
}

func NewInfoWithID(id string) *Info {
	return &Info{ctx: contextWithId(context.Background(), id)}
}

func (i *Info) SetError(err error) {
	i.err = err
}

func (i *Info) GetError() error {
	return i.err
}

func (i *Info) GetCtx() context.Context {
	return i.ctx
}

func (i *Info) MarshalJSON() ([]byte, error) {
	return json.Marshal(ReportInfo{
		Cmd:   i.Cmd,
		CtxID: getContextId(i.ctx),
		Error: getErrorString(i.err),
	})
}

func contextWithId(ctx context.Context, id string) context.Context {
	if id != "" {
		ctx = context.WithValue(ctx, Id, id)
	} else {
		ctx = context.WithValue(ctx, Id, GenerateRandomString(UUIDLength))
	}
	return ctx
}

func getContextId(ctx context.Context) string {
	if ctx != nil {
		ret := ctx.Value(Id)
		if ret == nil {
			return ""
		}
		return ret.(string)
	}
	return ""
}

func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
