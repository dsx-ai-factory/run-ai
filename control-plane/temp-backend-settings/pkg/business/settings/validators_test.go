package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== ValidateBoolean ====================

func TestValidateBoolean_True(t *testing.T) {
	assert.Equal(t, "", ValidateBoolean(true))
}

func TestValidateBoolean_False(t *testing.T) {
	assert.Equal(t, "", ValidateBoolean(false))
}

func TestValidateBoolean_String(t *testing.T) {
	assert.NotEmpty(t, ValidateBoolean("true"))
}

func TestValidateBoolean_Number(t *testing.T) {
	assert.NotEmpty(t, ValidateBoolean(42))
}

func TestValidateBoolean_Nil(t *testing.T) {
	assert.NotEmpty(t, ValidateBoolean(nil))
}

// ==================== ValidateNumber ====================

func TestValidateNumber_Float64(t *testing.T) {
	assert.Equal(t, "", ValidateNumber(float64(3.14)))
}

func TestValidateNumber_Int(t *testing.T) {
	assert.Equal(t, "", ValidateNumber(42))
}

func TestValidateNumber_Float32(t *testing.T) {
	assert.Equal(t, "", ValidateNumber(float32(1.5)))
}

func TestValidateNumber_Int64(t *testing.T) {
	assert.Equal(t, "", ValidateNumber(int64(100)))
}

func TestValidateNumber_String(t *testing.T) {
	assert.NotEmpty(t, ValidateNumber("42"))
}

func TestValidateNumber_Bool(t *testing.T) {
	assert.NotEmpty(t, ValidateNumber(true))
}

func TestValidateNumber_Nil(t *testing.T) {
	assert.NotEmpty(t, ValidateNumber(nil))
}

// ==================== ValidateNone ====================

func TestValidateNone(t *testing.T) {
	assert.Equal(t, "", ValidateNone("anything"))
	assert.Equal(t, "", ValidateNone(nil))
	assert.Equal(t, "", ValidateNone(42))
}

// ==================== ValidateAdminMessage ====================

func TestValidateAdminMessage_ValidObjectWithMsg(t *testing.T) {
	val := map[string]interface{}{"msg": "<b>Hello</b>", "id": "1"}
	assert.Equal(t, "", ValidateAdminMessage(val))
}

func TestValidateAdminMessage_ValidObjectNoMsg(t *testing.T) {
	val := map[string]interface{}{"id": "1"}
	assert.Equal(t, "", ValidateAdminMessage(val))
}

func TestValidateAdminMessage_NotAMap(t *testing.T) {
	assert.Contains(t, ValidateAdminMessage("string"), "must be an object")
}

func TestValidateAdminMessage_MsgNotString(t *testing.T) {
	val := map[string]interface{}{"msg": 123}
	assert.Contains(t, ValidateAdminMessage(val), "must be a string")
}

func TestValidateAdminMessage_XSSEventHandler(t *testing.T) {
	val := map[string]interface{}{"msg": `<div onclick="alert(1)">test</div>`}
	assert.Equal(t, disallowedContentMsg, ValidateAdminMessage(val))
}

func TestValidateAdminMessage_XSSJavascriptProtocol(t *testing.T) {
	val := map[string]interface{}{"msg": `<a href="javascript:alert(1)">click</a>`}
	assert.Equal(t, disallowedContentMsg, ValidateAdminMessage(val))
}

func TestValidateAdminMessage_DisallowedTag(t *testing.T) {
	val := map[string]interface{}{"msg": `<script>alert(1)</script>`}
	assert.Equal(t, disallowedContentMsg, ValidateAdminMessage(val))
}

func TestValidateAdminMessage_AllowedTags(t *testing.T) {
	val := map[string]interface{}{"msg": `<b>bold</b> <i>italic</i> <a href="https://example.com">link</a>`}
	assert.Equal(t, "", ValidateAdminMessage(val))
}

// ==================== validateHTMLXSS ====================

func TestValidateHTMLXSS_SafeHTML(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS("<b>hello</b>"))
}

func TestValidateHTMLXSS_EventHandler(t *testing.T) {
	assert.Equal(t, disallowedContentMsg, validateHTMLXSS(`<div onmouseover="hack()">x</div>`))
}

func TestValidateHTMLXSS_VBScriptProtocol(t *testing.T) {
	assert.Equal(t, disallowedContentMsg, validateHTMLXSS(`<a href="vbscript:run">x</a>`))
}

func TestValidateHTMLXSS_DataProtocol(t *testing.T) {
	assert.Equal(t, disallowedContentMsg, validateHTMLXSS(`data:text/html,<script>alert(1)</script>`))
}

func TestValidateHTMLXSS_DisallowedTag(t *testing.T) {
	assert.Equal(t, disallowedContentMsg, validateHTMLXSS("<iframe src='evil.com'></iframe>"))
}

func TestValidateHTMLXSS_SafeHref(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS(`<a href="https://example.com">link</a>`))
}

func TestValidateHTMLXSS_MailtoHref(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS(`<a href="mailto:x@y.com">mail</a>`))
}

func TestValidateHTMLXSS_RelativeHref(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS(`<a href="/page">page</a>`))
}

func TestValidateHTMLXSS_AnchorHref(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS(`<a href="#section">anchor</a>`))
}

func TestValidateHTMLXSS_UnsafeHref(t *testing.T) {
	assert.Equal(t, disallowedContentMsg, validateHTMLXSS(`<a href="ftp://evil.com/file">bad</a>`))
}

func TestValidateHTMLXSS_PlainText(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS("just plain text"))
}

func TestValidateHTMLXSS_EmptyString(t *testing.T) {
	assert.Equal(t, "", validateHTMLXSS(""))
}
