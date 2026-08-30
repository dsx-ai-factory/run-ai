package inputs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	styles "github.com/run-ai/runai/runai-installer/installer/internal/tui/styles"
)

type Input interface {
	Render(isFocused bool) string
	HandleInput(char string)
	SelectInput(char string)
	DeleteInput()
	DeleteForward()
	IsRequired() bool
	GetValue() string
	SetValue(value string)
	SetSelected(selected bool)
	// Text cursor movement methods
	MoveCursorLeft()
	MoveCursorRight()
	MoveCursorToBeginning()
	MoveCursorToEnd()
	GetTextCursor() int
	GetInputIdentifier() configinput.InstallInputType
}

const (
	// DefaultVisibleWidth is the maximum number of characters visible in the input field
	DefaultVisibleWidth = 50
)

type TextInput struct {
	configurationItem *configinput.ConfigurationItem
	required          bool
	cursorPosition    int // Position of cursor within the text
	selected          bool
	value             string
	visibleWidth      int // Maximum visible characters in the input field
	viewportStart     int // Starting index of visible portion of text
}

// NewTextInput creates a new TextInput instance
func NewTextInput(configurationItem *configinput.ConfigurationItem, required bool) *TextInput {
	return &TextInput{
		configurationItem: configurationItem,
		required:          required,
		visibleWidth:      DefaultVisibleWidth,
		viewportStart:     0,
	}
}

func (i *TextInput) GetValue() string {
	return i.value
}

// SetValue programmatically sets the input value and adjusts cursor
func (i *TextInput) SetValue(value string) {
	i.value = value
	if i.cursorPosition > len(i.value) {
		i.cursorPosition = len(i.value)
	}
}

func (i *TextInput) IsRequired() bool {
	return i.required
}

func (i *TextInput) GetInputIdentifier() configinput.InstallInputType {
	return i.configurationItem.InputIdentifier
}

func (i *TextInput) DeleteInput() {
	if len(i.value) > 0 && i.cursorPosition > 0 {
		// Delete character at cursor position
		if i.cursorPosition >= len(i.value) {
			// Cursor at end, delete last character
			i.value = i.value[:len(i.value)-1]
		} else {
			// Delete character at cursor position
			i.value = i.value[:i.cursorPosition-1] + i.value[i.cursorPosition:]
		}
		i.cursorPosition--
	}
}

func (i *TextInput) DeleteForward() {
	if len(i.value) > 0 && i.cursorPosition < len(i.value) {
		// Delete character at cursor position
		i.value = i.value[:i.cursorPosition] + i.value[i.cursorPosition+1:]
	}
}

func (i *TextInput) SelectInput(char string) {
	i.selected = !i.selected
}

// SetSelected allows programmatic enable/disable
func (i *TextInput) SetSelected(selected bool) {
	i.selected = selected
}
func (i *TextInput) HandleInput(text string) {
	// Insert text at cursor position
	if i.cursorPosition >= len(i.value) {
		// Cursor at end, append text
		i.value += text
	} else {
		// Insert text at cursor position
		i.value = i.value[:i.cursorPosition] + text + i.value[i.cursorPosition:]
	}
	// Move cursor to end of inserted text
	i.cursorPosition += len(text)
}

// MoveCursorLeft moves the cursor one position to the left
func (i *TextInput) MoveCursorLeft() {
	if i.cursorPosition > 0 {
		i.cursorPosition--
	}
}

// MoveCursorRight moves the cursor one position to the right
func (i *TextInput) MoveCursorRight() {
	if i.cursorPosition < len(i.value) {
		i.cursorPosition++
	}
}

// MoveCursorToBeginning moves the cursor to the beginning of the text
func (i *TextInput) MoveCursorToBeginning() {
	i.cursorPosition = 0
}

// MoveCursorToEnd moves the cursor to the end of the text
func (i *TextInput) MoveCursorToEnd() {
	i.cursorPosition = len(i.value)
}

// GetTextCursor returns the current cursor position within the text
func (i *TextInput) GetTextCursor() int {
	return i.cursorPosition
}

// updateViewport adjusts the viewport to ensure the cursor is always visible
func (i *TextInput) updateViewport() {
	textLen := len(i.value)

	// Ensure cursor position is valid
	if i.cursorPosition > textLen {
		i.cursorPosition = textLen
	}
	if i.cursorPosition < 0 {
		i.cursorPosition = 0
	}

	// If text fits within visible width, no scrolling needed
	if textLen <= i.visibleWidth {
		i.viewportStart = 0
		return
	}

	// After deletion: if viewport would show empty space on the right, scroll left
	// to fill the visible area with as much text as possible
	maxViewportStart := textLen - i.visibleWidth
	if maxViewportStart < 0 {
		maxViewportStart = 0
	}
	if i.viewportStart > maxViewportStart {
		i.viewportStart = maxViewportStart
	}

	// Cursor is before the viewport - scroll left
	if i.cursorPosition < i.viewportStart {
		i.viewportStart = i.cursorPosition
	}

	// Cursor is beyond the viewport - scroll right
	// Leave room for the cursor indicator if at end of text
	visibleEnd := i.viewportStart + i.visibleWidth - 1 // -1 to leave room for cursor
	if i.cursorPosition > visibleEnd {
		i.viewportStart = i.cursorPosition - i.visibleWidth + 1
	}

	// Ensure viewport doesn't go negative
	if i.viewportStart < 0 {
		i.viewportStart = 0
	}
}

