package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tesserato/DSBG/src/parse"
)

//go:embed src/assets
var assets embed.FS

// shareButtonsFlag is a custom flag type to handle repeated --share flags
type shareButtonsFlag []parse.ShareButton

func (s *shareButtonsFlag) String() string {
	return "Share buttons defined by Name:UrlTemplate"
}

func (s *shareButtonsFlag) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid share format, expected 'Name:URL', got '%s'", value)
	}
	*s = append(*s, parse.ShareButton{
		Name:        parts[0],
		UrlTemplate: parts[1],
	})
	return nil
}

// noFlagsPassed checks if any command-line flags were provided for a given FlagSet.
// It iterates through the flags visited by the FlagSet and returns false if any are found, true otherwise.
func noFlagsPassed(fs *flag.FlagSet) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		found = true
	})
	return !found
}

// logFlag is a helper function to log the details of a flag (name, default value, and usage) to stderr.
// It's used by the Usage function to print flag information in a formatted way.
func logFlag(f *flag.Flag) {
	defaultValue := f.DefValue
	if defaultValue != "" {
		defaultValue = fmt.Sprintf(" (default: %v)", defaultValue)
	}
	fmt.Fprintf(os.Stderr, "  -%v %v\n    %v \n", f.Name, defaultValue, f.Usage)
}

