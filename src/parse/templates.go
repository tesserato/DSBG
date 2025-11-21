package parse

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	texttemplate "text/template"
	"time"
)

// SiteTemplates holds the pre-parsed templates to avoid parsing them for every file.
type SiteTemplates struct {
	Article *texttemplate.Template
	Index   *texttemplate.Template
	RSS     *texttemplate.Template
	Style   *texttemplate.Template
}

// LoadTemplates parses all necessary templates from the embedded assets once at startup.
func LoadTemplates(assets fs.FS) (SiteTemplates, error) {
	var t SiteTemplates
	var err error

	// Define common template functions
	funcMap := template.FuncMap{
		"genRelativeLink": genRelativeLink,
		"stringsJoin":     strings.Join,
		"buildShareUrl":   BuildShareUrl,
		"lower":           strings.ToLower,
		"isImage":         IsImage, // Corrected to use exported IsImage
		"makeLink": func(title string) string {
			return strings.ReplaceAll(strings.ToLower(title), " ", "-") + "/"
		},
		// RSS specific functions
		"htmlEscape": func(s string) string {
			buf := &strings.Builder{}
			template.HTMLEscape(buf, []byte(s))
			return buf.String()
		},
		"formatPubDate": func(t interface{}) string {
			if tt, ok := t.(time.Time); ok {
				return tt.Format(time.RFC1123Z)
			}
			return ""
		},
		"buildArticleURL": func(a Article, s Settings) string {
			return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.BaseUrl, "/"), strings.TrimPrefix(a.LinkToSelf, "/"))
		},
	}

	// Parse Article Template
	t.Article, err = texttemplate.New("html-article.gohtml").Funcs(funcMap).ParseFS(assets, "src/assets/templates/html-article.gohtml")
	if err != nil {
		return t, fmt.Errorf("error parsing article template: %w", err)
	}

	// Parse Index Template
	t.Index, err = texttemplate.New("html-index.gohtml").Funcs(funcMap).ParseFS(assets, "src/assets/templates/html-index.gohtml")
	if err != nil {
		return t, fmt.Errorf("error parsing index template: %w", err)
	}

	// Parse RSS Template
	t.RSS, err = texttemplate.New("rss.goxml").Funcs(funcMap).ParseFS(assets, "src/assets/templates/rss.goxml")
	if err != nil {
		return t, fmt.Errorf("error parsing RSS template: %w", err)
	}

	// Parse Style Template
	t.Style, err = texttemplate.New("style.gocss").ParseFS(assets, "src/assets/templates/style.gocss")
	if err != nil {
		return t, fmt.Errorf("error parsing style template: %w", err)
	}

	return t, nil
}