// renderWithCursorHighlight renders the text with a cursor indicator and horizontal scrolling
func (i *TextInput) renderWithCursorHighlight(normalStyle, cursorStyle lipgloss.Style, isFocused bool) string {
	// Update viewport to ensure cursor is visible
	i.updateViewport()

	text := i.value
	textLen := len(text)

	// Calculate visible portion boundaries
	viewStart := i.viewportStart
	viewEnd := viewStart + i.visibleWidth
	if viewEnd > textLen {
		viewEnd = textLen
	}

	// Extract visible portion of text
	visibleText := ""
	if viewStart < textLen {
		visibleText = text[viewStart:viewEnd]
	}

	// Add scroll indicators with app palette styling (using subtitle for visibility)
	scrollIndicatorStyle := styles.AppStyles.Subtitle
	leftIndicator := ""
	rightIndicator := ""
	if viewStart > 0 {
		leftIndicator = scrollIndicatorStyle.Render("◀") // Indicates more text to the left
	}
	if viewEnd < textLen {
		rightIndicator = scrollIndicatorStyle.Render("▶") // Indicates more text to the right
	}

	if !isFocused {
		// If not focused, render visible portion normally with indicators
		return leftIndicator + normalStyle.Render(visibleText) + rightIndicator
	}

	// For focused state, show the text with a cursor indicator
	// Calculate cursor position relative to viewport
	relativeCursorPos := i.cursorPosition - viewStart

	// Build the rendered text with cursor
	var result string

	if i.cursorPosition == textLen {
		// Cursor at end of text - show visible text with a cursor at the end
		result = normalStyle.Render(visibleText) + cursorStyle.Render("_")
	} else if relativeCursorPos >= 0 && relativeCursorPos < len(visibleText) {
		// Cursor is within visible area - highlight the character at cursor position
		beforeCursor := visibleText[:relativeCursorPos]
		atCursor := string(visibleText[relativeCursorPos])
		afterCursor := visibleText[relativeCursorPos+1:]
		result = normalStyle.Render(beforeCursor) + cursorStyle.Render(atCursor) + normalStyle.Render(afterCursor)
	} else {
		// Fallback - just render the visible text
		result = normalStyle.Render(visibleText)
	}

	return leftIndicator + result + rightIndicator
}
func (i *TextInput) Render(isFocused bool) string {
	return i.RenderInput(false, isFocused)
}

func (i *TextInput) RenderInput(isConditional, isFocused bool) string {
	// Use consistent purple styling with other screens
	focusedStyle := styles.AppStyles.InputFocused
	blurredStyle := styles.AppStyles.Normal
	placeholderStyle := styles.AppStyles.Placeholder

	// Choose styles based on focus state
	labelStyle := blurredStyle
	inputStyle := blurredStyle // Default to white text for input values

	if isFocused {
		labelStyle = focusedStyle
		inputStyle = focusedStyle
	}

	// Determine what to display in the input field
	displayValue := i.value
	if displayValue == "" {
		// Use a subtle placeholder when field is empty
		displayValue = placeholderStyle.Render(i.configurationItem.Placeholder)
	} else {
		// Use cursor highlight rendering to prevent text shifting
		displayValue = i.renderWithCursorHighlight(inputStyle, styles.AppStyles.Cursor, isFocused)
	}

	// Create the input field with proper styling
	inputField := displayValue

	// Format the label and input field
	label := i.getFullLabel()
	if isConditional {
		label = fmt.Sprintf("\t%s", label)
	}
	// Handle indented labels (like CA Chain cert path)
	labelWidth := 24
	if strings.HasPrefix(i.configurationItem.Label, "  ") {
		labelWidth = 26 // Extra space for indented labels
	}

	result := fmt.Sprintf("%-*s %s", labelWidth, labelStyle.Render(label), inputField)
	if isFocused {
		return styles.AppStyles.Selected.Render(result)
	}
	return result
}

func (t *TextInput) getFullLabel() string {
	if t.IsRequired() {
		return t.configurationItem.Label + " *"
	}
	return t.configurationItem.Label
}