func main() {
	// Create FlagSets to manage flags for default blog generation and template creation modes.
	defaultFlagSet := flag.NewFlagSet("default", flag.ExitOnError)
	templateFlagSet := flag.NewFlagSet("template", flag.ExitOnError)

	var settings parse.Settings
	var shareButtons shareButtonsFlag

	// --- Default FlagSet Flags ---
	// Flags for controlling the blog generation process in default mode.
	defaultFlagSet.StringVar(&settings.Title, "title", "Blog", "The title of the blog, used in the header and page titles.")
	defaultFlagSet.StringVar(&settings.BaseUrl, "base-url", "", "The base URL of the blog (e.g., https://example.com), used for generating absolute URLs in the RSS feed and sitemap.")
	defaultFlagSet.StringVar(&settings.DescriptionMarkdown, "description", "This is my blog", "A short description of the blog, used in meta tags and the index page. Supports Markdown formatting.")
	defaultFlagSet.StringVar(&settings.InputDirectory, "input-path", "content", "Path to the directory containing source content files (Markdown or HTML).")
	defaultFlagSet.StringVar(&settings.OutputDirectory, "output-path", "public", "Path to the directory where the generated website files will be saved.")
	defaultFlagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Format for displaying dates on the website. Uses Go's time formatting (e.g., '2006-01-02').")
	defaultFlagSet.StringVar(&settings.IndexName, "index-name", "index.html", "Filename for the main index page.")
	defaultFlagSet.StringVar(&settings.PathToCustomCss, "css-path", "", "Path to a custom CSS file to override default styles.")
	defaultFlagSet.StringVar(&settings.PathToCustomJs, "js-path", "", "Path to a custom JavaScript file to add functionality to your site.")
	defaultFlagSet.StringVar(&settings.PathToCustomFavicon, "favicon-path", "", "Path to a custom favicon file (e.g., .ico, .png) for your site.")
	defaultFlagSet.BoolVar(&settings.DoNotExtractTagsFromPaths, "ignore-tags-from-paths", false, "Disable extracting tags from directory names.")
	defaultFlagSet.BoolVar(&settings.DoNotRemoveDateFromPaths, "keep-date-in-paths", false, "Do not remove date patterns (YYYY-MM-DD) from generated file paths.")
	defaultFlagSet.BoolVar(&settings.DoNotRemoveDateFromTitles, "keep-date-in-titles", false, "Do not remove date patterns (YYYY-MM-DD) from article titles.")
	defaultFlagSet.BoolVar(&settings.OpenInNewTab, "open-in-new-tab", false, "Open external links in new browser tabs.")

	// Custom share flag
	defaultFlagSet.Var(&shareButtons, "share", "Repeatable flag to add share buttons. Format: \"Name:UrlTemplate\".\n\tTemplate placeholders: {URL}, {TITLE}, {DESCRIPTION}, {TEXT}.")

	defaultFlagSet.StringVar(&settings.Sort, "sort", "date-created", "Sort order for articles on the index page. Possible values: date-created, reverse-date-created, date-updated, reverse-date-updated, title, reverse-title, path, reverse-path.")
	themeString := defaultFlagSet.String("theme", "default", "Predefined website style theme. Possible values: default, dark, clean, colorful.")
	pathToAdditionalElementsTop := defaultFlagSet.String("elements-top", "", "Path to HTML file to include at the top of each page's <head> (e.g., analytics).")
	pathToAdditionalElemensBottom := defaultFlagSet.String("elements-bottom", "", "Path to HTML file to include at the bottom of each page's <body> (e.g., scripts).")
	watch := defaultFlagSet.Bool("watch", false, "Enable watch mode: rebuild on changes and start local server.")

	// --- Template FlagSet Flags ---
	// Flags specific to the template creation mode.
	var templateSettings parse.TemplateSettings
	templateFlagSet.StringVar(&templateSettings.Title, "title", "", "Title to pre-fill in the template's 'title' field.")
	templateFlagSet.StringVar(&templateSettings.Description, "description", "", "Description to pre-fill in the template's 'description' field.")
	templateFlagSet.StringVar(&templateSettings.Created, "created", "", "Date to pre-fill in the template's 'created' field (format: 'date-format' flag).")
	templateFlagSet.StringVar(&templateSettings.Updated, "updated", "", "Date to pre-fill in the template's 'updated' field (format: 'date-format' flag).")
	templateFlagSet.StringVar(&templateSettings.CoverImagePath, "cover-image-path", "", "Path to a cover image (relative to 'output-path') to pre-fill in the template.")
	templateFlagSet.StringVar(&templateSettings.Tags, "tags", "", "Comma-separated tags to pre-fill in the template's 'tags' field.")
	templateFlagSet.StringVar(&templateSettings.OutputDirectory, "output-path", ".", "Directory to save the template file (defaults to current directory).")
	templateFlagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "(template mode) Date format used for pre-filling date fields in template.") // Date format flag also available in template mode, for consistency

	// Custom Usage function to display help for both default and template modes.
	defaultFlagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "DSBG (Dead Simple Blog Generator) - Generate static blogs from Markdown and HTML.\n")
		fmt.Fprintf(os.Stderr, "For complete documentation, visit: https://github.com/tesserato/dsbg\n\n")
		fmt.Fprintf(os.Stderr, "Usage: dsbg [flags] or dsbg -template [flags]\n\n") // Added usage instructions

		fmt.Fprintf(os.Stderr, "Default mode flags (dsbg [flags]):\n")
		defaultFlagSet.VisitAll(logFlag)

		fmt.Fprintf(os.Stderr, "\nTemplate mode flags (dsbg -template [flags]):\n") // Clarified template mode invocation
		templateFlagSet.VisitAll(logFlag)
	}

	if len(os.Args) <= 1 {
		defaultFlagSet.Usage()
		return
	}

	// Determine the mode of operation based on the first command-line argument.
	mode := strings.TrimPrefix(os.Args[1], "-")
	mode = strings.TrimPrefix(mode, "--")
	mode = strings.ToLower(mode)
	switch mode {
	case "template":
		log.Println("Running in template creation mode...")
		err := templateFlagSet.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Error parsing template flags: %v", err)
		}
		if err := createMarkdownTemplate(templateSettings); err != nil {
			log.Fatalf("Error creating markdown template: %v", err)
		}
		return // Exit after template creation
	default: // Default mode: blog generation
		err := defaultFlagSet.Parse(os.Args[1:])
		if err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}
		log.Println("Running in blog generation mode...")
	}

	settings.ShareButtons = shareButtons

	// Convert Markdown description to HTML (Do this *after* parsing flags so description is populated)
	var buf strings.Builder
	if err := parse.Markdown.Convert([]byte(settings.DescriptionMarkdown), &buf); err != nil {
		log.Fatalf("failed to convert description to HTML: %v", err)
	}
	settings.DescriptionHTML = template.HTML(buf.String())

	// Check if the input directory exists
	if _, err := os.Stat(settings.InputDirectory); os.IsNotExist(err) {
		if noFlagsPassed(defaultFlagSet) {
			defaultFlagSet.Usage()
			return
		}
		log.Fatalf("Input directory '%s' does not exist.", settings.InputDirectory)
	}

	// Read content of files specified by flags
	if *pathToAdditionalElementsTop != "" {
		content, err := os.ReadFile(*pathToAdditionalElementsTop)
		if err != nil {
			log.Fatalf("Error reading additional top elements file: %v", err)
		}
		settings.AdditionalElementsTop = template.HTML(content)
	}

	if *pathToAdditionalElemensBottom != "" {
		content, err := os.ReadFile(*pathToAdditionalElemensBottom)
		if err != nil {
			log.Fatalf("Error reading additional bottom elements file: %v", err)
		}
		settings.AdditionalElemensBottom = template.HTML(content)
	}

	// Ensure base URL is correctly formatted
	if settings.BaseUrl == "" {
		settings.BaseUrl = "http://localhost:666"
	} else {
		settings.BaseUrl = strings.TrimSuffix(settings.BaseUrl, "/")
	}

	// Set the website style: https://cdnjs.com/libraries/highlight.js
	switch *themeString {
	case "default":
		settings.Theme = parse.Default
		settings.HighlightTheme = "stackoverflow-light"
	case "dark":
		settings.Theme = parse.Dark
		settings.HighlightTheme = "github-dark-dimmed"
	case "clean":
		settings.Theme = parse.Clean
		settings.HighlightTheme = "github-dark-dimmed"
	case "colorful":
		settings.Theme = parse.Colorful
		settings.HighlightTheme = "github-dark-dimmed"
	default:
		settings.Theme = parse.Default
		settings.HighlightTheme = "stackoverflow-light"
		log.Printf("Unknown style '%s', using default.\n", *themeString)
	}

	// Perform the initial website build
	buildWebsite(settings)

	if *watch {
		startWatcher(settings)
	}

}

