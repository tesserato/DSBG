package parse

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	texttemplate "text/template"
	"time"
)

// SiteTemplates holds the pre-parsed templates for articles, index, and RSS.
type SiteTemplates struct {
	Article *texttemplate.Template
	Index   *texttemplate.Template
	RSS     *texttemplate.Template
}

// LoadTemplates parses all necessary templates from the embedded assets once at startup.
// It returns a SiteTemplates struct with initialized template pointers.
func LoadTemplates(assets fs.FS) (SiteTemplates, error) {
	var t SiteTemplates
	var err error

	funcMap := template.FuncMap{
		"genRelativeLink":   genRelativeLink,
		"stringsJoin":       strings.Join,
		"buildShareUrl":     BuildShareUrl,
		"lower":             strings.ToLower,
		"isImage":           IsImage,
		"articleSchemaType": ArticleSchemaType,
		"makeLink": func(title string) string {
			return strings.ReplaceAll(strings.ToLower(title), " ", "-") + "/"
		},
		// RSS-specific helpers.
		"htmlEscape": func(s string) string {
			buf := &strings.Builder{}
			template.HTMLEscape(buf, []byte(s))
			return buf.String()
		},
		"formatPubDate": func(timeObj interface{}) string {
			if tt, ok := timeObj.(time.Time); ok {
				return tt.Format(time.RFC1123Z)
			}
			return ""
		},
		"buildArticleURL": func(a Article, s Settings) string {
			return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.BaseUrl, "/"), strings.TrimPrefix(a.LinkToSelf, "/"))
		},
	}

	// Parse article template.
	t.Article, err = texttemplate.New("html-article.gohtml").Funcs(funcMap).ParseFS(assets, "src/assets/templates/html-article.gohtml")
	if err != nil {
		return t, fmt.Errorf("error parsing article template: %w", err)
	}

	// Parse index template.
	t.Index, err = texttemplate.New("html-index.gohtml").Funcs(funcMap).ParseFS(assets, "src/assets/templates/html-index.gohtml")
	if err != nil {
		return t, fmt.Errorf("error parsing index template: %w", err)
	}

	// Parse RSS template.
	t.RSS, err = texttemplate.New("rss.goxml").Funcs(funcMap).ParseFS(assets, "src/assets/templates/rss.goxml")
	if err != nil {
		return t, fmt.Errorf("error parsing RSS template: %w", err)
	}

	return t, nil
}
