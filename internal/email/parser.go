package email

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"regexp"
	"strings"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// ExtractBodyPreview reads the .eml file at filename, parses MIME,
// extracts plain text (fallback to HTML with tags stripped), truncates to maxLen chars.
// Returns empty string (not error) if parsing fails.
func ExtractBodyPreview(filename string, maxLen int) string {
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
		return truncate(cleanText(string(body)), maxLen)
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

	return truncate(cleanText(text), maxLen)
}

func extractFromMultipart(r io.Reader, boundary string) string {
	if boundary == "" {
		return ""
	}

	mr := multipart.NewReader(r, boundary)
	var plainText, htmlText string

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
			nested := extractFromMultipart(part, params["boundary"])
			if nested != "" {
				if plainText == "" {
					plainText = nested
				}
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

	if plainText != "" {
		return plainText
	}
	return stripHTML(htmlText)
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

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return fmt.Sprintf("%s...", string(runes[:maxLen]))
}