// createMarkdownTemplate generates a Markdown template file with predefined frontmatter.
// It uses the provided TemplateSettings to pre-fill fields in the frontmatter and saves the template to a file.
func createMarkdownTemplate(templateSettings parse.TemplateSettings) error {

	tmpl, err := template.New("md-article.gomd").ParseFS(assets, "src/assets/templates/md-article.gomd")
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	// Format dates if provided, otherwise use current date
	if templateSettings.Created == "" {
		templateSettings.Created = time.Now().Format(templateSettings.DateFormat)
	} else {
		parsed, err := time.Parse(templateSettings.DateFormat, templateSettings.Created)
		if err != nil {
			return fmt.Errorf("error parsing created date: %w", err)
		}
		templateSettings.Created = parsed.Format(templateSettings.DateFormat)
	}
	if templateSettings.Updated == "" {
		templateSettings.Updated = time.Now().Format(templateSettings.DateFormat)
	} else {
		parsed, err := time.Parse(templateSettings.DateFormat, templateSettings.Updated)
		if err != nil {
			return fmt.Errorf("error parsing updated date: %w", err)
		}
		templateSettings.Updated = parsed.Format(templateSettings.DateFormat)
	}

	// Generate filename based on current date and title
	filename := "new_template.md"
	if templateSettings.Title != "" || templateSettings.Created != "" {
		filename = templateSettings.Created + " " + templateSettings.Title + ".md"
	}

	templatePath := filepath.Join(templateSettings.OutputDirectory, filename)

	file, err := os.Create(templatePath)
	if err != nil {
		return fmt.Errorf("error creating template file: %w", err)
	}
	defer file.Close()

	// Execute the template, populating it with TemplateSettings data.
	if err := tmpl.Execute(file, templateSettings); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	fmt.Printf("Markdown template created at: %s\n", templatePath)
	return nil
}

