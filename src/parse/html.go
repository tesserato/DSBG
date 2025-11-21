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

// HTMLFile parses an HTML file, extracts metadata from tags, and populates an Article struct.
// It returns the parsed Article and a list of extracted resources.
func HTMLFile(path string) (Article, []string, error) {
	// Read the HTML file content.
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to read HTML file '%s': %w", path, err)
	}
	htmlContent := string(data)
	textContent := html2text.HTML2Text(htmlContent)

	// Create an Article struct with basic information.
	article := Article{
		OriginalPath: path,
		HtmlContent:  htmlContent,
		TextContent:  textContent,
	}
	// Parse the HTML content.
	htmlTree, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to parse HTML content of '%s': %w", path, err)
	}

	// Extract resources using the existing HTML tree.
	resources := ExtractResources(htmlTree)

	// Get info from <title> tag.
	titleNode := findFirstElement(htmlTree, "title")
	if titleNode != nil && titleNode.FirstChild != nil {
		article.Title = titleNode.FirstChild.Data
	}

	// Default title to filename if not found in <title> tag.
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Get info from meta tags.
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
		// Populate Article fields based on meta tag content.
		if key != "" && val != "" {
			switch key {
			case "description":
				article.Description = val
			case "keywords":
				tags := strings.ReplaceAll(val, ";", ",")
				tagsArray := strings.Split(tags, ",")
				for i, tag := range tagsArray {
					tag = strings.Trim(tag, " ")
					tagsArray[i] = tag
				}
				article.Tags = tagsArray
			case "created":
				createdTime, err := DateTimeFromString(val)
				if err != nil {
					log.Printf("Warning: Failed to parse 'created' date from meta tag in '%s': %v\n", path, err)
				} else {
					article.Created = createdTime
				}
			case "updated":
				updatedTime, err := DateTimeFromString(val)
				if err != nil {
					log.Printf("Warning: Failed to parse 'updated' date from meta tag in '%s': %v\n", path, err)
				} else {
					article.Updated = updatedTime
				}
			case "coverimagepath":
				article.CoverImagePath = val
			case "url":
				article.Url = val
			}
		}
	}

	// Set Created and Updated to file dates if not provided in meta tags.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return Article{}, nil, fmt.Errorf("failed to get file info for '%s': %w", path, err)
	}
	if article.Created.IsZero() {
		createdFromFile, err := DateTimeFromString(path) // Try to extract date from filename
		if err != nil {
			article.Created = fileInfo.ModTime() // Use file modification time
		} else {
			article.Created = createdFromFile
		}
	}
	if article.Updated.IsZero() {
		article.Updated = fileInfo.ModTime() // Use file modification time
	}

	return article, resources, nil
}

// GenerateHtmlIndex creates an HTML index page listing all processed articles.
func GenerateHtmlIndex(articles []Article, settings Settings, tmpl *texttemplate.Template, assets fs.FS) error {
	// Separate articles into pages and regular articles based on tags.
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

	// Execute the template with article data.
	var tp bytes.Buffer
	err := tmpl.Execute(&tp, struct {
		AllTags     []string
		PageList    []Article
		ArticleList []Article
		Settings    Settings
	}{allTags, pageList, articleList, settings})
	if err != nil {
		return fmt.Errorf("error executing HTML index template: %w", err)
	}

	// Write the generated HTML to the output file.
	filePath := filepath.Join(settings.OutputDirectory, settings.IndexName)
	err = os.WriteFile(filePath, tp.Bytes(), 0644)
	if err != nil {
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

func wrapNodeIfTable(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "table" {
		// Create a div element
		div := &html.Node{
			Type: html.ElementNode,
			Data: "div",
		}
		// Add class "table-wrapper" to the div for potential CSS styling
		div.Attr = []html.Attribute{{Key: "class", Val: "table-wrapper"}}

		n.Parent.InsertBefore(div, n)
		n.Parent.RemoveChild(n)
		div.AppendChild(n)
	}
}

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
	html.Render(&buf, doc)
	return buf.String(), nil
}
