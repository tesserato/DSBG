package parse

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/k3a/html2text"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	rendererhtml "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	"golang.org/x/net/html"
)

// regexPatterns defines a list of regular expression patterns to identify dates in strings.
var regexPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?P<year>\d{4})\D+(?P<month>\d{1,2})\D+(?P<day>\d{1,2})`),
	regexp.MustCompile(`(?P<day>\d{1,2})\D+(?P<month>\d{1,2})\D+(?P<year>\d{4})`),
	regexp.MustCompile(`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`),
}

// Configure Goldmark Markdown parser with frontmatter support.
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

// RemoveDateFromPath attempts to remove date patterns from a given string.
// It iterates through predefined regex patterns and replaces matching substrings with an empty string.
// Finally, it trims any leading/trailing hyphens, underscores, or spaces.
func RemoveDateFromPath(stringWithDate string) string {
	for _, regexPattern := range regexPatterns {
		stringWithDate = regexPattern.ReplaceAllString(stringWithDate, "")
	}
	stringWithDate = strings.Trim(stringWithDate, "-_ ")
	return stringWithDate
}

// DateTimeFromString attempts to parse a date and time from a string using predefined regex patterns.
// It extracts named capture groups (year, month, day, hour, min, sec) and constructs a time.Time value.
// Returns an error if a matched group cannot be converted to an integer or if no date information is found.
func DateTimeFromString(date string) (time.Time, error) {
	m := make(map[string]int)
	foundMatch := false
	for _, pattern := range regexPatterns {
		matches := pattern.FindStringSubmatch(date)
		if len(matches) > 0 {
			foundMatch = true
			for i, name := range pattern.SubexpNames()[1:] {
				if name == "" { // Skip unnamed capture groups
					continue
				}
				integer, err := strconv.Atoi(matches[i+1])
				if err != nil {
					return time.Time{}, fmt.Errorf("failed to convert '%s' to integer in '%s': %w", matches[i+1], date, err)
				}
				m[name] = integer
			}
		}
	}

	if !foundMatch {
		return time.Time{}, fmt.Errorf("no date information found in '%s'", date)
	}

	year := m["year"]
	month := time.Month(m["month"])
	day := m["day"]
	hour := m["hour"]
	min := m["min"]
	sec := m["sec"]
	dateTime := time.Date(year, month, day, hour, min, sec, 0, time.UTC)
	return dateTime, nil
}

// GetPaths retrieves all file paths within a directory and its subdirectories
// that match the specified extensions.
// Returns a slice of file paths and an error if the directory walk fails.
func GetPaths(root string, extensions []string) ([]string, error) {
	var files []string
	extMap := make(map[string]bool) // Create a map for efficient extension lookup

	for _, ext := range extensions {
		extLower := strings.ToLower(strings.TrimSpace(ext)) // Normalize extension (lowercase, trim whitespace)
		extMap[extLower] = true                             // Add normalized extension to the map for faster checking
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path)) // Get and normalize file extension
			if extMap[ext] {                           // Check if extension is in the allowed set
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// cleanString removes non-alphanumeric characters (except '/', '\', '.'), replaces backslashes with forward slashes,
// trims hyphens, underscores, and spaces from path segments, and joins them with hyphens.
func cleanString(url string) string {
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\/\\\. ]+`)
	url = nonAlphanumericRegex.ReplaceAllString(url, "")
	url = strings.ReplaceAll(url, "\\", "/")
	pieces := strings.Split(url, "/")
	for i, piece := range pieces {
		pieces[i] = strings.Trim(piece, "-_ ")
	}
	url = strings.Join(pieces, "/")
	pieces = strings.Fields(url)
	for i, piece := range pieces {
		pieces[i] = strings.Trim(piece, "-_ ")
	}
	url = strings.Join(pieces, "-")

	url = strings.Trim(url, "-")
	return url
}