// startWatcher initializes and starts the file system watcher for automatic rebuilds.
// It sets up a watcher on the input directory and specified custom assets.
// On file changes, it triggers the website rebuild process and logs the activity.
// It also starts a local HTTP server to serve the generated website during watch mode.
func startWatcher(settings parse.Settings) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the input directory to the watcher for content file changes.
	if err := watcher.Add(settings.InputDirectory); err != nil {
		log.Fatal(err)
	}

	// Recursively add all subdirectories within the input directory to watch for new content.
	err = filepath.WalkDir(settings.InputDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err := watcher.Add(path); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Add custom asset paths to the watcher to rebuild on CSS, JS, or favicon changes.
	if settings.PathToCustomCss != "" {
		if err := watcher.Add(settings.PathToCustomCss); err != nil {
			log.Fatal(err)
		}
	}
	if settings.PathToCustomJs != "" {
		if err := watcher.Add(settings.PathToCustomJs); err != nil {
			log.Fatal(err)
		}
	}
	if settings.PathToCustomFavicon != "" {
		if err := watcher.Add(settings.PathToCustomFavicon); err != nil {
			log.Fatal(err)
		}
	}

	// Start the HTTP server in a goroutine to serve the website in the background.
	go serve(settings)
	log.Printf("\n%s Watching for changes in '%s'...\n", time.Now().Format(time.RFC850), settings.InputDirectory)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Trigger rebuild on write events (file modifications).
			if event.Has(fsnotify.Write) {
				log.Println("File change detected:", event.Name, "- Rebuilding website...")
				buildWebsite(settings)
				log.Printf("\n%s Watching for changes in '%s'...\n", time.Now().Format(time.RFC850), settings.InputDirectory)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

// serve starts a simple HTTP server to serve the generated website files.
// It serves files from the output directory, making the generated blog accessible in a browser.
func serve(settings parse.Settings) {
	addr := ":666" // TODO: Allow the port to be specified as a command-line flag for customization.
	fmt.Printf("Serving website from '%s' at http://localhost%s. Press Ctrl+C to stop.\n", settings.OutputDirectory, addr)
	http.Handle("/", http.FileServer(http.Dir(settings.OutputDirectory)))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err) // Fatal error if server fails to start.
	}
}

// cleanContent prepares text content for Fuse.js search indexing.
// It removes or replaces characters that could negatively impact search accuracy,
// simplifying the text to improve search matching.
func cleanContent(s string) []string {
	replacements := map[string]string{
		"’": "'", // Replace curly apostrophe with straight apostrophe for consistency.
		"–": " ", // Replace en dash with space for better word separation.
	}

	removals := []string{
		"\n", "\r", "\t", "(", ")", "[", "]", "{", "}",
		"\"", "\\", "/", "”", "#", "-", "*", // Remove punctuation, formatting, and special characters.
	}

	// Apply character replacements.
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Apply character removals.
	for _, char := range removals {
		s = strings.ReplaceAll(s, char, " ")
	}

	// Split the cleaned string into words, using whitespace as delimiters.
	return strings.Fields(s)
}

