package finding

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// FormatText writes a human-readable text representation of findings to w.
// Each finding is formatted as: file:line:col [SEVERITY] rule: message.
// Suggestion (when present) is shown on the next line with a "Suggestion:" prefix.
func FormatText(w io.Writer, findings []Finding) error {
	for _, f := range findings {
		_, err := fmt.Fprintf(
			w, "%s [%s] %s: %s\n",
			f.Position.String(),
			strings.ToUpper(string(f.Severity)),
			string(f.Rule),
			f.Message,
		)
		if err != nil {
			return fmt.Errorf("format text: %w", err)
		}

		if err := writeSuggestionLine(w, f.Suggestion, "  Suggestion: ", "format text"); err != nil {
			return err
		}
	}

	return nil
}

// FormatTextRich writes a human-readable text representation of findings to w
// with emoji severity badges and category display.
// Each finding is formatted as: file:line:col BADGE  rule: message [category].
// Severity badges use emoji + uppercase name (e.g., "🟠 ERROR").
// Category is shown in brackets when present. Suggestion is prefixed with 💡.
func FormatTextRich(w io.Writer, findings []Finding) error {
	for _, f := range findings {
		_, err := fmt.Fprintf(
			w, "%s %s  %s: %s",
			f.Position.String(),
			f.Severity.Badge(),
			string(f.Rule),
			f.Message,
		)
		if err != nil {
			return fmt.Errorf("format text rich: %w", err)
		}

		if f.Category != "" {
			_, err = fmt.Fprintf(w, " [%s]", string(f.Category))
			if err != nil {
				return fmt.Errorf("format text rich category: %w", err)
			}
		}

		_, err = fmt.Fprintln(w)
		if err != nil {
			return fmt.Errorf("format text rich newline: %w", err)
		}

		if err := writeSuggestionLine(w, f.Suggestion, "  💡 ", "format text rich"); err != nil {
			return err
		}
	}

	return nil
}

// writeSuggestionLine writes the suggestion line with the given prefix if non-empty.
// errLabel is used in the wrapped error for identifying which formatter failed.
func writeSuggestionLine(w io.Writer, suggestion, prefix, errLabel string) error {
	if suggestion == "" {
		return nil
	}

	_, err := fmt.Fprintf(w, "%s%s\n", prefix, suggestion)
	if err != nil {
		return fmt.Errorf("%s suggestion: %w", errLabel, err)
	}

	return nil
}

// FormatMarkdown writes a markdown table of findings to w.
func FormatMarkdown(w io.Writer, findings []Finding) error {
	_, err := fmt.Fprintf(w, "| Location | Severity | Rule | Message |\n")
	if err != nil {
		return fmt.Errorf("format markdown header: %w", err)
	}

	_, err = fmt.Fprintf(w, "|----------|----------|------|--------|\n")
	if err != nil {
		return fmt.Errorf("format markdown separator: %w", err)
	}

	const maxMessageLen = 80

	for _, f := range findings {
		msg := escapeMarkdownCell(f.Message, maxMessageLen)
		rule := escapeMarkdownCell(string(f.Rule), 0)

		_, err := fmt.Fprintf(
			w, "| %s | %s | %s | %s |\n",
			escapeMarkdownCell(f.Position.String(), 0),
			string(f.Severity),
			rule,
			msg,
		)
		if err != nil {
			return fmt.Errorf("format markdown row: %w", err)
		}
	}

	return nil
}

// escapeMarkdownCell escapes pipe and newline characters in a markdown table cell.
// If maxLen > 0, truncates the string at a rune boundary.
func escapeMarkdownCell(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")

	if maxLen > 3 && utf8.RuneCountInString(s) > maxLen {
		runes := []rune(s)
		s = string(runes[:maxLen-3]) + "..."
	}

	return s
}

// FormatTable writes a human-readable table of findings to w with severity badges,
// file:line, rule, message, and category columns.
func FormatTable(w io.Writer, findings []Finding) error {
	_, err := fmt.Fprintln(w, "SEVERITY    LOCATION          RULE        MESSAGE")
	if err != nil {
		return fmt.Errorf("format table header: %w", err)
	}

	for _, f := range findings {
		category := ""
		if f.Category != "" {
			category = fmt.Sprintf(" [%s]", string(f.Category))
		}

		_, err := fmt.Fprintf(
			w, "%-11s %-18s %-11s %s%s\n",
			f.Severity.Badge(),
			f.Position.String(),
			string(f.Rule),
			f.Message,
			category,
		)
		if err != nil {
			return fmt.Errorf("format table row: %w", err)
		}
	}

	return nil
}
