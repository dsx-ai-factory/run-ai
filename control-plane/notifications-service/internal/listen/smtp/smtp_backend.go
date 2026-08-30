package smtp

import (
	"bytes"
	"io"
	"net/mail"

	"github.com/emersion/go-smtp"
)

type smtpBackend struct {
	l      *SmtpListener
	bodyCh chan string
}

func NewSmtpBackend(l *SmtpListener) *smtpBackend {
	return &smtpBackend{
		l:      l,
		bodyCh: make(chan string, 100),
	}
}

func (b *smtpBackend) BodyChannel() <-chan string {
	return b.bodyCh
}

// implements the smtp.Backend interface from go-smtp
func (b *smtpBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &smtpSession{l: b.l, bodyCh: b.bodyCh}, nil
}

type smtpSession struct {
	l      *SmtpListener
	from   string
	to     []string
	bodyCh chan string
}

func (s *smtpSession) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *smtpSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *smtpSession) Data(r io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return err
	}
	msg, err := mail.ReadMessage(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	extractor := &BodyExtractor{}
	body, err := extractor.Extract(msg)
	if err != nil {
		return err
	}
	if s.bodyCh != nil {
		s.bodyCh <- string(body)
	}

	return nil
}

func (s *smtpSession) Reset() {}

func (s *smtpSession) Logout() error { return nil }
