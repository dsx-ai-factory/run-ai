/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type LogoValidator interface {
	DecodeLogo(encodedLogo string) (*[]byte, *Error)
	ValidateMimeType(logo *[]byte) *Error
	ValidateLogoSize(logo *[]byte) *Error
}

func NewLogoValidate() LogoValidator {
	return &LogoValidate{}
}

type LogoValidate struct{}

func (*LogoValidate) DecodeLogo(encodedLogo string) (*[]byte, *Error) {
	base64Data := encodedLogo
	if strings.HasPrefix(encodedLogo, "data:") {
		parts := strings.SplitN(encodedLogo, ",", 2)
		if len(parts) != 2 {
			return nil, &Error{
				Code:    http.StatusBadRequest,
				Message: "upload failed, invalid data URL format",
				Err:     fmt.Errorf("invalid data URL format: %s", encodedLogo),
			}
		}
		base64Data = parts[1]
	}

	logob, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, &Error{
			Code:    http.StatusBadRequest,
			Message: "upload failed, bad format",
			Err:     fmt.Errorf("failed to decode logo file content from base64 error: : %e", err),
		}
	}
	return &logob, nil
}

func (lv *LogoValidate) ValidateMimeType(logo *[]byte) *Error {
	fileType := http.DetectContentType(*logo)

	if fileType == "image/png" ||
		fileType == "image/svg+xml" ||
		strings.HasPrefix(fileType, "text/xml") && lv.isSVG(*logo) ||
		strings.HasPrefix(fileType, "text/plain") && lv.isSVG(*logo) {
		return nil
	}

	return &Error{
		Code:    http.StatusUnsupportedMediaType,
		Message: "unsupported file type, only PNG and SVG images are supported",
		Err:     fmt.Errorf("not supported %v file format for logo", fileType),
	}
}

func (lv *LogoValidate) isSVG(data []byte) bool {
	content := strings.ToLower(string(data))
	if len(content) > 512 {
		content = content[:512]
	}
	return strings.Contains(content, "<svg") && strings.Contains(content, "xmlns")
}

func (lv *LogoValidate) ValidateLogoSize(logo *[]byte) *Error {
	if len(*logo) > 128*1024 {
		return &Error{
			Code:    http.StatusRequestEntityTooLarge,
			Message: "logo size is too big, limit is 128kb",
			Err:     fmt.Errorf("logo size is too big, was %v bytes", len(*logo)),
		}
	}
	return nil
}

func GetLogoIfValid(lv LogoValidator, encodedLogo string) (*[]byte, *Error) {
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
