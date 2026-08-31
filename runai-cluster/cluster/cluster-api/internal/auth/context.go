/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"

	"github.com/rs/zerolog/log"
)

type contextKey int

const (
	usernameKey contextKey = iota
	subjectKey
	originalSubjectKey
	uidKey
	gidKey
	groupsKey
	authLevelKey
	canSeeAllProj
	rawTokenKey
)

//	  inject authentication details into the context
//			- auth client
//			- username
//			- subject
//			- UID
//			- GID
//			- Groups
func ComposeAuthContext(ctx context.Context, authInfo *Info) context.Context {
	log.Debug().Msgf("setting ctx for user: %v", authInfo)

	ctx = context.WithValue(ctx, usernameKey, authInfo.Username)
	ctx = context.WithValue(ctx, subjectKey, authInfo.Subject)
	ctx = context.WithValue(ctx, originalSubjectKey, authInfo.OriginalSubject)
	ctx = context.WithValue(ctx, groupsKey, authInfo.Groups)
	ctx = context.WithValue(ctx, uidKey, authInfo.Uid)
	ctx = context.WithValue(ctx, gidKey, authInfo.Gid)
	ctx = context.WithValue(ctx, canSeeAllProj, authInfo.CanSeeAllProjects)
	ctx = context.WithValue(ctx, rawTokenKey, authInfo.RawToken)

	return ctx
}

func FromContext(ctx context.Context) *Info {
	uid, gid := UidGidFromContext(ctx)

	return &Info{
		Username:          UsernameFromContext(ctx),
		Subject:           SubjectFromContext(ctx),
		OriginalSubject:   OriginalSubjectFromContext(ctx),
		Groups:            GroupsFromContext(ctx),
		Uid:               uid,
		Gid:               gid,
		CanSeeAllProjects: CanSeeAllProjFromContext(ctx),
		RawToken:          RawTokenFromContext(ctx),
	}
}

func UsernameFromContext(ctx context.Context) string {
	value := ctx.Value(usernameKey)
	switch value.(type) {
	case string:
		return ctx.Value(usernameKey).(string)

	default:
		return ""
	}
}

func SubjectFromContext(ctx context.Context) string {
	value := ctx.Value(subjectKey)
	switch v := value.(type) {
	case string:
		return v

	default:
		return ""
	}
}

func OriginalSubjectFromContext(ctx context.Context) string {
	value := ctx.Value(originalSubjectKey)
	switch v := value.(type) {
	case string:
		return v

	default:
		return ""
	}
}

func GroupsFromContext(ctx context.Context) []string {
	value := ctx.Value(groupsKey)
	switch v := value.(type) {
	case []string:
		return v

	default:
		return nil
	}
}

func CanSeeAllProjFromContext(ctx context.Context) bool {
	value := ctx.Value(canSeeAllProj)
	switch v := value.(type) {
	case bool:
		return v

	default:
		return false
	}
}

func UidGidFromContext(ctx context.Context) (*int64, *int64) {
	var uid, gid *int64

	switch ctx.Value(uidKey).(type) {
	case *int64:
		uid = ctx.Value(uidKey).(*int64)
	}

	switch ctx.Value(gidKey).(type) {
	case *int64:
		gid = ctx.Value(gidKey).(*int64)
	}

	return uid, gid
}

func RawTokenFromContext(ctx context.Context) string {
	value := ctx.Value(rawTokenKey)
	switch v := value.(type) {
	case string:
		return v

	default:
		return ""
	}
}
