package email

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestEml(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.eml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractBodyPreview_PlainText(t *testing.T) {
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: text/plain\r\n\r\nHello, this is a plain text email body."
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if result != "Hello, this is a plain text email body." {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestExtractBodyPreview_HTMLOnly(t *testing.T) {
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: text/html\r\n\r\n<html><body><p>Hello <b>world</b></p></body></html>"
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "world") {
		t.Errorf("unexpected result: %q", result)
	}
	if strings.Contains(result, "<") {
		t.Errorf("HTML tags not stripped: %q", result)
	}
}

func TestExtractBodyPreview_Multipart(t *testing.T) {
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: multipart/alternative; boundary=\"boundary42\"\r\n\r\n--boundary42\r\nContent-Type: text/plain\r\n\r\nPlain text part\r\n--boundary42\r\nContent-Type: text/html\r\n\r\n<p>HTML part</p>\r\n--boundary42--\r\n"
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if result != "Plain text part" {
		t.Errorf("expected plain text part, got: %q", result)
	}
}

func TestExtractBodyPreview_MultipartHTMLFallback(t *testing.T) {
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: multipart/alternative; boundary=\"boundary42\"\r\n\r\n--boundary42\r\nContent-Type: text/html\r\n\r\n<p>Only HTML here</p>\r\n--boundary42--\r\n"
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if !strings.Contains(result, "Only HTML here") {
		t.Errorf("expected HTML fallback, got: %q", result)
	}
}

func TestExtractBodyPreview_Base64(t *testing.T) {
	// "Hello from base64" in base64
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: text/plain\r\nContent-Transfer-Encoding: base64\r\n\r\nSGVsbG8gZnJvbSBiYXNlNjQ=\r\n"
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if result != "Hello from base64" {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestExtractBodyPreview_QuotedPrintable(t *testing.T) {
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: text/plain\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\nHello =3D world\r\n"
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if result != "Hello = world" {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestExtractBodyPreview_Truncation(t *testing.T) {
	body := strings.Repeat("A", 500)
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: text/plain\r\n\r\n" + body
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	// 300 chars + "..."
	if len(result) != 303 {
		t.Errorf("expected length 303, got %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Errorf("expected '...' suffix, got: %q", result[len(result)-10:])
	}
}

func TestExtractBodyPreview_MissingFile(t *testing.T) {
	result := ExtractBodyPreview("/nonexistent/path/email.eml", 300)
	if result != "" {
		t.Errorf("expected empty string for missing file, got: %q", result)
	}
}

func TestExtractBodyPreview_InvalidContent(t *testing.T) {
	path := writeTestEml(t, "not a valid email at all\x00\x01\x02")
	result := ExtractBodyPreview(path, 300)
	// Should not error, just return something or empty
	_ = result
}

func TestExtractBodyPreview_HTMLWithStyleScript(t *testing.T) {
	eml := "From: test@example.com\r\nSubject: Test\r\nContent-Type: text/html\r\n\r\n<html><head><style>body{color:red}</style></head><body><script>alert('x')</script><p>Visible text</p></body></html>"
	path := writeTestEml(t, eml)

	result := ExtractBodyPreview(path, 300)
	if !strings.Contains(result, "Visible text") {
		t.Errorf("expected 'Visible text', got: %q", result)
	}
	if strings.Contains(result, "alert") || strings.Contains(result, "color:red") {
		t.Errorf("style/script content not stripped: %q", result)
	}
}
