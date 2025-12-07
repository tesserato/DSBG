package parse

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	texttemplate "text/template"

	"github.com/k3a/html2text"
	"golang.org/x/net/html"
)

// HTMLFile parses an HTML file, extracts metadata, and populates an Article.
// It returns the Article and a list of extracted resources.
func HTMLFile(path string, settings Settings) (Article, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to read HTML file '%s': %w", path, err)
	}
	htmlContent := string(data)
	textContent := html2text.HTML2Text(htmlContent)

	article := Article{
		OriginalPath: path,
		HtmlContent:  htmlContent,
		TextContent:  textContent,
	}

	// Parse the HTML content into a node tree.
	htmlTree, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to parse HTML content of '%s': %w", path, err)
	}

	// Extract resources using the HTML tree.
	resources := ExtractResources(htmlTree)

	// Extract just the body content for RSS (excludes head/scripts/styles usually).
	bodyContent, err := getBodyContent(htmlTree)
	if err == nil {
		article.BodyContent = bodyContent
	} else {
		// Fallback to full content if extraction fails
		article.BodyContent = htmlContent
	}

	// Extract <title>.
	titleNode := findFirstElement(htmlTree, "title")
	if titleNode != nil && titleNode.FirstChild != nil {
		article.Title = titleNode.FirstChild.Data
	}

	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Extract relevant <meta> tags.
	for _, metaTag := range findAllElements(htmlTree, "meta") {
		key := ""
		val := ""
		for _, attr := range metaTag.Attr {
			switch strings.Trim(strings.ToLower(attr.Key), " ") {
			case "name":
				key = strings.Trim(strings.ToLower(attr.Val), " ")
			case "content":
				val = strings.Trim(attr.Val, " ")
			}
		}
		if key == "" || val == "" {
			continue
		}

		switch key {
		case "description":
			article.Description = val
		case "keywords":
			tags := strings.ReplaceAll(val, ";", ",")
			tagsArray := strings.Split(tags, ",")
			for i, tag := range tagsArray {
				tagsArray[i] = strings.TrimSpace(tag)
			}
			article.Tags = tagsArray
		case "created":
			createdTime, err := DateTimeFromString(val)
			if err != nil {
				if !settings.IgnoreErrors {
					return Article{}, nil, fmt.Errorf("failed to parse 'created' date in '%s': %w", path, err)
				}
				log.Printf("Warning: Failed to parse 'created' date from meta tag in '%s': %v\n", path, err)
			} else {
				article.Created = createdTime
			}
		case "updated":
			updatedTime, err := DateTimeFromString(val)
			if err != nil {
				if !settings.IgnoreErrors {
					return Article{}, nil, fmt.Errorf("failed to parse 'updated' date in '%s': %w", path, err)
				}
				log.Printf("Warning: Failed to parse 'updated' date from meta tag in '%s': %v\n", path, err)
			} else {
				article.Updated = updatedTime
			}
		case "cover_image":
			article.CoverImage = val
		case "link":
			article.ExternalLink = val
		case "canonical_url":
			article.CanonicalUrl = val
		}
	}

	// Set Created and Updated to file dates if not provided in meta tags.
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

	return article, resources, nil
}

// GenerateHtmlIndex creates an HTML index page listing all processed articles
// using the provided template and settings.
func GenerateHtmlIndex(articles []Article, settings Settings, tmpl *texttemplate.Template, assets fs.FS) error {
	var allTags []string
	var pageList []Article
	var articleList []Article

	for _, article := range articles {
		if slices.Contains(article.Tags, "PAGE") {
			pageList = append(pageList, article)
		} else {
			allTags = append(allTags, article.Tags...)
			articleList = append(articleList, article)
		}
	}

	var tp bytes.Buffer
	err := tmpl.Execute(&tp, struct {
		AllTags     []string
		PageList    []Article
		ArticleList []Article
		Settings    Settings
	}{
		AllTags:     allTags,
		PageList:    pageList,
		ArticleList: articleList,
		Settings:    settings,
	})
	if err != nil {
		return fmt.Errorf("error executing HTML index template: %w", err)
	}

	filePath := filepath.Join(settings.OutputPath, settings.IndexName)
	if err := os.WriteFile(filePath, tp.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing HTML index file to '%s': %w", filePath, err)
	}
	return nil
}

// findFirstElement recursively searches for the first HTML element with the given tag name.
func findFirstElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// findAllElements recursively searches for all HTML elements with the given tag name.
func findAllElements(n *html.Node, tag string) []*html.Node {
	var elements []*html.Node
	if n.Type == html.ElementNode && n.Data == tag {
		elements = append(elements, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		elements = append(elements, findAllElements(c, tag)...)
	}
	return elements
}

// wrapNodeIfTable wraps the provided table node in a div.table-wrapper for styling.
func wrapNodeIfTable(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "table" {
		div := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{Key: "class", Val: "table-wrapper"},
			},
		}
		n.Parent.InsertBefore(div, n)
		n.Parent.RemoveChild(n)
		div.AppendChild(n)
	}
}

// wrapTables wraps all <table> elements in div.table-wrapper and returns the updated HTML.
func wrapTables(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML content: %w", err)
	}
	tables := findAllElements(doc, "table")
	for _, table := range tables {
		wrapNodeIfTable(table)
	}
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("failed to render updated HTML content: %w", err)
	}
	return buf.String(), nil
}

// getBodyContent extracts the inner HTML of the <body> tag.
// If no body tag is found, it renders the entire document.
func getBodyContent(doc *html.Node) (string, error) {
	body := findFirstElement(doc, "body")
	if body == nil {
		// No body tag, maybe it's a fragment. Render everything.
		var buf bytes.Buffer
		if err := html.Render(&buf, doc); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	// Render all children of body
	var buf bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}
