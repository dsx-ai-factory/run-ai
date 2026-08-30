package smtp

import (
	sasl "github.com/emersion/go-sasl"
	"github.com/rs/zerolog/log"
)

type SmtpClient interface {
	SendMail(from string, to []string, msg []byte) error
	Verify(from string, to []string) error
}

func NewSmtpClient(config *SmtpConfig) SmtpClient {
	var clientAuth sasl.Client
	switch config.AuthType {
	case AuthLogin:
		clientAuth = newLoginAuth(config.User, config.Password)
	case AuthPlain:
		clientAuth = sasl.NewPlainClient("", config.User, config.Password)
	case AuthNone:
		clientAuth = nil
	default:
		log.Warn().Msgf("Unsupported authentication type: %q. Defaulting to %q", config.AuthType, AuthLogin)
		clientAuth = newLoginAuth(config.User, config.Password)
	}
	return NewSmtpEmailSender(config, clientAuth)
}
