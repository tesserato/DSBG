package parse

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// regexPatterns defines a list of regular expression patterns to identify dates in strings.
var regexPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?P<year>\d{4})\D+(?P<month>\d{1,2})\D+(?P<day>\d{1,2})`),
	regexp.MustCompile(`(?P<day>\d{1,2})\D+(?P<month>\d{1,2})\D+(?P<year>\d{4})`),
	regexp.MustCompile(`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`),
}

var themesPath = "src/assets/themes"

// regexColorScheme finds the standard CSS color-scheme property (e.g., color-scheme: dark;).
var regexColorScheme = regexp.MustCompile(`(?i)color-scheme\s*:\s*([^;]+);`)

// regexHashtagCleanup matches characters that are NOT letters, numbers, or underscores.
// Used to sanitize tags for hashtag generation (e.g. "C++" -> "C", "Go Lang" -> "GoLang").
var regexHashtagCleanup = regexp.MustCompile(`[^\p{L}\p{N}_]+`)

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
	// Pre-process specific programming symbols before stripping to prevent
	// collisions (e.g. "C#" vs "C++" vs "C")
	url = strings.ReplaceAll(url, "#", "sharp")
	url = strings.ReplaceAll(url, "+", "plus")

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

// copyDirectoryRecursively copies all contents of srcDir to destDir.
func copyDirectoryRecursively(srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Calculate the relative path from the source directory
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read the source file
		input, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file '%s': %w", path, err)
		}

		// Write to the destination file
		if err := os.WriteFile(destPath, input, 0644); err != nil {
			return fmt.Errorf("error writing file '%s': %w", destPath, err)
		}
		return nil
	})
}

// CopyHtmlResources copies associated resources for an article and determines
// the article's output path. Resources include images and other linked assets.
func CopyHtmlResources(settings Settings, article *Article, resources []string) error {
	relativeInputPath, err := filepath.Rel(settings.InputPath, article.OriginalPath)
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

	outputPath := filepath.Join(settings.OutputPath, relativeInputPath)
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

	// SPECIAL CASE: If this is an HTML file marked as "PAGE", copy the entire directory contents.
	// This ensures complex HTML pages with relative dependencies (js, css, media) are preserved.
	if strings.HasSuffix(strings.ToLower(article.OriginalPath), ".html") && slices.Contains(article.Tags, "PAGE") {
		if err := copyDirectoryRecursively(originalDirectory, outputDirectory); err != nil {
			if !settings.IgnoreErrors {
				return fmt.Errorf("failed to copy directory for HTML PAGE '%s': %w", article.Title, err)
			}
			log.Printf("Warning: Failed to copy directory for HTML PAGE '%s': %v", article.Title, err)
		}
		// We still fall through to handle cover image metadata updates below,
		// but we skip the manual resource copy loop since we just copied everything.
	} else {
		// STANDARD CASE: Extract and copy specific resources.
		for _, resourceOrigRelPath := range resources {
			resourceOrigRelPath = strings.TrimSpace(resourceOrigRelPath)
			if resourceOrigRelPath == "" {
				continue
			}

			resourceOrigRelPathLower := strings.ToLower(resourceOrigRelPath)

			// 1. Skip absolute external URLs (http/https/ftp)
			if strings.HasPrefix(resourceOrigRelPathLower, "http://") ||
				strings.HasPrefix(resourceOrigRelPathLower, "https://") ||
				strings.HasPrefix(resourceOrigRelPathLower, "ftp://") ||
				strings.HasPrefix(resourceOrigRelPathLower, "//") {
				continue
			}

			// 2. Skip non-resource links (anchors, mailto, tel, sms, www.)
			if strings.HasPrefix(resourceOrigRelPathLower, "#") ||
				strings.HasPrefix(resourceOrigRelPathLower, "mailto:") ||
				strings.HasPrefix(resourceOrigRelPathLower, "tel:") ||
				strings.HasPrefix(resourceOrigRelPathLower, "sms:") ||
				strings.HasPrefix(resourceOrigRelPathLower, "www.") {
				continue
			}

			// Strip query string and fragment, and normalize root-relative paths
			cleanPath := resourceOrigRelPath
			if u, err := url.Parse(resourceOrigRelPath); err == nil {
				if u.Path != "" {
					cleanPath = u.Path
				} else {
					// If path is empty (e.g. "?query" or "#anchor"), skip it
					continue
				}
			}

			// Treat leading "/" as project-root-relative, not filesystem-root
			cleanPath = strings.TrimPrefix(cleanPath, "/")
			if cleanPath == "" {
				continue
			}

			// 3. Skip internal navigation to other source files (.md, .html)
			// These are likely links to other posts, not assets to be copied raw.
			ext := strings.ToLower(filepath.Ext(cleanPath))
			if ext == ".md" || ext == ".markdown" || ext == ".html" || ext == ".htm" {
				continue
			}

			// 4. Skip links without extensions (likely navigation links e.g., [About](about))
			// Unless they are explicit file references, we assume they are internal routes.
			if ext == "" {
				continue
			}

			resourceOrigPath := filepath.Join(originalDirectory, cleanPath)
			resourceDestPath := filepath.Join(outputDirectory, cleanPath)

			// Check if resource exists before reading
			stat, err := os.Stat(resourceOrigPath)
			if err != nil {
				if os.IsNotExist(err) {
					// If strictly missing a file with an extension (e.g. image.png), fail/warn.
					if !settings.IgnoreErrors {
						return fmt.Errorf("resource file '%s' not found (referenced in '%s')", resourceOrigPath, article.Title)
					}
					log.Printf("Warning: Resource file '%s' not found (referenced in '%s')", resourceOrigPath, article.Title)
					continue
				}
				// Other error
				if !settings.IgnoreErrors {
					return fmt.Errorf("failed to stat resource file '%s': %w", resourceOrigPath, err)
				}
				continue
			}

			// Skip if it is a directory
			if stat.IsDir() {
				continue
			}

			input, err := os.ReadFile(resourceOrigPath)
			if err != nil {
				if !settings.IgnoreErrors {
					return fmt.Errorf("failed to read resource file '%s': %w", resourceOrigPath, err)
				}
				log.Printf("Warning: Failed to read resource file '%s': %v", resourceOrigPath, err)
				continue
			}

			if err := os.MkdirAll(filepath.Dir(filepath.FromSlash(resourceDestPath)), 0755); err != nil {
				return fmt.Errorf("failed to create directory for resource '%s': %w", resourceDestPath, err)
			}

			if err := os.WriteFile(resourceDestPath, input, 0644); err != nil {
				return fmt.Errorf("failed to write resource file to '%s': %w", resourceDestPath, err)
			}
		}
	}

	linkToSelf, err := filepath.Rel(settings.OutputPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to get relative link from '%s' to '%s': %w", settings.OutputPath, outputPath, err)
	}
	article.LinkToSelf = filepath.ToSlash(linkToSelf)
	article.LinkToSave = filepath.ToSlash(outputPath)

	// If a cover image was specified, normalize its path to be relative to the article's output
	// location so templates and Open Graph tags can reference it consistently.
	// Only update if it was a local path.
	originalCoverRel := article.CoverImage
	if originalCoverRel != "" && !strings.HasPrefix(strings.ToLower(originalCoverRel), "http") {
		// Even if we copied the whole folder, we ensure the cover image path logic is consistent.
		// If we DIDN'T copy the whole folder (standard case), we must copy the cover image specifically.
		isHtmlPage := strings.HasSuffix(strings.ToLower(article.OriginalPath), ".html") && slices.Contains(article.Tags, "PAGE")

		if !isHtmlPage {
			coverImageOrigPath := filepath.Join(originalDirectory, originalCoverRel)
			coverImageArticleDestPath := filepath.Join(outputDirectory, originalCoverRel)

			file, err := os.ReadFile(coverImageOrigPath)
			if err != nil {
				if !settings.IgnoreErrors {
					return fmt.Errorf("failed to read cover image '%s' for article '%s': %w", coverImageOrigPath, article.Title, err)
				}
				log.Printf("Warning: Could not read cover image '%s' for article '%s': %v", coverImageOrigPath, article.Title, err)
			} else {
				if err := os.MkdirAll(filepath.Dir(coverImageArticleDestPath), 0755); err != nil {
					return fmt.Errorf("error creating directory for cover image '%s': %w", coverImageArticleDestPath, err)
				}
				if err := os.WriteFile(coverImageArticleDestPath, file, 0644); err != nil {
					return fmt.Errorf("error writing cover image file '%s': %w", coverImageArticleDestPath, err)
				}
			}
		}

		coverRootRel := filepath.Join(filepath.Dir(article.LinkToSelf), originalCoverRel)
		article.CoverImage = filepath.ToSlash(coverRootRel)
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

// ExtractResources traverses an HTML node tree and extracts src/href/data values
// from tags like img, script, link, video, audio, source, track, object, iframe, and a.
func ExtractResources(n *html.Node) []string {
	var resources []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Define tags and the attributes we want to extract from them.
			// This now includes standard media tags and generic links/anchors.
			targetAttrs := map[string][]string{
				"img":    {"src"},
				"script": {"src"},
				"link":   {"href"},
				"video":  {"src", "poster"},
				"audio":  {"src"},
				"source": {"src"},
				"track":  {"src"},
				"object": {"data"},
				"iframe": {"src"},
				"embed":  {"src"},
				"a":      {"href"},
			}

			if attrsToScan, found := targetAttrs[n.Data]; found {
				for _, attr := range n.Attr {
					for _, target := range attrsToScan {
						if attr.Key == target {
							resources = append(resources, attr.Val)
						}
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

// toAbsoluteUrl handles logic to prevent double-concatenation of BaseURL
// It is used by BuildShareUrl and registered as "absURL" in templates.
func toAbsoluteUrl(urlStr string, baseUrl string) string {
	if urlStr == "" {
		return ""
	}
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		return urlStr
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseUrl, "/"), strings.TrimPrefix(urlStr, "/"))
}

// BuildShareUrl replaces placeholders in the urlTemplate with encoded article data,
// and returns the final URL as a template.URL value.
func BuildShareUrl(urlTemplate string, article Article, settings Settings) template.URL {
	// 1. Determine the "Official" URL of the post.
	finalUrl := article.ExternalLink
	if finalUrl == "" {
		finalUrl = toAbsoluteUrl(article.LinkToSelf, settings.BaseUrl)
	}

	// 2. Determine the "Target Link" ({LINK}).
	// This uses the 'ExternalLink' if present.
	// If not, it tries to find the first link inside the article body (smart link-blog behavior).
	// If neither, it remains empty.
	targetLink := article.ExternalLink
	if targetLink == "" {
		targetLink = extractFirstLink(article.HtmlContent)
	}

	// 3. Determine the Image URL ({IMAGE}).
	imageUrl := toAbsoluteUrl(article.CoverImage, settings.BaseUrl)

	// 4. Build Hashtags List ({TAGS}) and First Tag ({TAG})
	// We strip non-alphanumeric chars (except underscore) to create valid hashtags for most platforms.
	var cleanedTags []string
	var cleanedFirstTag string

	for _, tag := range article.Tags {
		// e.g. "C++" -> "C", "Web Dev" -> "WebDev"
		clean := regexHashtagCleanup.ReplaceAllString(tag, "")
		if clean != "" {
			cleanedTags = append(cleanedTags, "#"+clean)
			if cleanedFirstTag == "" {
				cleanedFirstTag = clean
			}
		}
	}

	tagsString := strings.Join(cleanedTags, " ")

	encodedUrl := encodeComponent(finalUrl)
	encodedTitle := encodeComponent(article.Title)
	encodedDesc := encodeComponent(article.Description)
	encodedText := encodeComponent(article.TextContent)
	encodedTargetLink := encodeComponent(targetLink)
	encodedImage := encodeComponent(imageUrl)
	encodedTags := encodeComponent(tagsString)
	encodedFirstTag := encodeComponent(cleanedFirstTag)

	result := strings.ReplaceAll(urlTemplate, "{URL}", encodedUrl)
	result = strings.ReplaceAll(result, "{TITLE}", encodedTitle)
	result = strings.ReplaceAll(result, "{DESCRIPTION}", encodedDesc)
	result = strings.ReplaceAll(result, "{TEXT}", encodedText)
	result = strings.ReplaceAll(result, "{LINK}", encodedTargetLink)
	result = strings.ReplaceAll(result, "{IMAGE}", encodedImage)
	// Replace {TAGS} -> space-separated list of hash-prefixed tags
	result = strings.ReplaceAll(result, "{TAGS}", encodedTags)
	// Replace {TAG} -> just the text of the first tag (no hash), good for specific category params
	result = strings.ReplaceAll(result, "{TAG}", encodedFirstTag)

	return template.URL(result)
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

// SaveThemeCSS copies the selected theme CSS file from embedded assets to style.css in the output directory.
// If themeName is empty or invalid, it attempts to use "default.css".
func SaveThemeCSS(assets fs.FS, themeName string, outputDirectory string, ignoreErrors bool) error {
	if themeName == "" {
		themeName = "default"
	}

	themeFile := themeName + ".css"
	srcPath := path.Join(themesPath, themeFile)

	fileContent, err := fs.ReadFile(assets, srcPath)
	if err != nil {
		available, _ := GetAvailableThemes(assets)
		if !ignoreErrors {
			return fmt.Errorf("theme '%s' not found (Available: %s)", themeName, strings.Join(available, ", "))
		}
		log.Printf("Warning: Theme '%s' not found (Available: %s). Falling back to default theme.", themeName, strings.Join(available, ", "))

		// Fallback to default
		srcPath = path.Join(themesPath, "default.css")
		fileContent, err = fs.ReadFile(assets, srcPath)
		if err != nil {
			return fmt.Errorf("failed to load default theme CSS: %w", err)
		}
	} else {
		log.Printf("Using theme: %s", themeName)
	}

	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDirectory, err)
	}

	destPath := filepath.Join(outputDirectory, "style.css")
	if err := os.WriteFile(destPath, fileContent, 0644); err != nil {
		return fmt.Errorf("error writing style.css: %w", err)
	}
	return nil
}

// GetAvailableThemes scans the embedded assets for available CSS themes.
// It returns a sorted list of theme names (filenames without extension).
func GetAvailableThemes(assets fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(assets, themesPath)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
			name := strings.TrimSuffix(entry.Name(), ".css")
			themes = append(themes, name)
		}
	}
	slices.Sort(themes)
	return themes, nil
}

// GetThemeType determines if a theme is "light" or "dark" by inspecting the CSS file.
func GetThemeType(assets fs.FS, themeName string) string {
	themeFile := themeName + ".css"
	srcPath := path.Join(themesPath, themeFile)

	content, err := fs.ReadFile(assets, srcPath)
	if err != nil {
		return "dark"
	}

	match := regexColorScheme.FindStringSubmatch(string(content))
	if len(match) > 1 {
		val := strings.ToLower(strings.TrimSpace(match[1]))
		if strings.Contains(val, "dark") {
			return "dark"
		}
		return "light"
	}
	return "dark"
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

// ArticleSchemaType determines which schema.org type to use for an article.
func ArticleSchemaType(a Article) string {
	for _, tag := range a.Tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "news" || t == "article" {
			return "NewsArticle"
		}
	}
	return "BlogPosting"
}
