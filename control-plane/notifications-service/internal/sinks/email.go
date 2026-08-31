/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"bytes"
	"fmt"
	"html/template"
	"net/mail"
	"path"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/smtp"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

const (
	HeadersDelimiter = "\r\n"
	EmailSinkType    = "email"
)

type EmailSinkConfig struct {
	TemplatesFolder     string          `mapstructure:"templates_folder,omitempty" json:"templates_folder,omitempty"`
	TemplateFile        string          `mapstructure:"template_file,omitempty" json:"template_file,omitempty"`
	LogoUrl             string          `mapstructure:"logo_url" json:"logo_url"`
	FromDisplayName     string          `mapstructure:"from_display_name" json:"from_display_name"`
	From                string          `mapstructure:"from" json:"from"`
	Subject             string          `mapstructure:"subject" json:"subject"`
	DirectNotifications bool            `mapstructure:"direct_notifications" json:"direct_notifications"`
	UseAttachments      bool            `mapstructure:"use_attachments" json:"use_attachments"`
	Recipients          []string        `mapstructure:"recipients" json:"recipients"`
	Smtp                smtp.SmtpConfig `mapstructure:",squash"`
}

func NewEmailSinkConfig(viper *viper.Viper, id string) (*EmailSinkConfig, error) {
	setEmailDefaults(viper)
	cfg := &EmailSinkConfig{}
	if err := viper.Unmarshal(cfg, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed to unmarshal configuration of email sink %q", id)
		return nil, err
	}
	return cfg, nil
}

func setEmailDefaults(v *viper.Viper) {
	v.SetDefault("templates_folder", "sinks_templates/email")
	v.SetDefault("template_file", "generic_notification.tpl")
	v.SetDefault("logo_url", "https://s3.amazonaws.com/www.run.ai/OS_NVIDIA_Runai_black.svg")
	v.SetDefault("from_display_name", "NVIDIA Run:ai")
	v.SetDefault("from", "nvidia-runai@nvidia.com")
	v.SetDefault("subject", "NVIDIA Run:ai notification - {{ .category }} {{ .reason }}")
	v.SetDefault("direct_notifications", false)
	v.SetDefault("recipients", []string{})

	v.SetDefault("smtp_port", 587)
	v.SetDefault("smtp_tls_enabled", true)
	v.SetDefault("auth_type", string(smtp.AuthLogin))
}

type BodyEncoder interface {
	Encode(body string) string
}

func NewEmailSink(config *EmailSinkConfig) *EmailSink {
	return &EmailSink{
		config:     config,
		smtpClient: smtp.NewSmtpClient(&config.Smtp),
		encoder:    &smtp.BodyEncoder{},
	}
}

type EmailSink struct {
	config     *EmailSinkConfig
	smtpClient smtp.SmtpClient
	encoder    BodyEncoder
}

func (es *EmailSink) Type() string {
	return EmailSinkType
}

func (es *EmailSink) Verify() error {
	return es.smtpClient.Verify(es.config.From, es.config.Recipients)
}

func (es *EmailSink) Sink(event *events.Event) error {
	log.Debug().Strs("destinations", es.config.Recipients).Msgf("Preparing to send email notification for event %s", event.UID)

	recipients := es.getRecipients(event)
	if len(recipients) == 0 {
		log.Debug().Msgf("No recipients found for event %s", event.UID)
		return nil
	}
	log.Debug().Strs("destinations", recipients).Msgf("Sending email notification for event: %s", event.UID)

	message, err := es.buildMessage(recipients, event)
	if err != nil {
		return errors.Wrap(err, "failed to build email message")
	}

	err = es.smtpClient.SendMail(es.config.From, recipients, message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send email")
		return err
	}

	log.Debug().Msgf("Email on event %s sent successfully", event.UID)
	return nil
}

func (es *EmailSink) buildMessage(recipients []string, event *events.Event) ([]byte, error) {
	headers, err := es.getHeaders(recipients, event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse email template headers - form,subject, etc...")
	}

	body, err := es.getEmailData(event)
	if err != nil {
		return nil, err
	}
	body = es.encoder.Encode(body)

	message := []byte(headers + body)
	log.Debug().Msgf("Email message: %s", message)

	return message, err
}

