package parse

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
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
		"rssUrl": safeRSSUrl,
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
								if val == "" || strings.HasPrefix(val, "mailto:") {
									continue
								}

								// If absolute, just escape/clean it.
								// If relative, make absolute based on article path.
								var finalUrl string
								if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") || strings.HasPrefix(val, "//") {
									finalUrl = safeRSSUrl(val, "") // No base needed for absolute
								} else {
									// LinkToSelf e.g. "posts/my-post/index.html" -> Dir "posts/my-post"
									baseDir := filepath.Dir(a.LinkToSelf)
									// Clean path join
									fullPath := filepath.Join(baseDir, val)
									// Make sure it uses forward slashes for URL consistency
									fullPath = filepath.ToSlash(fullPath)

									finalUrl = safeRSSUrl(fullPath, s.BaseUrl)
								}

								n.Attr[i].Val = finalUrl
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

// safeRSSUrl takes a URL (relative or absolute) and a base URL.
// It resolves the URL to be absolute and ensures path segments are properly escaped (e.g. spaces -> %20).
func safeRSSUrl(urlStr, baseUrl string) string {
	if urlStr == "" {
		return ""
	}

	// 1. Resolve to absolute URL if needed
	targetUrl := urlStr
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "//") {
		// Clean and join
		cleanBase := strings.TrimSuffix(baseUrl, "/")
		cleanPath := strings.TrimPrefix(urlStr, "/")
		targetUrl = fmt.Sprintf("%s/%s", cleanBase, cleanPath)
	}

	// 2. Parse the URL to handle escaping correctly
	u, err := url.Parse(targetUrl)
	if err != nil {
		// Fallback: simple escape if parse fails, though rare for valid input
		return EncodePathSegments(targetUrl)
	}

	// 3. Re-encode the path to ensure spaces and special chars are valid for XML/RSS
	// url.Parse decodes %20 back to space in u.Path, so we re-encode it safely.
	// EncodePathSegments splits by / and escapes each segment.
	u.Path = EncodePathSegments(u.Path)

	return u.String()
}
