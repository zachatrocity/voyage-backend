package email

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// ExtractBodyPreview reads the .eml file at filename, parses MIME,
// extracts plain text (fallback to HTML with tags stripped), truncates to maxLen chars.
// Returns empty string (not error) if parsing fails.
func ExtractBodyPreview(filename string, maxLen int) string {
	return truncate(extractBodyText(filename), maxLen)
}

// ExtractBodyFull reads the .eml file and returns the full normalized body text.
// Returns empty string (not error) if parsing fails.
func ExtractBodyFull(filename string) string {
	return extractBodyText(filename)
}

// ExtractBodyHTML reads the .eml file and returns sanitized HTML suitable for rendering.
// Returns empty string if no HTML content is available.
func ExtractBodyHTML(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	msg, err := mail.ReadMessage(f)
	if err != nil {
		return ""
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		body, _ := io.ReadAll(msg.Body)
		return plainTextToHTML(cleanText(string(body)))
	}

	var htmlBody, plainBody string
	if strings.HasPrefix(mediaType, "multipart/") {
		htmlBody, plainBody = extractMultipartBodies(msg.Body, params["boundary"])
	} else {
		decoded := decodeBody(msg.Body, msg.Header.Get("Content-Transfer-Encoding"))
		if strings.HasPrefix(mediaType, "text/html") {
			htmlBody = decoded
		} else {
			plainBody = decoded
		}
	}

	if strings.TrimSpace(htmlBody) != "" {
		return sanitizeHTML(htmlBody)
	}
	if strings.TrimSpace(plainBody) != "" {
		return plainTextToHTML(cleanText(plainBody))
	}
	return ""
}

func extractBodyText(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	msg, err := mail.ReadMessage(f)
	if err != nil {
		return ""
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Try reading body as plain text
		body, _ := io.ReadAll(msg.Body)
		return cleanText(string(body))
	}

	var text string
	if strings.HasPrefix(mediaType, "multipart/") {
		text = extractFromMultipart(msg.Body, params["boundary"])
	} else {
		decoded := decodeBody(msg.Body, msg.Header.Get("Content-Transfer-Encoding"))
		if strings.HasPrefix(mediaType, "text/html") {
			text = stripHTML(decoded)
		} else {
			text = decoded
		}
	}

	return cleanText(text)
}

func extractFromMultipart(r io.Reader, boundary string) string {
	htmlText, plainText := extractMultipartBodies(r, boundary)
	if plainText != "" {
		return plainText
	}
	return stripHTML(htmlText)
}

func extractMultipartBodies(r io.Reader, boundary string) (htmlText string, plainText string) {
	if boundary == "" {
		return "", ""
	}

	mr := multipart.NewReader(r, boundary)

	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		ct := part.Header.Get("Content-Type")
		if ct == "" {
			ct = "text/plain"
		}
		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil {
			continue
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			nestedHTML, nestedPlain := extractMultipartBodies(part, params["boundary"])
			if htmlText == "" && nestedHTML != "" {
				htmlText = nestedHTML
			}
			if plainText == "" && nestedPlain != "" {
				plainText = nestedPlain
			}
			continue
		}

		decoded := decodeBody(part, part.Header.Get("Content-Transfer-Encoding"))

		if mediaType == "text/plain" && plainText == "" {
			plainText = decoded
		} else if mediaType == "text/html" && htmlText == "" {
			htmlText = decoded
		}
	}

	return htmlText, plainText
}

func decodeBody(r io.Reader, encoding string) string {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		decoded, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, r))
		if err != nil {
			return ""
		}
		return string(decoded)
	case "quoted-printable":
		decoded, err := io.ReadAll(quotedprintable.NewReader(r))
		if err != nil {
			return ""
		}
		return string(decoded)
	default:
		data, err := io.ReadAll(r)
		if err != nil {
			return ""
		}
		return string(data)
	}
}

func stripHTML(s string) string {
	// Remove style and script blocks
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	s = reStyle.ReplaceAllString(s, "")
	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	s = reScript.ReplaceAllString(s, "")
	// Remove tags
	s = htmlTagRe.ReplaceAllString(s, " ")
	// Decode common entities
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return s
}

func cleanText(s string) string {
	// Collapse whitespace
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func sanitizeHTML(s string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td", "img")
	policy.AllowAttrs("href").OnElements("a")
	policy.AllowAttrs("src", "alt", "title").OnElements("img")
	return policy.Sanitize(s)
}

func plainTextToHTML(s string) string {
	escaped := html.EscapeString(s)
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")
	return "<div>" + escaped + "</div>"
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return fmt.Sprintf("%s...", string(runes[:maxLen]))
}
