/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
)

type BodyExtractor struct{}

func (ep *BodyExtractor) Extract(msg *mail.Message) (string, error) {
	mediaType, params, err := parseContentType(msg.Header)
	if err != nil {
		return "", err
	}

	var body string
	if isMultipart(mediaType) {
		body, err = handleMultipart(msg.Body, params["boundary"])
	} else {
		body, err = handleSinglePart(msg.Header, msg.Body)
	}
	if err != nil {
		return "", err
	}

	return html.UnescapeString(body), nil
}

func parseContentType(header mail.Header) (mediaType string, params map[string]string, err error) {
	contentType := header.Get("Content-Type")
	if contentType == "" {
		return "", nil, fmt.Errorf("mail: Content-Type header not in message")
	}
	mediaType, params, err = mime.ParseMediaType(contentType)
	return mediaType, params, err
}

func isMultipart(mediaType string) bool {
	return strings.HasPrefix(mediaType, "multipart/")
}

func handleSinglePart(header mail.Header, body io.Reader) (string, error) {
	transferEnc := header.Get("Content-Transfer-Encoding")
	bodyStr, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	content := string(bodyStr)
	return decodeIfNeeded(content, transferEnc)
}

func handleMultipart(body io.Reader, boundary string) (string, error) {
	mr := multipart.NewReader(body, boundary)
	var parts []string
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		partBody, err := io.ReadAll(part)
		if err != nil {
			return "", err
		}
		transferEnc := part.Header.Get("Content-Transfer-Encoding")
		decoded, err := decodeIfNeeded(string(partBody), transferEnc)
		if err != nil {
			return "", err
		}
		parts = append(parts, decoded)
		parts = append(parts, "\r\n")
	}
	return strings.Join(parts, ""), nil
}

func decodeIfNeeded(content, transferEnc string) (string, error) {
	if strings.EqualFold(transferEnc, "quoted-printable") {
		return parseQuotedPrintable(content)
	}
	return content, nil
}

// decodes the email body if it's Content-Transfer-Encoding is quoted-printable
func parseQuotedPrintable(body string) (string, error) {
	reader := quotedprintable.NewReader(strings.NewReader(body))
	decodedBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
