/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/crewjam/saml"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	"k8s.io/utils/ptr"
)

func GetStringFromCtx(ctx context.Context, value any) string {
	str, ok := ctx.Value(value).(string)
	if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing %s in ctx", value)
	}
	return str
}

func FormatDate(ctx context.Context, date *string) *time.Time {
	if date == nil || *date == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("can't parse empty string to time object")
		return nil
	}
	unixTime, err := epochStringToTime(*date)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("can't parse string date to time: %s. err: %v", *date, err.Error())
		return nil
	}
	return &unixTime
}

func ConvertArrayToMap(arr []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range arr {
		m[v] = true
	}
	return m
}

func Contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func TemplateBaseUrl(hostTemplate string, name string) (string, error) {
	t, err := template.New("hostTemplate").Parse(hostTemplate)
	if err != nil {
		return "", err
	}
	var value bytes.Buffer
	err = t.Execute(&value, strings.ToLower(name))
	if err != nil {
		return "", err
	}

	return value.String(), nil
}

const (
	temporaryPasswordLength = 16
	lowercaseChars          = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars              = "0123456789"
	specialChars            = "-@_+="
	allChars                = lowercaseChars + uppercaseChars + digitChars + specialChars
)

// GenerateTemporaryUserPassword returns a cryptographically secure random
// password suitable for use as a Keycloak temporary password. It always
// contains at least one lowercase letter, one uppercase letter, one digit,
// and one special character.
func GenerateTemporaryUserPassword() string {
	chars := []byte{
		randomChar(lowercaseChars),
		randomChar(uppercaseChars),
		randomChar(digitChars),
		randomChar(specialChars),
	}
	for len(chars) < temporaryPasswordLength {
		chars = append(chars, randomChar(allChars))
	}
	// Shuffle so the guaranteed characters are not in fixed positions.
	for i := len(chars) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func randomChar(charset string) byte {
	return charset[randomInt(len(charset))]
}

// randomInt returns a uniform random integer in [0, n) using crypto/rand.
func randomInt(n int) int {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		logging.Errorf("failed getting random int err: %+v. returninig 1", err)
		return n
	}
	return int(i.Int64())
}

func epochStringToTime(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now().Unix()
	if now >= sec {
		return time.Unix(sec, 0), nil
	}
	logging.Warnf("bad format of epoch time: %v. time should be in unix seconds", sec)
	return time.Time{}, fmt.Errorf("bad format of epoch time %v. time should be in unix seconds", sec)
}

func MergeListsWithUniqueness(list1 []string, list2 []string) []string {
	if len(list1) == 0 {
		return list2
	}
	if len(list2) == 0 {
		return list1
	}

	uniqueElements := make(map[string]bool)
	var combindList []string
	combindList = append(combindList, list1...)
	for _, element := range list1 {
		uniqueElements[element] = true
	}

	for _, element := range list2 {
		if _, ok := uniqueElements[element]; !ok {
			combindList = append(combindList, element)
		}
	}
	return combindList
}

func GetTenantNameFromHost(host string) string {
	return strings.ToLower(strings.Split(host, ".")[0])
}

func GetUserPassword(password *string) string {
	if password != nil && *password != "" {
		return *password
	}
	return GenerateTemporaryUserPassword()
}

func IsValidURL(input string) bool {
	parsedURL, err := url.Parse(input)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	return true
}

var allowHTTPScheme bool

func SetAllowHTTPScheme(enabled bool) {
	allowHTTPScheme = enabled
}

func isAllowedScheme(scheme string) bool {
	return scheme == "https" || (scheme == "http" && allowHTTPScheme)
}

func ValidateExternalURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if !isAllowedScheme(parsedURL.Scheme) {
		return fmt.Errorf("only https urls are allowed, got scheme %q", parsedURL.Scheme)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("url is missing a host")
	}
	if ip := net.ParseIP(host); ip != nil {
		if isDisallowedIP(ip) {
			return fmt.Errorf("url %q targets a disallowed address: %s", rawURL, ip)
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("url %q: failed to resolve host %q: %w", rawURL, host, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("url %q: host %q resolved to no addresses", rawURL, host)
	}
	for _, addr := range addrs {
		if isDisallowedIP(addr.IP) {
			return fmt.Errorf("url %q targets a disallowed address: %s", rawURL, addr.IP)
		}
	}
	return nil
}

func isDisallowedIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}
	addr = addr.Unmap()
	return addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified()
}

func DecodeBase64String(encodedBase64 string) (*[]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(encodedBase64)
	if err != nil {
		logging.Errorf("failed to decode base64 string. err: %s", err.Error())
		return nil, err
	}
	return &decodedStr, nil
}

func ByteArrayToReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func ValidateSamlMetadata(metadata []byte) error {
	// 1. Attempt to parse as a standard single EntityDescriptor
	var entity saml.EntityDescriptor
	entityErr := xml.Unmarshal(metadata, &entity)
	if entityErr == nil {
		if entity.EntityID == "" {
			return fmt.Errorf("SAML structural validation error: missing mandatory 'entityID' attribute")
		}
		return nil
	}

	// 2. Fallback: Parse as an EntitiesDescriptor aggregate
	var entities saml.EntitiesDescriptor
	if err := xml.Unmarshal(metadata, &entities); err != nil {
		// Report the single-descriptor failure as it's the most common format
		return fmt.Errorf("invalid XML formatting or layout schema: %w", entityErr)
	}

	// 3. Ensure the aggregate actually contains descriptors to read
	if len(entities.EntityDescriptors) == 0 {
		return fmt.Errorf("SAML structural validation error: EntitiesDescriptor aggregate contains no child EntityDescriptor blocks")
	}

	// 4. Validate the first child element
	if entities.EntityDescriptors[0].EntityID == "" {
		return fmt.Errorf("SAML structural validation error: the first entity descriptor in the aggregate is missing a mandatory 'entityID'")
	}

	return nil
}

func DecodeBase64ToReader(encodedBase64 string) (io.Reader, error) {
	decodedFile, err := DecodeBase64String(encodedBase64)
	if err != nil {
		return nil, err
	}
	reader := ByteArrayToReader(*decodedFile)
	return reader, nil
}

func safeHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after too many redirects")
			}
			if req.URL.Scheme != "https" && !allowHTTPScheme {
				return fmt.Errorf("only https redirects are allowed, got scheme %q", req.URL.Scheme)
			}
			return nil
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
				if err != nil {
					return nil, err
				}
				for _, ip := range ips {
					if isDisallowedIP(ip) {
						return nil, fmt.Errorf("blocked connection to disallowed address: %s", ip)
					}
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
			},
		},
	}
}

func ReadUrlContent(ctx context.Context, rawURL string) (*[]byte, error) {
	if err := ValidateExternalURL(rawURL); err != nil {
		logging.Errorf("rejected url for metadata fetch. err: %v", err.Error())
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		logging.Errorf("error building request for URL. err: %v", err.Error())
		return nil, err
	}
	response, err := safeHTTPClient().Do(req)
	if err != nil {
		logging.Errorf("error fetching XML from URL. err: %v", err.Error())
		return nil, err
	}
	defer response.Body.Close()
	xmlData, err := io.ReadAll(response.Body)
	if err != nil {
		logging.Errorf("error reading response body. err: %v", err.Error())
		return nil, err
	}
	return &xmlData, nil
}

func MergeMaps(dest map[string]string, src map[string]string) {
	for key, value := range src {
		if destValue, ok := dest[key]; !ok || destValue == value {
			dest[key] = value
		}
	}
}

func Intersection(slice1, slice2 []string) []string {
	elements := make(map[string]bool)
	result := []string{}
	if slice1 == nil || slice2 == nil {
		return result
	}
	for _, s := range slice1 {
		elements[s] = true
	}

	for _, s := range slice2 {
		if elements[s] {
			result = append(result, s)
			delete(elements, s)
		}
	}

	return result
}

func Iso8601ToEpoch(isoTime string) (string, error) {
	parsedTime, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return "", fmt.Errorf("failed to parse time %s. err: %s", isoTime, err.Error())
	}
	return strconv.FormatInt(parsedTime.Unix(), 10), nil
}

func GenerateClientID() *string {
	uuidStr := uuid.New().String()
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")
	return ptr.To(uuidStr[:16])
}

func JsonToMap(input interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("JsonToMap failed. failed marshaling object")
	}
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, fmt.Errorf("JsonToMap failed. failed unmarshaling object")
	}
	return result, nil
}

func FormatSecureCookie(name, value, path string, maxAge int) string {
	return fmt.Sprintf("%s=%s; HttpOnly; Secure; SameSite=Strict; Path=%s; Max-Age=%d",
		name, value, path, maxAge)
}

func FormatClearCookie(name, path string) string {
	return FormatSecureCookie(name, "", path, 0)
}

func CreateSetCookieHeaders(cookies ...string) map[string][]string {
	return map[string][]string{
		"Set-Cookie": cookies,
	}
}

func ShortHash() (string, bool) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", false
	}
	return hex.EncodeToString(b), true
}
