package parse

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	texttemplate "text/template"
	"time"

	"golang.org/x/net/html"
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
		"absURL":            toAbsoluteUrl,
		"makeLink": func(title string) string {
			return strings.ReplaceAll(strings.ToLower(title), " ", "-") + "/"
		},
		"urlPathEscape": EncodePathSegments,
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
		"fixRSSContent": func(body string, a Article, s Settings) string {
			doc, err := html.Parse(strings.NewReader(body))
			if err != nil {
				return body
			}
			var f func(*html.Node)
			f = func(n *html.Node) {
				if n.Type == html.ElementNode {
					var attrName string
					switch n.Data {
					case "img", "audio", "video", "source", "track":
						attrName = "src"
					case "a", "link":
						attrName = "href"
					case "object":
						attrName = "data"
					}

					if attrName != "" {
						for i, attr := range n.Attr {
							if attr.Key == attrName {
								val := strings.TrimSpace(attr.Val)
								if val == "" || strings.HasPrefix(val, "http") || strings.HasPrefix(val, "//") || strings.HasPrefix(val, "mailto:") {
									continue
								}
								// Build absolute path
								// LinkToSelf e.g. "posts/my-post/index.html" -> Dir "posts/my-post"
								baseDir := filepath.Dir(a.LinkToSelf)
								// Clean path join
								fullPath := filepath.Join(baseDir, val)
								// Encode path segments (e.g. spaces)
								encodedPath := EncodePathSegments(fullPath)
								// Prepend base URL
								absoluteUrl := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.BaseUrl, "/"), encodedPath)
								n.Attr[i].Val = absoluteUrl
							}
						}
					}
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
			f(doc)

			// Render just the nodes (body is implicit in parse if not present, but parse adds html/head/body wrappers)
			// We want to render the content inside the body if it was added.
			bodyNode := findFirstElement(doc, "body")
			target := doc
			if bodyNode != nil {
				target = bodyNode
			}

			var buf bytes.Buffer
			// If we found a body, render its children.
			if bodyNode != nil {
				for c := target.FirstChild; c != nil; c = c.NextSibling {
					if err := html.Render(&buf, c); err != nil {
						return body
					}
				}
			} else {
				// Fallback
				if err := html.Render(&buf, doc); err != nil {
					return body
				}
			}

			return buf.String()
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
