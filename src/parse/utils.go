package parse

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"

	"golang.org/x/net/html"
)

// regexPatterns defines a list of regular expression patterns to identify dates in strings.
var regexPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?P<year>\d{4})\D+(?P<month>\d{1,2})\D+(?P<day>\d{1,2})`),
	regexp.MustCompile(`(?P<day>\d{1,2})\D+(?P<month>\d{1,2})\D+(?P<year>\d{4})`),
	regexp.MustCompile(`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`),
}

// RemoveDateFromPath attempts to remove date patterns from a given string.
func RemoveDateFromPath(stringWithDate string) string {
	for _, regexPattern := range regexPatterns {
		stringWithDate = regexPattern.ReplaceAllString(stringWithDate, "")
	}
	stringWithDate = strings.Trim(stringWithDate, "-_ ")
	return stringWithDate
}

// DateTimeFromString attempts to parse a date and time from a string using predefined regex patterns.
func DateTimeFromString(date string) (time.Time, error) {
	m := make(map[string]int)
	foundMatch := false
	for _, pattern := range regexPatterns {
		matches := pattern.FindStringSubmatch(date)
		if len(matches) > 0 {
			foundMatch = true
			for i, name := range pattern.SubexpNames()[1:] {
				if name == "" {
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
// matching the provided list of file extensions.
func GetPaths(root string, extensions []string) ([]string, error) {
	var files []string
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extLower := strings.ToLower(strings.TrimSpace(ext))
		extMap[extLower] = true
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if extMap[ext] {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// cleanString normalizes path-like strings by removing non-alphanumeric characters
// and redundant separators, making them safe to use as URL fragments.
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

// CopyHtmlResources copies associated resources for an article and determines
// the article's output path. Resources include images and other linked assets.
//
// resources is a list of relative file paths found in the article (images, scripts, etc).
// Cover images are copied to live next to the article's index file.
func CopyHtmlResources(settings Settings, article *Article, resources []string) error {
	relativeInputPath, err := filepath.Rel(settings.InputDirectory, article.OriginalPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path for '%s': %w", article.OriginalPath, err)
	}

	if !settings.DoNotRemoveDateFromTitles {
		datelessTitle := RemoveDateFromPath(article.Title)
		if datelessTitle != "" {
			article.Title = datelessTitle
		}
	}

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

	outputPath := filepath.Join(settings.OutputDirectory, relativeInputPath)
	outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	outputPath = filepath.Join(outputPath, settings.IndexName)

	if !settings.DoNotRemoveDateFromPaths {
		datelessOutputPath := RemoveDateFromPath(outputPath)
		if !(strings.Contains(datelessOutputPath, "\\") || strings.Contains(datelessOutputPath, "//")) {
			outputPath = datelessOutputPath
		}
	}
	outputPath = cleanString(outputPath)
	outputDirectory := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDirectory, err)
	}

	originalDirectory := filepath.Dir(article.OriginalPath)
	originalCoverRel := article.CoverImagePath

	// Copy cover image so it lives next to the article's index (within outputDirectory).
	if originalCoverRel != "" {
		coverImageOrigPath := filepath.Join(originalDirectory, originalCoverRel)
		coverImageArticleDestPath := filepath.Join(outputDirectory, originalCoverRel)

		if !slices.Contains(article.Tags, "PAGE") {
			file, err := os.ReadFile(coverImageOrigPath)
			if err != nil {
				return fmt.Errorf("error reading cover image '%s': %w", coverImageOrigPath, err)
			}
			if err := os.MkdirAll(filepath.Dir(coverImageArticleDestPath), 0755); err != nil {
				return fmt.Errorf("error creating directory for cover image '%s': %w", coverImageArticleDestPath, err)
			}
			if err := os.WriteFile(coverImageArticleDestPath, file, 0644); err != nil {
				return fmt.Errorf("error writing cover image file '%s': %w", coverImageArticleDestPath, err)
			}
		}
	}

	for _, resourceOrigRelPath := range resources {
		resourceOrigRelPathLower := strings.ToLower(resourceOrigRelPath)
		if strings.Contains(resourceOrigRelPathLower, "http") {
			continue
		}
		resourceOrigPath := filepath.Join(originalDirectory, resourceOrigRelPath)
		resourceDestPath := filepath.Join(outputDirectory, resourceOrigRelPath)

		input, err := os.ReadFile(resourceOrigPath)
		if err != nil {
			return fmt.Errorf("failed to read resource file '%s': %w", resourceOrigPath, err)
		}

		if err := os.MkdirAll(filepath.Dir(filepath.FromSlash(resourceDestPath)), 0755); err != nil {
			return fmt.Errorf("failed to create directory for resource '%s': %w", resourceDestPath, err)
		}

		if err := os.WriteFile(resourceDestPath, input, 0644); err != nil {
			return fmt.Errorf("failed to write resource file to '%s': %w", resourceDestPath, err)
		}
	}

	// Compute LinkToSelf and LinkToSave.
	linkToSelf, err := filepath.Rel(settings.OutputDirectory, outputPath)
	if err != nil {
		return fmt.Errorf("failed to get relative link from '%s' to '%s': %w", settings.OutputDirectory, outputPath, err)
	}
	article.LinkToSelf = filepath.ToSlash(linkToSelf)
	article.LinkToSave = filepath.ToSlash(outputPath)

	// If a cover image was specified, normalize its path to be relative to the article's output
	// location so templates and Open Graph tags can reference it consistently.
	if originalCoverRel != "" && !slices.Contains(article.Tags, "PAGE") {
		coverRootRel := filepath.Join(filepath.Dir(article.LinkToSelf), originalCoverRel)
		article.CoverImagePath = filepath.ToSlash(coverRootRel)
	}

	return nil
}

// genRelativeLink computes a relative link from linkToSelf to name, unless name is absolute.
func genRelativeLink(linkToSelf string, name string) string {
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		return name
	}

	linkToSelf = strings.ToLower(linkToSelf)
	linkToSelf = strings.ReplaceAll(linkToSelf, "http://", "")
	linkToSelf = strings.ReplaceAll(linkToSelf, "https://", "")
	linkToSelf = strings.ReplaceAll(linkToSelf, "\\", "/")
	parts := strings.Split(linkToSelf, "/")
	upDir := strings.Repeat("../", len(parts)-1)
	return upDir + name
}

// ExtractResources traverses an HTML node tree and extracts src/href values
// from img, script, and link tags.
func ExtractResources(n *html.Node) []string {
	var resources []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "img" || n.Data == "script" || n.Data == "link" {
				for _, attr := range n.Attr {
					if attr.Key == "src" || attr.Key == "href" {
						resources = append(resources, attr.Val)
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return resources
}

// extractFirstLink parses HTML content and returns the value of the first "href" attribute
// found in an anchor "a" tag. It returns an empty string if no link is found.
func extractFirstLink(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var link string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if link != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link = attr.Val
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return link
}

// encodeComponent encodes a string for use in a URL query, replacing '+' with '%20'.
func encodeComponent(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// BuildShareUrl replaces placeholders in the urlTemplate with encoded article data,
// and returns the final URL as a template.URL value.
func BuildShareUrl(urlTemplate string, article Article, settings Settings) template.URL {
	finalUrl := article.Url
	if finalUrl == "" {
		finalUrl = fmt.Sprintf("%s/%s", strings.TrimSuffix(settings.BaseUrl, "/"), strings.TrimPrefix(article.LinkToSelf, "/"))
	}

	linkUrl := article.Url
	if linkUrl == "" {
		linkUrl = extractFirstLink(article.HtmlContent)
	}

	encodedUrl := encodeComponent(finalUrl)
	encodedTitle := encodeComponent(article.Title)
	encodedDesc := encodeComponent(article.Description)
	encodedText := encodeComponent(article.TextContent)
	encodedLinkUrl := encodeComponent(linkUrl)

	result := strings.ReplaceAll(urlTemplate, "{URL}", encodedUrl)
	result = strings.ReplaceAll(result, "{TITLE}", encodedTitle)
	result = strings.ReplaceAll(result, "{DESCRIPTION}", encodedDesc)
	result = strings.ReplaceAll(result, "{TEXT}", encodedText)
	result = strings.ReplaceAll(result, "{LINK-URL}", encodedLinkUrl)

	return template.URL(result)
}

// CleanContent normalizes text content into a slice of tokens for search indexing.
func CleanContent(s string) []string {
	replacements := map[string]string{
		"’": "'",
		"–": " ",
	}
	removals := []string{
		"\n", "\r", "\t", "(", ")", "[", "]", "{", "}",
		"\"", "\\", "/", "”", "#", "-", "*",
	}
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}
	for _, char := range removals {
		s = strings.ReplaceAll(s, char, " ")
	}
	return strings.Fields(s)
}

// IsImage reports whether a filename has a common image extension.
// It is exported so it can be used by main and templates.
func IsImage(s string) bool {
	s = strings.ToLower(s)
	exts := []string{".svg", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".bmp", ".tiff"}
	for _, e := range exts {
		if strings.HasSuffix(s, e) {
			return true
		}
	}
	return false
}

// GetThemeData returns theme configuration for a given Style.
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
	default:
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

// ApplyCSSTemplate applies the selected theme's styles to the default CSS template
// and writes the generated CSS to style.css in the output directory.
func ApplyCSSTemplate(themeData Theme, outputDirectory string, tmpl *texttemplate.Template) error {
	var output strings.Builder
	if err := tmpl.Execute(&output, themeData); err != nil {
		return fmt.Errorf("error executing style template: %w", err)
	}

	pathToSave := filepath.Join(outputDirectory, "style.css")
	if err := os.WriteFile(pathToSave, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("error saving processed css file: %w", err)
	}
	return nil
}

// ParseSortOrder converts a string into a SortOrder, validating supported options.
func ParseSortOrder(s string) (SortOrder, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch SortOrder(s) {
	case SortDateCreated,
		SortReverseDateCreated,
		SortDateUpdated,
		SortReverseDateUpdated,
		SortTitle,
		SortReverseTitle,
		SortPath,
		SortReversePath:
		return SortOrder(s), nil
	default:
		return "", fmt.Errorf("unsupported sort order: %s", s)
	}
}

// ParseStyle converts a string representation into a Style constant.
func ParseStyle(s string) (Style, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "", "default":
		return Default, nil
	case "dark":
		return Dark, nil
	case "clean":
		return Clean, nil
	case "colorful":
		return Colorful, nil
	default:
		return Default, fmt.Errorf("unknown style %q", s)
	}
}

// ArticleSchemaType determines which schema.org type to use for an article.
// If any tag looks like "news" or "article", it returns "NewsArticle";
// otherwise it returns "BlogPosting".
func ArticleSchemaType(a Article) string {
	for _, tag := range a.Tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "news" || t == "article" {
			return "NewsArticle"
		}
	}
	return "BlogPosting"
}
