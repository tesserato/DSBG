package parse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	texttemplate "text/template"

	"golang.org/x/net/html"
)

func TestWrapTablesReturnsFragmentWithoutDocumentWrapper(t *testing.T) {
	t.Parallel()

	in := `<p>Before</p><table><tr><td>A</td></tr></table><p>After</p>`
	out, err := wrapTables(in)
	if err != nil {
		t.Fatalf("wrapTables returned error: %v", err)
	}

	if strings.Contains(strings.ToLower(out), "<html") || strings.Contains(strings.ToLower(out), "<body") {
		t.Fatalf("wrapTables should return an HTML fragment, got: %q", out)
	}
	if !strings.Contains(out, `class="table-wrapper"`) {
		t.Fatalf("wrapTables should wrap tables with .table-wrapper, got: %q", out)
	}
}

func TestGetBodyContentExtractsOnlyBodyInnerHTML(t *testing.T) {
	t.Parallel()

	doc := `<!DOCTYPE html><html><head><title>X</title></head><body><h1>Hello</h1><p>World</p></body></html>`
	node, err := parseHTMLNode(doc)
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	out, err := getBodyContent(node)
	if err != nil {
		t.Fatalf("getBodyContent returned error: %v", err)
	}

	lower := strings.ToLower(out)
	if strings.Contains(lower, "<html") || strings.Contains(lower, "<head") || strings.Contains(lower, "<body") {
		t.Fatalf("getBodyContent should not include wrapper tags, got: %q", out)
	}
	if !strings.Contains(out, "<h1>Hello</h1>") || !strings.Contains(out, "<p>World</p>") {
		t.Fatalf("getBodyContent missing expected body children, got: %q", out)
	}
}

func TestHTMLFileBodyContentExcludesDocumentWrapper(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.html")
	content := `<!DOCTYPE html><html><head><title>Doc</title><meta name="description" content="desc"></head><body><article><p>Inside</p></article></body></html>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp HTML file: %v", err)
	}

	article, _, err := HTMLFile(path, Settings{})
	if err != nil {
		t.Fatalf("HTMLFile returned error: %v", err)
	}

	lower := strings.ToLower(article.BodyContent)
	if strings.Contains(lower, "<html") || strings.Contains(lower, "<head") || strings.Contains(lower, "<body") {
		t.Fatalf("BodyContent should be inner body HTML, got: %q", article.BodyContent)
	}
	if !strings.Contains(article.BodyContent, "<article>") {
		t.Fatalf("BodyContent should include body content, got: %q", article.BodyContent)
	}
}

func TestFormatMarkdownDoesNotInjectNestedHTMLWrapper(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mdPath := filepath.Join(dir, "post.md")
	md := "# Title\n\n| A |\n|---|\n| B |\n"
	if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
		t.Fatalf("failed to write temp markdown file: %v", err)
	}

	article, _, err := MarkdownFile(mdPath, Settings{})
	if err != nil {
		t.Fatalf("MarkdownFile returned error: %v", err)
	}

	tmpl, err := texttemplate.New("test").Parse("<!DOCTYPE html><html><body><article>{{.Ctt}}</article></body></html>")
	if err != nil {
		t.Fatalf("failed to parse test template: %v", err)
	}

	if err := FormatMarkdown(&article, Settings{}, tmpl, nil); err != nil {
		t.Fatalf("FormatMarkdown returned error: %v", err)
	}

	outLower := strings.ToLower(article.HtmlContent)
	if strings.Count(outLower, "<html") != 1 {
		t.Fatalf("expected exactly one <html> tag in rendered document, got %d: %q", strings.Count(outLower, "<html"), article.HtmlContent)
	}
	if strings.Count(outLower, "<body") != 1 {
		t.Fatalf("expected exactly one <body> tag in rendered document, got %d: %q", strings.Count(outLower, "<body"), article.HtmlContent)
	}
	if !strings.Contains(article.HtmlContent, `class="table-wrapper"`) {
		t.Fatalf("expected wrapped table markup in article content, got: %q", article.HtmlContent)
	}
}

func parseHTMLNode(s string) (*html.Node, error) {
	return html.Parse(strings.NewReader(s))
}