// applyCSSTemplate applies the selected theme's styles to the default CSS template.
// If no custom CSS path is provided, this function generates the 'style.css' file
// in the output directory based on the chosen predefined theme.
func applyCSSTemplate(themeData parse.Theme, outputDirectory string) error {
	tmpl, err := texttemplate.ParseFS(assets, "src/assets/templates/style.gocss")
	if err != nil {
		return fmt.Errorf("error parsing style template: %w", err)
	}

	var output strings.Builder
	// Execute the CSS template with the theme data to generate the final CSS content.
	err = tmpl.Execute(&output, themeData)
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

// buildWebsite is the main function for generating the static website.
// It orchestrates the entire build process: clearing output directory, parsing content files,
// generating index and RSS feeds, handling CSS and JavaScript, copying assets, and logging completion.
func buildWebsite(settings parse.Settings) {
	// Clear the output directory to ensure a clean website build.
	err := os.RemoveAll(settings.OutputDirectory)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error clearing output directory: %v", err)
	}

	if err := os.MkdirAll(settings.OutputDirectory, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Retrieve a list of all Markdown and HTML content files from the input directory.
	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatalf("Error getting content files: %v", err)
	}

	var articles []parse.Article             // Slice to store parsed and processed articles.
	var searchIndex []map[string]interface{} // Slice to store article data for Fuse.js search index.

	// Process each content file to parse metadata and generate HTML content.
	for _, path := range files {
		article, err := processFile(path, settings)
		if err != nil {
			log.Printf("Error processing file %s: %v\n", path, err) // Log error and continue to next file.
			continue
		}
		articles = append(articles, article) // Add processed article to the list.

		// Prepare article data for the search index.
		searchIndex = append(searchIndex, map[string]interface{}{
			"title":       article.Title,
			"content":     cleanContent(article.TextContent), // Clean text content for better search.
			"description": article.Description,
			"tags":        article.Tags,
			"url":         article.LinkToSelf,
		})
	}

	// Sort articles based on the sorting method specified in settings.
	switch settings.Sort {
	case "date-created":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Created.After(articles[j].Created)
		})
	case "reverse-date-created":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Created.Before(articles[j].Created)
		})
	case "date-updated":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Updated.After(articles[j].Updated)
		})
	case "reverse-date-updated":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Updated.Before(articles[j].Updated)
		})
	case "title":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Title < articles[j].Title
		})
	case "reverse-title":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Title > articles[j].Title
		})
	case "path":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].OriginalPath < articles[j].OriginalPath
		})
	case "reverse-path":
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].OriginalPath > articles[j].OriginalPath
		})
	}

	// Generate and save the Fuse.js search index as JSON.
	searchIndexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		log.Fatalf("Error marshaling search index to JSON: %v", err)
	}
	searchIndexPath := filepath.Join(settings.OutputDirectory, "search_index.json")
	if err := os.WriteFile(searchIndexPath, searchIndexJSON, 0644); err != nil {
		log.Fatalf("Error saving search index JSON file: %v", err)
	}

	// Generate the main HTML index page listing all articles.
	if err := parse.GenerateHtmlIndex(articles, settings, assets); err != nil {
		log.Fatalf("Error generating HTML index page: %v", err)
	}

	// Generate the RSS feed XML file for content syndication.
	if err := parse.GenerateRSS(articles, settings, assets); err != nil {
		log.Fatalf("Error generating RSS feed: %v", err)
	}

	// Handle CSS: use custom CSS if provided, otherwise apply the selected theme template.
	if settings.PathToCustomCss == "" {
		theme := parse.GetThemeData(settings.Theme)
		applyCSSTemplate(theme, settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomCss, filepath.Join(settings.OutputDirectory, "style.css")); err != nil {
			log.Fatalf("Error handling custom CSS file: %v", err)
		}
	}

	// Handle JavaScript: use custom JS if provided, otherwise copy the default script.
	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomJs, filepath.Join(settings.OutputDirectory, "script.js")); err != nil {
			log.Fatalf("Error handling custom JavaScript file: %v", err)
		}
	}

	// Handle favicon: use custom favicon if provided, otherwise copy the default favicon.
	if settings.PathToCustomFavicon == "" {
		saveAsset("favicon.ico", "favicon.ico", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomFavicon, filepath.Join(settings.OutputDirectory, "favicon.ico")); err != nil {
			log.Fatalf("Error handling custom favicon file: %v", err)
		}
	}

	// Copy static assets to the output directory (search script, social media icons, etc.).
	saveAsset("search.js", "search.js", settings.OutputDirectory)
	saveAsset("rss.svg", "rss.svg", settings.OutputDirectory)
	saveAsset("telegram.svg", "telegram.svg", settings.OutputDirectory)
	saveAsset("bluesky.svg", "bluesky.svg", settings.OutputDirectory)
	saveAsset("mastodon.svg", "mastodon.svg", settings.OutputDirectory)
	saveAsset("threads.svg", "threads.svg", settings.OutputDirectory)
	saveAsset("x.svg", "x.svg", settings.OutputDirectory)
	saveAsset("share.svg", "share.svg", settings.OutputDirectory)
	saveAsset("follow.svg", "follow.svg", settings.OutputDirectory)
	saveAsset("reddit.svg", "reddit.svg", settings.OutputDirectory)
	saveAsset("linkedin.svg", "linkedin.svg", settings.OutputDirectory)
	saveAsset("hackernews.svg", "hackernews.svg", settings.OutputDirectory)
	saveAsset("facebook.svg", "facebook.svg", settings.OutputDirectory)
	saveAsset("copy.svg", "copy.svg", settings.OutputDirectory)

	log.Println("Website generated successfully in:", settings.OutputDirectory) // Success log message
}