func (es *EmailSink) getSubject(event *events.Event) (string, error) {
	eventProperties := event.Properties
	if value, ok := eventProperties["subject"]; ok {
		subject, ok := value.(string)
		if ok {
			return sanitizeHeaderValue(subject), nil
		}
		log.Warn().Msgf("failed to parse subject property to string, event: %s", event.UID)
	}

	subjectTemplate, err := template.New("subject").Parse(es.config.Subject)
	if err != nil {
		return "", err
	}

	var parsedSubject bytes.Buffer
	err = subjectTemplate.Execute(&parsedSubject, eventProperties)
	if err != nil {
		return "", err
	}
	return sanitizeHeaderValue(parsedSubject.String()), nil
}

func (es *EmailSink) getHeaders(recipients []string, event *events.Event) (string, error) {
	fromHeader := fmt.Sprintf("From: %s  <%s>%s",
		sanitizeHeaderValue(es.config.FromDisplayName), sanitizeHeaderValue(es.config.From), HeadersDelimiter)
	toHeader := fmt.Sprintf("To: %s%s", sanitizeHeaderValue(joinRecipients(recipients)), HeadersDelimiter)
	subjectText, err := es.getSubject(event)
	if err != nil {
		return "", err
	}
	subjectHeader := fmt.Sprintf("Subject: %s %s", subjectText, HeadersDelimiter)
	return fromHeader + toHeader + subjectHeader, nil
}

var headerSanitizer = strings.NewReplacer("\r", " ", "\n", " ", "\x00", "")

func sanitizeHeaderValue(value string) string {
	return headerSanitizer.Replace(value)
}

func (es *EmailSink) getEmailData(event *events.Event) (string, error) {
	parsedTemplate, err := es.parseTemplate()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = parsedTemplate.Execute(&buf, es.getTemplateData(event))
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type TemplateData struct {
	EventProperties map[string]interface{}
	MetaProperties  map[string]interface{}
	Description     template.HTML
	CurrentYear     int
}

func (es *EmailSink) getTemplateData(event *events.Event) *TemplateData {
	metaProperties := map[string]interface{}{
		"logoUrl":         es.config.LogoUrl,
		"from":            es.config.From,
		"fromDisplayName": es.config.FromDisplayName,
		"useAttachments":  es.config.UseAttachments,
	}

	templateData := &TemplateData{
		EventProperties: event.Properties,
		MetaProperties:  metaProperties,
		CurrentYear:     time.Now().Year(),
	}

	var err error
	if templateData.MetaProperties["subject"], err = es.getSubject(event); err != nil {
		log.Warn().Err(err).Msg("failed to parse email subject, will use a default subject")
		templateData.MetaProperties["subject"] = "Run:ai notification"
	}

	if description, ok := event.Properties["description"].(string); ok {
		safeDescription := sanitizeDescriptionHTML(description)
		//#nosec G203: description is sanitized, so this lint issue is a false positive.
		// G203 is a low-confidence rule and cannot guarantee an actual security issue.
		templateData.Description = template.HTML(safeDescription)
	}

	return templateData
}

func (es *EmailSink) parseTemplate() (*template.Template, error) {
	if es.config.TemplatesFolder == "" {
		return nil, fmt.Errorf("missing TemplatesFolder for sink config")
	}
	if es.config.TemplateFile == "" {
		return nil, fmt.Errorf("missing TemplateFile for sink config")
	}
	return template.ParseFiles(path.Join(es.config.TemplatesFolder, es.config.TemplateFile))
}

func (es *EmailSink) getRecipients(event *events.Event) []string {
	c := es.config
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
		return c.Recipients // already contained
	}
	return append(c.Recipients, user)
}

func joinRecipients(recipients []string) string {
	to := []string{}
	for _, recipient := range recipients {
		to = append(to, fmt.Sprintf("%s <%s>", recipient, recipient))
	}
	return strings.Join(to, ",")
}

func sanitizeDescriptionHTML(html string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("style", "class").Globally()

	return policy.Sanitize(html)
}