// CopyHtmlResources copies associated resources (like images, scripts) for an article,
// determines the output path, and handles special cases for "PAGE" tagged articles.
// Returns an error if any file operations fail.
func CopyHtmlResources(settings Settings, article *Article) error {
	relativeInputPath, err := filepath.Rel(settings.InputDirectory, article.OriginalPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path for '%s': %w", article.OriginalPath, err)
	}

	// Optionally remove date from the article title based on settings.
	if !settings.DoNotRemoveDateFromTitles {
		datelessTitle := RemoveDateFromPath(article.Title)
		if datelessTitle != "" {
			article.Title = datelessTitle
		}
	}

	// Optionally extract tags from the input path based on settings.
	if !settings.DoNotExtractTagsFromPaths {
		relativeInputPathNoDate := RemoveDateFromPath(relativeInputPath)
		relativeInputPathNoDate = filepath.Clean(relativeInputPathNoDate)
		pathTags := strings.Split(relativeInputPathNoDate, string(os.PathSeparator))
		for i, tag := range pathTags {
			pathTags[i] = strings.Trim(tag, "-_ ")
		}
		if len(pathTags) > 1 {
			pathTags = pathTags[:len(pathTags)-1]
			for _, tag := range pathTags {
				if !slices.Contains(article.Tags, tag) {
					article.Tags = append(article.Tags, tag)
				}
			}
		}
	}

	// Determine the output path for the article.
	outputPath := filepath.Join(settings.OutputDirectory, relativeInputPath)
	outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	outputPath = filepath.Join(outputPath, settings.IndexName)

	// Optionally remove date from the output path based on settings.
	if !settings.DoNotRemoveDateFromPaths {
		datelessOutputPath := RemoveDateFromPath(outputPath)
		if !(strings.Contains(datelessOutputPath, "\\") || strings.Contains(datelessOutputPath, "//")) {
			outputPath = datelessOutputPath
		}
	}
	outputPath = cleanString(outputPath)
	outputDirectory := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDirectory, err)
	}

	originalDirectory := filepath.Dir(article.OriginalPath)

	// Copy the cover image if it exists.
	if article.CoverImagePath != "" {
		coverImageOrigPath := filepath.Join(originalDirectory, article.CoverImagePath)
		coverImageDestPath := filepath.Join(settings.OutputDirectory, article.CoverImagePath)
		fmt.Printf("  Copying %s to %s\n", coverImageOrigPath, coverImageDestPath)

		if !(slices.Contains(article.Tags, "PAGE")) {
			file, err := os.ReadFile(coverImageOrigPath)
			if err != nil {
				return fmt.Errorf("error reading file '%s': %w", coverImageOrigPath, err)
			}

			err = os.WriteFile(coverImageDestPath, file, 0644)
			if err != nil {
				return fmt.Errorf("error writing file '%s': %w", coverImageDestPath, err)
			}
		}

	}

	// Copy individual resources (images, scripts, etc.) linked in the HTML content.
	resourcePaths, err := extractResources(article.HtmlContent)
	if err != nil {
		return fmt.Errorf("failed to extract resources from '%s': %w", article.OriginalPath, err)
	}
	for _, resourceOrigRelPath := range resourcePaths {
		resourceOrigRelPathLower := strings.ToLower(resourceOrigRelPath)
		if strings.Contains(resourceOrigRelPathLower, "http") {
			continue
		}
		resourceOrigPath := filepath.Join(originalDirectory, resourceOrigRelPath)
		resourceDestPath := filepath.Join(outputDirectory, resourceOrigRelPath)
		fmt.Printf("  Copying %s to %s\n", resourceOrigPath, resourceDestPath)

		input, err := os.ReadFile(resourceOrigPath)
		if err != nil {
			return fmt.Errorf("failed to read resource file '%s': %w", resourceOrigPath, err)
		}

		err = os.MkdirAll(filepath.Dir(filepath.FromSlash(resourceDestPath)), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for resource '%s': %w", resourceDestPath, err)
		}

		err = os.WriteFile(resourceDestPath, input, 0644)
		if err != nil {
			return fmt.Errorf("failed to write resource file to '%s': %w", resourceDestPath, err)
		}
	}

	// Set the relative link to the generated article.
	LinkToSelf, err := filepath.Rel(settings.OutputDirectory, outputPath)
	if err != nil {
		return fmt.Errorf("failed to get relative link from '%s' to '%s': %w", settings.OutputDirectory, outputPath, err)
	}
	article.LinkToSelf = filepath.ToSlash(LinkToSelf)
	article.LinkToSave = filepath.ToSlash(outputPath)
	return nil
}