// processFile parses a single Markdown or HTML content file.
// It reads the file, parses its content based on file type, extracts metadata,
// copies associated resources, and prepares an Article struct for website generation.
func processFile(filePath string, settings parse.Settings) (parse.Article, error) {
	var article parse.Article
	var err error

	filePathLower := strings.ToLower(filePath)

	// Process Markdown files (.md).
	if strings.HasSuffix(filePathLower, ".md") {
		article, err = parse.MarkdownFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing markdown file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources for markdown file: %w", err)
		}
		if err := parse.FormatMarkdown(&article, settings, assets); err != nil {
			return parse.Article{}, fmt.Errorf("error formatting markdown: %w", err)
		}
	} else if strings.HasSuffix(filePath, ".html") { // Process HTML files (.html).
		article, err = parse.HTMLFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing HTML file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources for HTML file: %w", err)
		}
	} else { // Return error for unsupported file types.
		return parse.Article{}, fmt.Errorf("unsupported file type: %s", filePath)
	}

	// Write the processed HTML content to the output directory, creating the article's HTML file.
	if err := os.WriteFile(article.LinkToSave, []byte(article.HtmlContent), 0644); err != nil {
		return parse.Article{}, fmt.Errorf("error writing processed file: %w", err)
	}
	return article, nil
}

// saveAsset copies a static asset file from the embedded 'assets' directory to the output directory.
// Used for deploying default CSS, JavaScript, images, and other static files included with DSBG.
func saveAsset(assetName string, saveName string, outputDirectory string) {
	file, err := assets.ReadFile("src/assets/" + assetName)
	if err != nil {
		log.Fatalf("Error reading asset '%s' from embedded assets: %v", assetName, err)
	}

	pathToSave := filepath.Join(outputDirectory, saveName)
	if err := os.WriteFile(pathToSave, file, 0644); err != nil {
		log.Fatalf("Error saving asset '%s' to output directory: %v", assetName, err)
	}
}

// copyFile copies an arbitrary file from a source path to a destination path.
// This utility function is used for copying custom user-provided files like CSS, JavaScript, or favicons.
func copyFile(srcPath string, destPath string) error {
	input, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("error reading file from source path '%s': %w", srcPath, err)
	}
	err = os.WriteFile(destPath, input, 0644)
	if err != nil {
		return fmt.Errorf("error writing file to destination path '%s': %w", destPath, err)
	}
	return nil
}
