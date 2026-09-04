/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/utils"
)

func TestServiceUtils_TenantErrorToResponse(t *testing.T) {
	type args struct {
		err *dao.DaoError
	}
	tests := []struct {
		name string
		su   utils.ErrorUtilser
		args args
		want *api.Error
	}{
		{
			name: "conflict",
			su:   utils.NewErrorUtils(),
			args: args{
				err: &dao.DaoError{
					Reason: dao.Conflict,
				},
			},
			want: &api.Error{
				Code:    http.StatusConflict,
				Message: "Tenant name already exists",
			},
		},
		{
			name: "not found",
			su:   utils.NewErrorUtils(),
			args: args{
				err: &dao.DaoError{
					Reason: dao.NotFound,
				},
			},
			want: &api.Error{
				Code:    http.StatusNotFound,
				Message: "Tenant not found",
			},
		},
		{
			name: "general",
			su:   utils.NewErrorUtils(),
			args: args{
				err: &dao.DaoError{
					Reason: dao.General,
				},
			},
			want: &api.Error{
				Code:    http.StatusInternalServerError,
				Message: "Tenant operation Failed",
			},
		},
		{
			name: "nil",
			su:   utils.NewErrorUtils(),
			args: args{
				err: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			su := &utils.ServiceUtils{}
			if got := su.TenantErrorToResponse(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceUtils.TenantErrorToResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceUtils_ErrorToApiError(t *testing.T) {
	type args struct {
		err        error
		statusCode int32
	}
	tests := []struct {
		name string
		su   utils.ErrorUtilser
		args args
		want *api.Error
	}{
		{
			name: "conflict",
			su:   utils.NewErrorUtils(),
			args: args{
				err:        fmt.Errorf("conflict"),
				statusCode: http.StatusConflict,
			},
			want: &api.Error{
				Code:    http.StatusConflict,
				Message: "conflict",
			},
		},
		{
			name: "not found",
			su:   utils.NewErrorUtils(),
			args: args{
				err:        fmt.Errorf("not found"),
				statusCode: http.StatusNotFound,
			},
			want: &api.Error{
				Code:    http.StatusNotFound,
				Message: "not found",
			},
		},
		{
			name: "general",
			su:   utils.NewErrorUtils(),
			args: args{
				err:        fmt.Errorf("general"),
				statusCode: http.StatusInternalServerError,
			},
			want: &api.Error{
				Code:    http.StatusInternalServerError,
				Message: "general",
			},
		},
		{
			name: "nil",
			su:   utils.NewErrorUtils(),
			args: args{
				err: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			su := &utils.ServiceUtils{}
			if got := su.ErrorToApiError(tt.args.err, tt.args.statusCode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceUtils.TenantErrorToResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