// GenerateHtmlIndex creates an HTML index page listing all processed articles.
// Returns an error if template parsing or execution fails, or if writing the output file fails.
func GenerateHtmlIndex(articles []Article, settings Settings, assets fs.FS) error {
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

	// Define template functions.
	funcMap := template.FuncMap{
		"makeLink": func(title string) string {
			return strings.ReplaceAll(strings.ToLower(title), " ", "-") + "/"
		},
		"stringsJoin":    strings.Join,
		"slicesContains": slices.Contains[[]string],
		"gen_share_url":  gen_share_url,
	}
	// Load the HTML template from assets
	htmlIndexTemplate, err := texttemplate.New("html-index.gohtml").Funcs(funcMap).ParseFS(assets, "assets/templates/html-index.gohtml")
	if err != nil {
		return fmt.Errorf("error parsing article index template: %w", err)
	}

	// Execute the template with article data.
	var tp bytes.Buffer
	err = htmlIndexTemplate.Execute(&tp, struct {
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

// GenerateRSS creates an RSS feed XML file from the processed articles.
// It ensures proper formatting for RSS elements like title, link, description, and pubDate.
// Returns an error if template parsing or execution fails, or if writing the output file fails.
func GenerateRSS(articles []Article, settings Settings, assets fs.FS) error {
	// Sort articles by creation date in descending order for the RSS feed.
	slices.SortFunc(articles, func(a, b Article) int {
		return b.Created.Compare(a.Created)
	})

	// Define template functions specific to RSS generation.
	funcMap := texttemplate.FuncMap{
		"htmlEscape": func(s string) string {
			// Escape HTML entities in the description to ensure valid XML.
			buf := &strings.Builder{}
			template.HTMLEscape(buf, []byte(s))
			return buf.String()
		},
		"formatPubDate": func(t time.Time) string {
			// Format the publication date according to RFC1123Z (required by RSS).
			return t.Format(time.RFC1123Z)
		},
		"buildArticleURL": func(a Article, s Settings) string {
			// Construct the full URL for the article.
			return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.BaseUrl, "/"), strings.TrimPrefix(a.LinkToSelf, "/"))
		},
	}

	// Parse the RSS template with the custom function map.

	tmpl, err := texttemplate.New("rss.goxml").Funcs(funcMap).ParseFS(assets, "assets/templates/rss.goxml")
	if err != nil {
		return fmt.Errorf("error parsing RSS template: %w", err)
	}

	// Execute the template with article data and settings.
	var tp bytes.Buffer
	err = tmpl.Execute(&tp, struct {
		Articles  []Article
		Settings  Settings
		BuildDate string // Add a build date for the feed.
	}{articles, settings, time.Now().Format(time.RFC1123Z)})
	if err != nil {
		return fmt.Errorf("error executing RSS template: %w", err)
	}

	// Write the generated RSS feed to the output file.
	filePath := filepath.Join(settings.OutputDirectory, "rss.xml")
	err = os.WriteFile(filePath, tp.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing RSS file to '%s': %w", filePath, err)
	}
	return nil
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

// MarkdownFile parses a Markdown file, extracts frontmatter, and populates an Article struct.
// Returns the parsed Article and an error if reading, parsing, or decoding fails.
func MarkdownFile(path string) (Article, error) {
	// Read the Markdown file content.
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read Markdown file '%s': %w", path, err)
	}

	// Create a context to store frontmatter.
	context := parser.NewContext()

	// Parse the Markdown content and render to HTML, storing frontmatter in the context.
	var buf strings.Builder
	if err := Markdown.Convert(data, &buf, parser.WithContext(context)); err != nil {
		return Article{}, fmt.Errorf("failed to convert Markdown to HTML for '%s': %w", path, err)
	}
	rawHtmlContent := buf.String()

	// Wrap tables in divs for potential CSS styling.
	wrappedHtmlContent, err := wrapTables(rawHtmlContent)
	if err != nil {
		return Article{}, fmt.Errorf("failed to wrap tables for '%s': %w", path, err)
	}

	// Create an Article struct with basic information.

	// Retrieve frontmatter from the context.
	var article = Article{OriginalPath: path, TextContent: string(data), HtmlContent: wrappedHtmlContent}
	fm := frontmatter.Get(context)
	if fm != nil {
		var d map[string]any
		if err := fm.Decode(&d); err != nil {
			return Article{}, fmt.Errorf("failed to decode frontmatter in '%s': %w", path, err)
		}
		// Populate Article fields from frontmatter.
		for name, value := range d {
			name = strings.ToLower(name)
			name = strings.Trim(name, " ")
			if value == nil {
				continue
			}
			switch name {
			case "title":
				article.Title = value.(string)
			case "description":
				article.Description = value.(string)
			case "created":
				if reflect.TypeOf(value).Kind() == reflect.String {
					createdTime, err := DateTimeFromString(value.(string))
					if err != nil {
						fmt.Printf("Warning: Failed to parse 'created' date in '%s': %v\n", path, err)
					} else {
						article.Created = createdTime
					}
				} else if t, ok := value.(time.Time); ok {
					article.Created = t
				}
			case "updated":
				if reflect.TypeOf(value).Kind() == reflect.String {
					updatedTime, err := DateTimeFromString(value.(string))
					if err != nil {
						fmt.Printf("Warning: Failed to parse 'updated' date in '%s': %v\n", path, err)
					} else {
						article.Updated = updatedTime
					}
				} else if t, ok := value.(time.Time); ok {
					article.Updated = t
				}
			case "coverimagepath":
				article.CoverImagePath = value.(string)
			case "url":
				article.Url = value.(string)
			case "tags":
				switch reflect.TypeOf(value).Kind() {

				case reflect.Slice:
					tags := value.([]any)
					for _, tag := range tags {
						tagString := strings.Trim(tag.(string), " ")
						article.Tags = append(article.Tags, tagString)
					}

				case reflect.String:
					tags := strings.ReplaceAll(value.(string), ";", ",")
					tagsArray := strings.Split(tags, ",")
					for i, tag := range tagsArray {
						tag = strings.Trim(tag, " ")
						tagsArray[i] = tag
					}
					article.Tags = tagsArray
				}
			}
		}
	}

	// Set Created and Updated to file dates if not provided in frontmatter.
	fileInfo, err := os.Stat(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to get file info for '%s': %w", path, err)
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

	// Default title to filename if not provided.
	if article.Title == "" {
		article.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return article, nil
}

func genRelativeLink(linkToSelf string, name string) string {
	linkToSelf = strings.ToLower(linkToSelf)
	linkToSelf = strings.ReplaceAll(linkToSelf, "http://", "")
	linkToSelf = strings.ReplaceAll(linkToSelf, "https://", "")
	linkToSelf = strings.ReplaceAll(linkToSelf, "\\", "/")
	// Split the path by the OS specific separator

	parts := strings.Split(linkToSelf, "/")
	// // Handle files on the root directory
	// if len(parts) <= 1 {
	// 	return "style.css"
	// }

	// // Calculate the number of ".." needed
	// upDirCount := len(parts) - 1
	// if parts[len(parts)-1] == "index.html" {
	// 	upDirCount = len(parts) - 2
	// }

	// Create the ".." string
	upDir := strings.Repeat("../", len(parts)-1)
	// Combine ".." with "style.css"
	return upDir + name
}

// FormatMarkdown applies an HTML template to the Markdown content of an article.
// Returns an error if template parsing or execution fails.
func FormatMarkdown(article *Article, settings Settings, assets fs.FS) error {
	funcMap := template.FuncMap{
		"genRelativeLink": genRelativeLink,
		"stringsJoin":     strings.Join,
		"gen_share_url":   gen_share_url,
		"slicesContains":  slices.Contains[[]string],
	}

	htmlArticleTemplate, err := texttemplate.New("html-article.gohtml").Funcs(funcMap).ParseFS(assets, "assets/templates/html-article.gohtml")
	if err != nil {
		return fmt.Errorf("error parsing index template: %w", err)
	}

	// Execute the template with article data and settings.
	var tp bytes.Buffer
	err = htmlArticleTemplate.Execute(&tp, struct {
		Art      Article
		Ctt      template.HTML
		Settings Settings
	}{*article, template.HTML(article.HtmlContent), settings})
	if err != nil {
		return fmt.Errorf("error executing markdown article template: %w", err)
	}
	article.HtmlContent = tp.String()
	return nil
}

// HTMLFile parses an HTML file, extracts metadata from tags, and populates an Article struct.
// Returns the parsed Article and an error if reading or parsing fails.
func HTMLFile(path string) (Article, error) {
	// Read the HTML file content.
	data, err := os.ReadFile(path)
	if err != nil {
		return Article{}, fmt.Errorf("failed to read HTML file '%s': %w", path, err)
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
		return Article{}, fmt.Errorf("failed to parse HTML content of '%s': %w", path, err)
	}

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
					fmt.Printf("Warning: Failed to parse 'created' date from meta tag in '%s': %v\n", path, err)
				} else {
					article.Created = createdTime
				}
			case "updated":
				updatedTime, err := DateTimeFromString(val)
				if err != nil {
					fmt.Printf("Warning: Failed to parse 'updated' date from meta tag in '%s': %v\n", path, err)
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
		return Article{}, fmt.Errorf("failed to get file info for '%s': %w", path, err)
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

	return article, nil
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

// extractResources parses HTML content and extracts the values of "src" and "href" attributes
// from "img", "script", and "link" tags, returning a list of resource paths and an error if parsing fails.
func extractResources(htmlContent string) ([]string, error) {
	var resources []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML content for resource extraction: %w", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "img" || n.Data == "script" || n.Data == "link" {
				for _, attr := range n.Attr {
					if attr.Key == "src" || attr.Key == "href" {
						resources = append(resources, attr.Val)
						break // Assuming only one relevant attribute per tag
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return resources, nil
}

func gen_share_url(article Article, settings Settings, service string) string {
	var hashtags []string
	for _, tag := range article.Tags {
		hashtags = append(hashtags, strings.ReplaceAll(tag, " ", ""))
	}
	var blueskyHashTags []string
	for _, tag := range hashtags {
		blueskyHashTags = append(blueskyHashTags, fmt.Sprintf("#%s", tag))
	}
	articleURL := fmt.Sprintf("%s/%s", settings.BaseUrl, article.LinkToSelf)
	switch service {
	case "x":
		return fmt.Sprintf("https://twitter.com/intent/tweet?url=%s&text=%s&hashtags=%s",
			url.QueryEscape(articleURL), url.QueryEscape(article.Description), strings.Join(hashtags, ","))
	case "bluesky":
		text := fmt.Sprintf("%s\n%s\n%s", article.Description, strings.Join(blueskyHashTags, " "), articleURL)
		return fmt.Sprintf("https://bsky.app/intent/compose?text=%s", url.QueryEscape(text))
	case "threads":
		text := fmt.Sprintf("%s\n%s\n%s", article.Description, strings.Join(blueskyHashTags, " "), articleURL)
		return fmt.Sprintf("https://www.threads.net/intent/post?text=%s", url.QueryEscape(text))
	case "mastodon":
		text := fmt.Sprintf("%s\n%s\n%s", article.Description, strings.Join(blueskyHashTags, " "), articleURL)
		return fmt.Sprintf("https://mastodon.social/?text=%s", url.QueryEscape(text))
	case "telegram":
		return fmt.Sprintf("https://t.me/share/url?url=%s&text=%s", url.QueryEscape(articleURL), url.QueryEscape(article.Description))
	case "reddit":
		return fmt.Sprintf("https://www.reddit.com/submit?url=%s&title=%s", url.QueryEscape(articleURL), url.QueryEscape(article.Title))
	case "linkedin":
		return fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?url=%s", url.QueryEscape(articleURL))
	case "hackernews":
		postURL := articleURL
		if article.Url != "" {
			postURL = article.Url
		}
		return fmt.Sprintf("https://news.ycombinator.com/submitlink?u=%s&t=%s", url.QueryEscape(postURL), url.QueryEscape(article.Title))
	// case "facebook":
	// 	return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", url.QueryEscape(article.LinkToSelf))
	default:
		return ""
	}
}

// getThemeData returns a parse.Theme struct populated with style settings for the given theme.
// This function defines the visual appearance of the website based on predefined theme options
// (default, dark, colorful), setting colors, fonts, and other style-related variables.
func GetThemeData(style Style) Theme {
	switch style {
	case Dark:
		return Theme{
			Dark:           true,
			HeaderFont:     "\"Georgia\"",
			BodyFont:       "\"Garamond\"",
			Background:     "rgb(69, 69, 69)",
			Text:           "rgb(216, 216, 216)",
			Card:           "rgb(85, 85, 91)",
			Link:           "rgb(255, 75, 75)",
			Shadow:         "rgba(0, 0, 0, 0.777)",
			FontSize:       1.0,
			HeaderFontSize: 1.0,
			BodyFontSize:   1.0,
		}
	case Clean:
		return Theme{
			Dark:           true,
			HeaderFont:     "Times New Roman, serif",
			BodyFont:       "Verdana , sans-serif",
			Background:     "rgb(20, 20, 20)",
			Text:           "rgb(216, 216, 216)",
			Card:           "rgb(50, 50, 56)",
			Link:           "rgb(255, 75, 75)",
			Shadow:         "rgba(0, 0, 0, 0.777)",
			FontSize:       1.0,
			HeaderFontSize: 1.0,
			BodyFontSize:   0.8,
		}
	case Colorful:
		return Theme{
			Dark:           false,
			HeaderFont:     "'Georgia', 'Times New Roman', Times, serif",
			BodyFont:       "'Raleway', sans-serif",
			Background:     "rgb(100, 100, 100)",
			Text:           "rgb(0, 0, 0)",
			Card:           "rgba(80, 212, 89, 0.65)",
			Button:         "rgb(230, 91, 91)",
			Link:           "rgb(21, 89, 138)",
			Shadow:         "rgba(98, 0, 0, 0.777)",
			FontSize:       1.0,
			HeaderFontSize: 1.0,
			BodyFontSize:   1.0,
		}
	default: // Default style.
		return Theme{
			Dark:           false,
			HeaderFont:     "\"Georgia\"",
			BodyFont:       "\"Garamond\"",
			Background:     "rgb(234, 234, 234)",
			Text:           "rgb(85, 85, 85)",
			Card:           "rgb(237, 237, 237)",
			Link:           "rgb(201, 38, 38)",
			Shadow:         "rgba(0, 0, 0, 0.25)",
			FontSize:       1.0,
			HeaderFontSize: 1.0,
			BodyFontSize:   1.0,
		}
	}
}
