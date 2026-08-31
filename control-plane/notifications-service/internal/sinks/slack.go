/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"bytes"
	"encoding/json"
	"net/mail"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/maypok86/otter/v2"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/slack"
	slack_go "github.com/slack-go/slack"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

const (
	DefaultTemplateFolder    = "sinks_templates/slack"
	DefaultTemplateFile      = "generic_notification.tpl"
	SlackSinkType            = "slack"
	SlackChannelRegexPattern = "^[C][A-Z0-9]+$"
	SlackUserRegexPattern    = "^[UW][A-Z0-9]+$"

	slackUserByEmailCacheTTL = 1 * time.Hour
)

var (
	slackChannelRegex = regexp.MustCompile(SlackChannelRegexPattern)
	slackUserRegex    = regexp.MustCompile(SlackUserRegexPattern)

	slackUserByEmailCache = otter.Must(&otter.Options[string, string]{
		ExpiryCalculator: otter.ExpiryWriting[string, string](slackUserByEmailCacheTTL),
	})
)

type Message struct {
	Blocks []slack_go.Block `json:"blocks"`
}

type SlackClienter interface {
	PostMessage(channel string, messageBlocks []slack_go.Block) error
	PostDMMessage(userIds []string, messageBlocks []slack_go.Block) error
	GetUserByEmail(email string) (*slack_go.User, error)
	AuthTest() (*slack_go.AuthTestResponse, error)
}

type SlackSinkConfig struct {
	Url                 string   `mapstructure:"url,omitempty" json:"url,omitempty"`
	ApiToken            string   `mapstructure:"api_token" json:"api_token" secret:"true"`
	Recipients          []string `mapstructure:"recipients" json:"recipients"`
	DirectNotifications bool     `mapstructure:"direct_notifications" json:"direct_notifications"`
	TemplatesFolder     string   `mapstructure:"templates_folder,omitempty" json:"templates_folder,omitempty"`
	TemplateFile        string   `mapstructure:"template_file,omitempty" json:"template_file,omitempty"`
}

func NewSlackSinkConfig(vip *viper.Viper, id string) (*SlackSinkConfig, error) {
	setSlackDefaults(vip)
	cfg := &SlackSinkConfig{}
	if err := vip.Unmarshal(cfg, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed to unmarshal configuration of slack sink %q", id)
		return nil, err
	}
	return cfg, nil
}

func setSlackDefaults(v *viper.Viper) {
	v.SetDefault("url", slack_go.APIURL)
	v.SetDefault("recipients", []string{})
	v.SetDefault("templates_folder", DefaultTemplateFolder)
	v.SetDefault("template_file", DefaultTemplateFile)
}

func NewSlackSink(config *SlackSinkConfig) *SlackSink {
	return &SlackSink{
		config:      config,
		slackClient: slack.NewSlackClient(config.Url, config.ApiToken),
	}
}

type SlackSink struct {
	config      *SlackSinkConfig
	slackClient SlackClienter
}

func (ss *SlackSink) Type() string {
	return SlackSinkType
}

func (ss *SlackSink) Verify() error {
	_, err := ss.slackClient.AuthTest()
	return err
}

func (ss *SlackSink) Sink(event *events.Event) error {
	log.Debug().Strs("destinations", ss.config.Recipients).Msgf("Preparing to send slack notification for event %s", event.UID)

	recipients := ss.getRecipients(event)
	if len(recipients) == 0 {
		log.Debug().Msgf("No recipients found for event %s", event.UID)
		return nil
	}
	log.Debug().Strs("destinations", recipients).Msgf("Sending Slack notification for event: %s", event.UID)

	blocks, err := ss.generateSlackBlocks(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate slack blocks")
		return err
	}

	ss.sendMessage(recipients, blocks)

	return nil
}

