package smtp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

const ListenerType = "smtp_listener"

type SmtpBodyProvider interface {
	BodyChannel() <-chan string
}

type SmtpListener struct {
	id           string
	config       *SmtpListenerConfig
	server       *smtp.Server
	bodyProvider SmtpBodyProvider
	eventCh      chan *events.Event
}

func NewSmtpListener(id string, eventCh chan *events.Event, config *SmtpListenerConfig) *SmtpListener {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	backend := NewSmtpBackend(nil) // Listener not yet constructed, can be set later if needed

	server := smtp.NewServer(backend)
	server.Addr = addr
	server.AllowInsecureAuth = true
	server.Domain = config.Host

	return &SmtpListener{
		id:           id,
		config:       config,
		bodyProvider: backend,
		server:       server,
		eventCh:      eventCh,
	}
}

func (l *SmtpListener) Id() string {
	return l.id
}

func (l *SmtpListener) Listen(ctx context.Context) {
	go func() {
		if err := l.server.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("SMTP server stopped")
		}
	}()

	go func() {
		<-ctx.Done()
		l.server.Close()
	}()

	l.ListenAndParseBodies()
}

func (l *SmtpListener) ListenAndParseBodies() {
	for body := range l.bodyProvider.BodyChannel() {
		event := l.parseBodyToEvent(body)
		if event != nil {
			l.eventCh <- event
		}
	}
}

func (l *SmtpListener) parseBodyToEvent(body string) *events.Event {
	re := regexp.MustCompile(`(?s)-----BEGIN NOTIFICATION-----\s*(.*?)\s*-----END NOTIFICATION-----`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		log.Error().Msg("No notification block found in mail body")
		return nil
	}

	notificationContent := matches[1]

	// remove soft line breaks according to smtp rfc, =\r\n - wrap long lines https://www.rfc-editor.org/rfc/rfc2822.html#section-2.1.1
	// also removes unescaped \r\n from inside the json fields otherwise json unmarshal will fail
	softLineBreaks := regexp.MustCompile(`=?\r?\n`)
	notificationContent = softLineBreaks.ReplaceAllString(notificationContent, "")

	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(notificationContent), &properties); err != nil {
		log.Error().Err(err).Msgf("Failed to parse notification block as JSON: %v", err)
		return nil
	}
	log.Debug().Msgf("Parsed notification block to properties: %+v", properties)
	return &events.Event{
		RawEvent:   notificationContent,
		UID:        events.UID(uuid.New().String()),
		ListenerId: l.id,
		Properties: properties,
	}
}
