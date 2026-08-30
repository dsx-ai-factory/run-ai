package smtp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginAuth_ChallengeResponse(t *testing.T) {
	auth := newLoginAuth("alice", "secret")

	mech, ir, err := auth.Start()
	require.NoError(t, err)
	require.Equal(t, "LOGIN", mech)
	require.Nil(t, ir) // no initial response — required for Office365 / MailHog

	user, err := auth.Next([]byte("Username:"))
	require.NoError(t, err)
	require.Equal(t, []byte("alice"), user)

	pass, err := auth.Next([]byte("Password:"))
	require.NoError(t, err)
	require.Equal(t, []byte("secret"), pass)
}

func TestLoginAuth_UnexpectedChallenge(t *testing.T) {
	auth := newLoginAuth("alice", "secret")
	_, err := auth.Next([]byte("SomethingElse"))
	require.Error(t, err)
}

func TestNewSmtpClient_AuthLogin(t *testing.T) {
	cfg := &SmtpConfig{AuthType: AuthLogin, User: "u", Password: "p"}
	c, ok := NewSmtpClient(cfg).(*SmtpEmailSender)
	require.True(t, ok)
	require.NotNil(t, c.auth)

	mech, ir, err := c.auth.Start()
	require.NoError(t, err)
	require.Equal(t, "LOGIN", mech)
	require.Nil(t, ir)
}

func TestNewSmtpClient_AuthPlain(t *testing.T) {
	cfg := &SmtpConfig{AuthType: AuthPlain, User: "u", Password: "p"}
	c, ok := NewSmtpClient(cfg).(*SmtpEmailSender)
	require.True(t, ok)
	require.NotNil(t, c.auth)

	mech, _, err := c.auth.Start()
	require.NoError(t, err)
	require.Equal(t, "PLAIN", mech)
}

func TestNewSmtpClient_AuthNone(t *testing.T) {
	cfg := &SmtpConfig{AuthType: AuthNone}
	sender, ok := NewSmtpClient(cfg).(*SmtpEmailSender)
	require.True(t, ok)
	require.Nil(t, sender.auth)
}

func TestNewSmtpClient_UnknownType_DefaultsToLogin(t *testing.T) {
	cfg := &SmtpConfig{AuthType: "auth_oauth", User: "u", Password: "p"}
	c, ok := NewSmtpClient(cfg).(*SmtpEmailSender)
	require.True(t, ok)
	require.NotNil(t, c.auth)

	mech, ir, err := c.auth.Start()
	require.NoError(t, err)
	require.Equal(t, "LOGIN", mech)
	require.Nil(t, ir)
}
