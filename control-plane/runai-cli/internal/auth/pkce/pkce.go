package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

var Plain Params

const (
	methodS256                   = "S256"
	CodeChallengeParamName       = "code_challenge"
	CodeChallengeMethodParamName = "code_challenge_method"
	CodeVerifierParamName        = "code_verifier"
)

type Params struct {
	CodeChallenge       string
	CodeChallengeMethod string
	CodeVerifier        string
}

func (p Params) IsZero() bool {
	return p.CodeChallenge == "" && p.CodeChallengeMethod == "" && p.CodeVerifier == ""
}

func New() (Params, error) {
	b, err := random32()
	if err != nil {
		return Plain, fmt.Errorf("could not generate a random: %v", err)
	}
	return computeS256(b), nil
}

func random32() ([]byte, error) {
	b := make([]byte, 32)
	if err := binary.Read(rand.Reader, binary.LittleEndian, b); err != nil {
		return nil, fmt.Errorf("read error: %v", err)
	}
	return b, nil
}

func computeS256(b []byte) Params {
	v := base64URLEncode(b)
	s := sha256.New()
	_, _ = s.Write([]byte(v))
	return Params{
		CodeChallenge:       base64URLEncode(s.Sum(nil)),
		CodeChallengeMethod: methodS256,
		CodeVerifier:        v,
	}
}

func base64URLEncode(b []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}
