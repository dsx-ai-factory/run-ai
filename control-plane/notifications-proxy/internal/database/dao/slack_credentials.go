/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"errors"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/database/models"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const slackCredentialId = 1 // we allow at most one row in the table, fixed id lets us safely upsert via ON CONFLICT

type SlackCredentials struct{}

func NewSlackCredentials() *SlackCredentials {
	return &SlackCredentials{}
}

func (s *SlackCredentials) Save(ctx context.Context, c *slack.SlackCredentials) error {
	slackCredential := &models.SlackCredential{
		Id:           slackCredentialId,
		AppId:        c.AppId,
		ClientId:     c.ClientId,
		ClientSecret: c.ClientSecret,
	}

	_, err := runaidb.DB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(slackCredential)
	return err
}

func (s *SlackCredentials) Get(ctx context.Context) (*slack.SlackCredentials, error) {
	var slackCredential models.SlackCredential
	err := runaidb.DB().WithContext(ctx).First(&slackCredential)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, slack.ErrMissingAppCredentials
		}
		return nil, err
	}

	return &slack.SlackCredentials{
		AppId:        slackCredential.AppId,
		ClientId:     slackCredential.ClientId,
		ClientSecret: slackCredential.ClientSecret,
	}, nil
}

func (s *SlackCredentials) Delete(ctx context.Context) error {
	_, err := runaidb.DB().WithContext(ctx).
		Where(&models.SlackCredential{Id: slackCredentialId}).
		Delete(&models.SlackCredential{})
	return err
}
