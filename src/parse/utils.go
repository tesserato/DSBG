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

// GetPaths retrieves all file paths within a directory and its subdirectories matching extensions.
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

// cleanString cleans url/path strings.
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

// CopyHtmlResources copies associated resources for an article and determines output path.
func CopyHtmlResources(settings Settings, article *Article) error {
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
	err = os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDirectory, err)
	}

	originalDirectory := filepath.Dir(article.OriginalPath)

	if article.CoverImagePath != "" {
		coverImageOrigPath := filepath.Join(originalDirectory, article.CoverImagePath)
		coverImageDestPath := filepath.Join(settings.OutputDirectory, article.CoverImagePath)

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

	LinkToSelf, err := filepath.Rel(settings.OutputDirectory, outputPath)
	if err != nil {
		return fmt.Errorf("failed to get relative link from '%s' to '%s': %w", settings.OutputDirectory, outputPath, err)
	}
	article.LinkToSelf = filepath.ToSlash(LinkToSelf)
	article.LinkToSave = filepath.ToSlash(outputPath)
	return nil
}

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

// extractFirstLink parses HTML content and returns the value of the first "href" attribute
// found in an anchor "a" tag. Returns an empty string if no link is found.
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

// encodeComponent encodes a string for use in a URL query, then replaces `+` with `%20`.
func encodeComponent(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// BuildShareUrl replaces placeholders in the urlTemplate with encoded article data.
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

// CleanContent prepares text content for Fuse.js search indexing.
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

// IsImage checks if a string (filename) has a common image extension.
// Exported so it can be used in main.go and templates.
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

// ApplyCSSTemplate applies the selected theme's styles to the default CSS template.
func ApplyCSSTemplate(themeData Theme, outputDirectory string, tmpl *texttemplate.Template) error {
	var output strings.Builder
	// Execute the CSS template with the theme data to generate the final CSS content.
	err := tmpl.Execute(&output, themeData)
	if err != nil {
		return fmt.Errorf("error executing style template: %w", err)
	}

	pathToSave := filepath.Join(outputDirectory, "style.css")
	// Write the processed CSS content to the 'style.css' file in the output directory.
	if err := os.WriteFile(pathToSave, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("error saving processed css file: %w", err)
	}
	return nil
}