func (ss *SlackSink) sendMessage(recipients []string, blocks []slack_go.Block) {
	for _, recipient := range recipients {
		switch {
		case isSlackChannelId(recipient):
			ss.sendMessageToChannel(recipient, blocks)
		case isSlackUserId(recipient):
			ss.sendMessageToUser(recipient, blocks)
		case isEmail(recipient):
			ss.sendMessageToEmail(recipient, blocks)
		default:
			log.Error().Msgf("Invalid Slack recipient: %s", recipient)
		}
	}
}

func isSlackChannelId(input string) bool {
	return slackChannelRegex.MatchString(input)
}

func isSlackUserId(input string) bool {
	return slackUserRegex.MatchString(input)
}

func isEmail(input string) bool {
	_, err := mail.ParseAddress(input)
	return err == nil
}

func (ss *SlackSink) sendMessageToChannel(channelId string, blocks []slack_go.Block) {
	if err := ss.slackClient.PostMessage(channelId, blocks); err != nil {
		log.Error().Err(err).Msgf("Failed to send Slack message to channel: %s", channelId)
	}
}

func (ss *SlackSink) sendMessageToUser(userId string, blocks []slack_go.Block) {
	if err := ss.slackClient.PostDMMessage([]string{userId}, blocks); err != nil {
		log.Error().Err(err).Msgf("Failed to send DM to user with id: %s", userId)
	}
}

func (ss *SlackSink) sendMessageToEmail(email string, blocks []slack_go.Block) {
	if userId, ok := slackUserByEmailCache.GetIfPresent(email); ok {
		ss.sendMessageToUser(userId, blocks)
		return
	}

	user, err := ss.slackClient.GetUserByEmail(email)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to resolve Slack user for email: %s", email)
		return
	}

	slackUserByEmailCache.Set(email, user.ID)
	ss.sendMessageToUser(user.ID, blocks)
}

func (ss *SlackSink) generateSlackBlocks(event *events.Event) ([]slack_go.Block, error) {
	parsedTemplate, err := ss.parseTemplate()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = parsedTemplate.Execute(&buf, ss.getTemplateData(event))
	if err != nil {
		return nil, err
	}

	var sectionBlocks []slack_go.SectionBlock
	err = json.Unmarshal(buf.Bytes(), &sectionBlocks)
	if err != nil {
		return nil, err
	}

	return ss.convertToSlackBlocks(sectionBlocks), nil
}

func (ss *SlackSink) parseTemplate() (*template.Template, error) {
	return template.ParseFiles(path.Join(ss.config.TemplatesFolder, ss.config.TemplateFile))
}

func (ss *SlackSink) getTemplateData(event *events.Event) map[string]interface{} {
	if desc, ok := event.Properties["description"].(string); ok {
		event.Properties["description"] = strings.TrimSpace(convertDescriptionHTMLToMarkdown(desc))
	}

	return map[string]interface{}{
		"EventProperties": event.Properties,
	}
}

func (ss *SlackSink) convertToSlackBlocks(sectionBlocks []slack_go.SectionBlock) []slack_go.Block {
	var blocks []slack_go.Block
	for _, section := range sectionBlocks {
		blocks = append(blocks, section)
	}
	return blocks
}

func (ss *SlackSink) getRecipients(event *events.Event) []string {
	c := ss.config
	if !c.DirectNotifications {
		return c.Recipients
	}

	user, ok := event.Properties["user_id"].(string)
	if !ok || user == "" {
		log.Debug().Msg("event is missing a string user field, not sending direct notification")
		return c.Recipients
	}

	if _, err := mail.ParseAddress(user); err != nil {
		log.Debug().Msgf("user %s is not a valid email address, not sending direct notification", user)
		return c.Recipients
	}

	log.Debug().Msgf("user %s is a valid email address, sending direct notification", user)
	if slices.Contains(c.Recipients, user) {
		return c.Recipients
	}
	return append(c.Recipients, user)
}

func convertDescriptionHTMLToMarkdown(html string) string {
	markdown, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert HTML description to markdown. Using original HTML description.")
		return html
	}
	return markdown
}
