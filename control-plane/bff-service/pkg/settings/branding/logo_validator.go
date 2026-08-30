package branding

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type logoValidator interface {
	DecodeLogo(encodedLogo string) (*[]byte, *logoError)
	ValidateMimeType(logo *[]byte) *logoError
	ValidateLogoSize(logo *[]byte) *logoError
}

type logoError struct {
	Code    int
	Message string
	Err     error
}

func newLogoValidate() logoValidator {
	return &logoValidate{}
}

type logoValidate struct{}

func (*logoValidate) DecodeLogo(encodedLogo string) (*[]byte, *logoError) {
	base64Data := encodedLogo
	if strings.HasPrefix(encodedLogo, "data:") {
		parts := strings.SplitN(encodedLogo, ",", 2)
		if len(parts) != 2 {
			return nil, &logoError{
				Code:    http.StatusBadRequest,
				Message: "upload failed, invalid data URL format",
				Err:     fmt.Errorf("invalid data URL format"),
			}
		}
		base64Data = parts[1]
	}

	logob, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, &logoError{
			Code:    http.StatusBadRequest,
			Message: "upload failed, bad format",
			Err:     fmt.Errorf("failed to decode logo file content from base64: %w", err),
		}
	}
	return &logob, nil
}

func (lv *logoValidate) ValidateMimeType(logo *[]byte) *logoError {
	fileType := http.DetectContentType(*logo)

	if fileType == "image/png" ||
		fileType == "image/svg+xml" ||
		strings.HasPrefix(fileType, "text/xml") && lv.isSVG(*logo) ||
		strings.HasPrefix(fileType, "text/plain") && lv.isSVG(*logo) {
		return nil
	}

	return &logoError{
		Code:    http.StatusUnsupportedMediaType,
		Message: "unsupported file type, only PNG and SVG images are supported",
		Err:     fmt.Errorf("not supported %v file format for logo", fileType),
	}
}

func (lv *logoValidate) isSVG(data []byte) bool {
	content := strings.ToLower(string(data))
	if len(content) > 512 {
		content = content[:512]
	}
	return strings.Contains(content, "<svg") && strings.Contains(content, "xmlns")
}

func (lv *logoValidate) ValidateLogoSize(logo *[]byte) *logoError {
	if len(*logo) > 128*1024 {
		return &logoError{
			Code:    http.StatusRequestEntityTooLarge,
			Message: "logo size is too big, limit is 128kb",
			Err:     fmt.Errorf("logo size is too big, was %v bytes", len(*logo)),
		}
	}
	return nil
}

func getLogoIfValid(lv logoValidator, encodedLogo string) (*[]byte, *logoError) {
	logob, err := lv.DecodeLogo(encodedLogo)
	if err != nil {
		return nil, err
	}
	err = lv.ValidateLogoSize(logob)
	if err != nil {
		return nil, err
	}
	err = lv.ValidateMimeType(logob)
	if err != nil {
		return nil, err
	}
	return logob, nil
}
