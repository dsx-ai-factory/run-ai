# Accessibility Audit Script

A manual accessibility audit tool using axe-core that allows you to audit forms and pages in their current state without refreshing the page.

## Prerequisites

1. **Google Chrome** installed on your machine
2. **Node.js** (v18+)

## Usage

### 1. Launch Chrome with Remote Debugging

First, start Chrome with remote debugging enabled:

**macOS:**
```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222
```

**Linux:**
```bash
google-chrome --remote-debugging-port=9222
```

**Windows:**
```bash
"C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222
```

### 2. Prepare the Page to Audit

1. In the Chrome window that opened, navigate to the application
2. Log in if required
3. Navigate to the specific page or form you want to audit
4. **Important:** Set up the form state you want to test (fill in fields, expand sections, open modals, etc.)

### 3. Run the Audit Script

In a separate terminal, run:

```bash
npm run a11y:audit
```

This will automatically install the required dependencies (`puppeteer-core`, `@axe-core/puppeteer`) on first run.

The script will prompt you to press Enter when ready. The audit runs on the **current DOM state** without refreshing the page, preserving your form state.

### 4. Review Results

Reports are saved to the `a11y-reports/` directory with timestamps:
```
a11y-reports/<page-path>-<timestamp>.json
```

Example: `a11y-reports/projects-edit-v2-4500020-2026-02-03T14-05.json`

## Output Format

The JSON report contains axe-core results with:

- **violations**: Accessibility issues that need to be fixed
- **passes**: Rules that passed
- **incomplete**: Rules that couldn't be fully evaluated
- **inapplicable**: Rules that don't apply to the page

Each violation includes:
- **id**: Rule identifier (e.g., `button-name`, `label`)
- **impact**: Severity level (`critical`, `serious`, `moderate`, `minor`)
- **description**: What the rule checks
- **nodes**: Specific DOM elements that failed the rule

## Workflow Tips

1. **Multi-step forms**: Run the audit on each step separately to catch step-specific issues
2. **Modals/Dialogs**: Open the modal before running the audit
3. **Dropdown menus**: Expand dropdowns you want to audit
4. **Different states**: Test empty forms vs. filled forms, error states, etc.

## Interpreting Results

Focus on violations by impact:

| Impact | Priority |
|--------|----------|
| critical | Must fix - blocks users |
| serious | Should fix - significant barrier |
| moderate | Should fix - causes difficulty |
| minor | Nice to fix - minor inconvenience |

## Common Violations

| Rule ID | Description | Common Fix |
|---------|-------------|------------|
| `button-name` | Buttons need accessible names | Add `aria-label` to icon-only buttons |
| `label` | Form inputs need labels | Add `<label>` or `aria-label` |
| `aria-toggle-field-name` | Toggle controls need names | Add `aria-label` to checkboxes/switches |
| `aria-progressbar-name` | Progress bars need names | Add `aria-label` to progress indicators |
| `color-contrast` | Text needs sufficient contrast | Adjust colors to meet WCAG AA (4.5:1) |

## Troubleshooting

**"No browser found"**: Make sure Chrome is running with `--remote-debugging-port=9222`

**"Cannot connect"**: Close all Chrome windows and restart with the debug flag

**Empty report**: Ensure you're on a page with content before pressing Enter