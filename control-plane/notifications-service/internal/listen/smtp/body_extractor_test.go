package smtp_test

import (
	"net/mail"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen/smtp"
	"github.com/stretchr/testify/require"
)

func TestExtract_PlainNoEncoding(t *testing.T) {
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type": []string{"text/plain; charset=UTF-8"},
		},
		Body: strings.NewReader("Hello world!"),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	require.NoError(t, err)
	require.Equal(t, "Hello world!", out)
}

func TestExtract_PlainQuotedPrintable(t *testing.T) {
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type":              []string{"text/plain; charset=UTF-8"},
			"Content-Transfer-Encoding": []string{"quoted-printable"},
		},
		Body: strings.NewReader("Hello=3Dworld!"),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	require.NoError(t, err)
	require.Equal(t, "Hello=world!", out)
}

func TestExtract_MultipartAlternative(t *testing.T) {
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type": []string{`multipart/alternative; boundary="BOUNDARY"`},
		},
		Body: strings.NewReader(
			"--BOUNDARY\nContent-Type: text/plain; charset=UTF-8\r\n\r\nPlain part.\n--BOUNDARY\nContent-Type: text/html; charset=UTF-8\r\n\r\n<html>HTML part.</html>\n--BOUNDARY--",
		),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	want := "Plain part.\r\n<html>HTML part.</html>\r\n"
	require.NoError(t, err)
	require.Equal(t, want, out)
}

func TestExtract_MultipartMixedWithEncoding(t *testing.T) {
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type": []string{`multipart/mixed; boundary="MIXED"`},
		},
		Body: strings.NewReader(
			"--MIXED\nContent-Type: text/plain; charset=UTF-8\nContent-Transfer-Encoding: quoted-printable\r\n\r\nHello=3Dworld!\n--MIXED\nContent-Type: text/plain\r\n\nAnother part!\n--MIXED--",
		),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	want := "Hello=world!\r\nAnother part!\r\n"
	require.NoError(t, err)
	require.Equal(t, want, out)
}

func TestExtract_MissingContentType(t *testing.T) {
	msg := &mail.Message{
		Header: mail.Header{},
		Body:   strings.NewReader("Body without headers."),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	require.Error(t, err)
	require.Empty(t, out)
}

func TestExtract_EmptyBody(t *testing.T) {
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type": []string{"text/plain; charset=UTF-8"},
		},
		Body: strings.NewReader(""),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestExtract_HtmlEntitiesUnescaped(t *testing.T) {
	htmlBody := `{"link": "https://example.com?a=1&amp;b=2&amp;c=3", "text": "&lt;hello&gt; &quot;world&quot;"}`
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type": []string{"text/html; charset=UTF-8"},
		},
		Body: strings.NewReader(htmlBody),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	require.NoError(t, err)
	expected := `{"link": "https://example.com?a=1&b=2&c=3", "text": "<hello> "world""}`
	require.Equal(t, expected, out)
}

func TestExtract_PlainTextUnchanged(t *testing.T) {
	plainBody := `{"link": "https://example.com?a=1", "text": "hello world"}`
	msg := &mail.Message{
		Header: mail.Header{
			"Content-Type": []string{"text/plain; charset=UTF-8"},
		},
		Body: strings.NewReader(plainBody),
	}
	ext := smtp.BodyExtractor{}
	out, err := ext.Extract(msg)
	require.NoError(t, err)
	require.Equal(t, plainBody, out)
}
