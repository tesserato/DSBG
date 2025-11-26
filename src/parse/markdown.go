package parse

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	texttemplate "text/template"
	"time"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	rendererhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

// Markdown is the configured Goldmark Markdown parser with frontmatter support.
var Markdown = goldmark.New(
	goldmark.WithRendererOptions(
		rendererhtml.WithUnsafe(),
		rendererhtml.WithHardWraps(),
		rendererhtml.WithXHTML(),
	),
	goldmark.WithParserOptions(
		parser.WithAttribute(),
		parser.WithAutoHeadingID(),
	),
	goldmark.WithExtensions(
		&frontmatter.Extender{},
		mathjax.MathJax,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.Footnote,
		extension.Typographer,
	),
)

// MarkdownFile parses a Markdown file, extracts frontmatter, and populates an Article.
// It returns the Article and a list of extracted resource paths (e.g., images).
func MarkdownFile(path string) (Article, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to read Markdown file '%s': %w", path, err)
	}

	// Create a context to store frontmatter.
	context := parser.NewContext()

	// Parse the Markdown content into an AST.
	p := Markdown.Parser()
	reader := text.NewReader(data)
	doc := p.Parse(reader, parser.WithContext(context))

	// Extract resources from the AST (images and links).
	var resources []string
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Extract images
		if n.Kind() == ast.KindImage {
			if img, ok := n.(*ast.Image); ok {
				resources = append(resources, string(img.Destination))
			}
		}
		// Extract generic links (useful for linked PDFs, videos, zip files, etc.)
		if n.Kind() == ast.KindLink {
			if link, ok := n.(*ast.Link); ok {
				resources = append(resources, string(link.Destination))
			}
		}
		return ast.WalkContinue, nil
	})

	// Render to HTML.
	var buf bytes.Buffer
	r := Markdown.Renderer()
	if err := r.Render(&buf, data, doc); err != nil {
		return Article{}, nil, fmt.Errorf("failed to render Markdown to HTML for '%s': %w", path, err)
	}
	rawHtmlContent := buf.String()

	// Wrap tables in divs for potential CSS styling.
	wrappedHtmlContent, err := wrapTables(rawHtmlContent)
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to wrap tables for '%s': %w", path, err)
	}

	// Initialize Article with basic information.
	article := Article{
		OriginalPath: path,
		TextContent:  string(data),
		HtmlContent:  wrappedHtmlContent,
	}

	// Decode frontmatter into the Article.
	if fm := frontmatter.Get(context); fm != nil {
		var d map[string]any
		if err := fm.Decode(&d); err != nil {
			return Article{}, nil, fmt.Errorf("failed to decode frontmatter in '%s': %w", path, err)
		}
		for name, value := range d {
			name = strings.ToLower(strings.TrimSpace(name))
			if value == nil {
				continue
			}
			switch name {
			case "title":
				article.Title = value.(string)
			case "description":
				article.Description = value.(string)
			case "created":
				switch reflect.TypeOf(value).Kind() {
				case reflect.String:
					createdTime, err := DateTimeFromString(value.(string))
					if err != nil {
						log.Printf("Warning: Failed to parse 'created' date in '%s': %v\n", path, err)
					} else {
						article.Created = createdTime
					}
				default:
					if t, ok := value.(time.Time); ok {
						article.Created = t
					}
				}
			case "updated":
				switch reflect.TypeOf(value).Kind() {
				case reflect.String:
					updatedTime, err := DateTimeFromString(value.(string))
					if err != nil {
						log.Printf("Warning: Failed to parse 'updated' date in '%s': %v\n", path, err)
					} else {
						article.Updated = updatedTime
					}
				default:
					if t, ok := value.(time.Time); ok {
						article.Updated = t
					}
				}
			case "cover_image":
				article.CoverImage = value.(string)
			case "share_url":
				article.ShareUrl = value.(string)
			case "canonical_url":
				article.CanonicalUrl = value.(string)
			case "tags":
				switch reflect.TypeOf(value).Kind() {
				case reflect.Slice:
					tags := value.([]any)
					for _, tag := range tags {
						tagString := strings.TrimSpace(tag.(string))
						article.Tags = append(article.Tags, tagString)
					}
				case reflect.String:
					tags := strings.ReplaceAll(value.(string), ";", ",")
					tagsArray := strings.Split(tags, ",")
					for i, tag := range tagsArray {
						tagsArray[i] = strings.TrimSpace(tag)
					}
					article.Tags = tagsArray
				}
			}
		}
	}

	// Set Created and Updated to file dates if not provided in frontmatter.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to get file info for '%s': %w", path, err)
	}
	if article.Created.IsZero() {
		if createdFromFile, err := DateTimeFromString(path); err == nil {
			article.Created = createdFromFile
		} else {
			article.Created = fileInfo.ModTime()
		}
	}
	if article.Updated.IsZero() {
		article.Updated = fileInfo.ModTime()
	}

	// Default title to filename if not provided.
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return article, resources, nil
}

// FormatMarkdown applies an HTML template to the Markdown content of an article.
// It injects article and settings into the provided template and updates HtmlContent.
func FormatMarkdown(article *Article, settings Settings, tmpl *texttemplate.Template, assets fs.FS) error {
	var tp bytes.Buffer
	err := tmpl.Execute(&tp, struct {
		Art      Article
		Ctt      template.HTML
		Settings Settings
	}{
		Art:      *article,
		Ctt:      template.HTML(article.HtmlContent),
		Settings: settings,
	})
	if err != nil {
		return fmt.Errorf("error executing markdown article template: %w", err)
	}
	article.HtmlContent = tp.String()
	return nil
}
