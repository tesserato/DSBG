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
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tesserato/DSBG/src/parse"
)

//go:embed src/assets
var assets embed.FS

// shareButtonsFlag is a custom flag type to handle repeated --share flags
type shareButtonsFlag []parse.ShareButton

func (s *shareButtonsFlag) String() string {
	return "Share buttons defined by Name|Display|UrlTemplate"
}

func (s *shareButtonsFlag) Set(value string) error {
	// Split by pipe. Limit to 3 parts so the URL can contain pipes if necessary.
	parts := strings.SplitN(value, "|", 3)

	if len(parts) == 2 {
		// Format: Name|UrlTemplate
		// Default Display to Name
		*s = append(*s, parse.ShareButton{
			Name:        parts[0],
			Display:     parts[0],
			UrlTemplate: parts[1],
		})
	} else if len(parts) == 3 {
		// Format: Name|Display|UrlTemplate
		*s = append(*s, parse.ShareButton{
			Name:        parts[0],
			Display:     parts[1],
			UrlTemplate: parts[2],
		})
	} else {
		return fmt.Errorf("invalid share format. Expected 'Name|URL' or 'Name|Display|URL', got '%s'", value)
	}
	return nil
}

// noFlagsPassed checks if any command-line flags were provided for a given FlagSet.
func noFlagsPassed(fs *flag.FlagSet) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		found = true
	})
	return !found
}

// logFlag is a helper function to log the details of a flag.
func logFlag(f *flag.Flag) {
	defaultValue := f.DefValue
	if defaultValue != "" {
		defaultValue = fmt.Sprintf(" (default: %v)", defaultValue)
	}
	fmt.Fprintf(os.Stderr, "  -%v %v\n    %v \n", f.Name, defaultValue, f.Usage)
}

func main() {
	defaultFlagSet := flag.NewFlagSet("default", flag.ExitOnError)
	templateFlagSet := flag.NewFlagSet("template", flag.ExitOnError)

	var settings parse.Settings
	var shareButtons shareButtonsFlag

	// --- Default FlagSet Flags ---
	defaultFlagSet.StringVar(&settings.Title, "title", "Blog", "The title of the blog, used in the header and page titles.")
	defaultFlagSet.StringVar(&settings.BaseUrl, "base-url", "", "The base URL of the blog (e.g., https://example.com).")
	defaultFlagSet.StringVar(&settings.DescriptionMarkdown, "description", "This is my blog", "A short description of the blog.")
	defaultFlagSet.StringVar(&settings.InputDirectory, "input-path", "content", "Path to source content files.")
	defaultFlagSet.StringVar(&settings.OutputDirectory, "output-path", "public", "Path to output directory.")
	defaultFlagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Date format (Go style).")
	defaultFlagSet.StringVar(&settings.IndexName, "index-name", "index.html", "Filename for the main index page.")
	defaultFlagSet.StringVar(&settings.PathToCustomCss, "css-path", "", "Path to a custom CSS file.")
	defaultFlagSet.StringVar(&settings.PathToCustomJs, "js-path", "", "Path to a custom JavaScript file.")
	defaultFlagSet.StringVar(&settings.PathToCustomFavicon, "favicon-path", "", "Path to a custom favicon file.")
	defaultFlagSet.BoolVar(&settings.DoNotExtractTagsFromPaths, "ignore-tags-from-paths", false, "Disable extracting tags from directory names.")
	defaultFlagSet.BoolVar(&settings.DoNotRemoveDateFromPaths, "keep-date-in-paths", false, "Do not remove date patterns from paths.")
	defaultFlagSet.BoolVar(&settings.DoNotRemoveDateFromTitles, "keep-date-in-titles", false, "Do not remove date patterns from titles.")
	defaultFlagSet.BoolVar(&settings.OpenInNewTab, "open-in-new-tab", false, "Open external links in new browser tabs.")
	defaultFlagSet.StringVar(&settings.Port, "port", "666", "Port for the local server (default: 666).")

	// Custom share flag
	defaultFlagSet.Var(&shareButtons, "share", "Repeatable flag to add share buttons. Format: \"Name|Display|UrlTemplate\".")

	defaultFlagSet.StringVar(&settings.Sort, "sort", "date-created", "Sort order for articles.")
	themeString := defaultFlagSet.String("theme", "default", "Predefined website style theme.")
	pathToAdditionalElementsTop := defaultFlagSet.String("elements-top", "", "HTML file to include at the top of <head>.")
	pathToAdditionalElemensBottom := defaultFlagSet.String("elements-bottom", "", "HTML file to include at the bottom of <body>.")
	watch := defaultFlagSet.Bool("watch", false, "Enable watch mode.")

	// --- Template FlagSet Flags ---
	var templateSettings parse.TemplateSettings
	templateFlagSet.StringVar(&templateSettings.Title, "title", "", "Title for template.")
	templateFlagSet.StringVar(&templateSettings.Description, "description", "", "Description for template.")
	templateFlagSet.StringVar(&templateSettings.Created, "created", "", "Created date for template.")
	templateFlagSet.StringVar(&templateSettings.Updated, "updated", "", "Updated date for template.")
	templateFlagSet.StringVar(&templateSettings.CoverImagePath, "cover-image-path", "", "Cover image path for template.")
	templateFlagSet.StringVar(&templateSettings.Tags, "tags", "", "Comma-separated tags for template.")
	templateFlagSet.StringVar(&templateSettings.OutputDirectory, "output-path", ".", "Directory to save the template.")
	templateFlagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Date format.")

	defaultFlagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "DSBG (Dead Simple Blog Generator)\n")
		fmt.Fprintf(os.Stderr, "Usage: dsbg [flags] or dsbg -template [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Default mode flags:\n")
		defaultFlagSet.VisitAll(logFlag)
		fmt.Fprintf(os.Stderr, "\nTemplate mode flags:\n")
		templateFlagSet.VisitAll(logFlag)
	}

	if len(os.Args) <= 1 {
		defaultFlagSet.Usage()
		return
	}

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
		return
	default:
		err := defaultFlagSet.Parse(os.Args[1:])
		if err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}
		log.Println("Running in blog generation mode...")
	}

	settings.ShareButtons = shareButtons

	var buf strings.Builder
	if err := parse.Markdown.Convert([]byte(settings.DescriptionMarkdown), &buf); err != nil {
		log.Fatalf("failed to convert description to HTML: %v", err)
	}
	settings.DescriptionHTML = template.HTML(buf.String())

	if _, err := os.Stat(settings.InputDirectory); os.IsNotExist(err) {
		if noFlagsPassed(defaultFlagSet) {
			defaultFlagSet.Usage()
			return
		}
		log.Fatalf("Input directory '%s' does not exist.", settings.InputDirectory)
	}

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

	if settings.BaseUrl == "" {
		settings.BaseUrl = fmt.Sprintf("http://localhost:%s", settings.Port)
	} else {
		settings.BaseUrl = strings.TrimSuffix(settings.BaseUrl, "/")
	}

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

	// Parse templates once
	templates, err := parse.LoadTemplates(assets)
	if err != nil {
		log.Fatalf("Error loading templates: %v", err)
	}

	buildWebsite(settings, templates)

	if *watch {
		startWatcher(settings, templates)
	}
}

func createMarkdownTemplate(templateSettings parse.TemplateSettings) error {
	tmpl, err := template.New("md-article.gomd").ParseFS(assets, "src/assets/templates/md-article.gomd")
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

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

	if err := tmpl.Execute(file, templateSettings); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	fmt.Printf("Markdown template created at: %s\n", templatePath)
	return nil
}

func startWatcher(settings parse.Settings, templates parse.SiteTemplates) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(settings.InputDirectory); err != nil {
		log.Fatal(err)
	}

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

	if settings.PathToCustomCss != "" {
		watcher.Add(settings.PathToCustomCss)
	}
	if settings.PathToCustomJs != "" {
		watcher.Add(settings.PathToCustomJs)
	}
	if settings.PathToCustomFavicon != "" {
		watcher.Add(settings.PathToCustomFavicon)
	}

	go serve(settings)
	log.Printf("\n%s Watching for changes in '%s'...\n", time.Now().Format(time.RFC850), settings.InputDirectory)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				log.Println("File change detected:", event.Name, "- Rebuilding website...")
				buildWebsite(settings, templates)
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

func serve(settings parse.Settings) {
	addr := ":" + settings.Port
	fmt.Printf("Serving website from '%s' at http://localhost%s. Press Ctrl+C to stop.\n", settings.OutputDirectory, addr)
	http.Handle("/", http.FileServer(http.Dir(settings.OutputDirectory)))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func buildWebsite(settings parse.Settings, templates parse.SiteTemplates) {
	err := os.RemoveAll(settings.OutputDirectory)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error clearing output directory: %v", err)
	}

	if err := os.MkdirAll(settings.OutputDirectory, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Handle Share Assets
	for i, btn := range settings.ShareButtons {
		if strings.HasPrefix(btn.Display, "http://") || strings.HasPrefix(btn.Display, "https://") {
			continue
		}
		if parse.IsImage(btn.Display) { // Fixed: use exported function IsImage
			src := btn.Display
			destName := filepath.Base(src)
			destPath := filepath.Join(settings.OutputDirectory, destName)
			if err := copyFile(src, destPath); err != nil {
				log.Printf("Warning: Failed to copy share icon '%s': %v", src, err)
			}
			settings.ShareButtons[i].Display = destName
		}
	}

	files, err := parse.GetPaths(settings.InputDirectory, []string{".md", ".html"})
	if err != nil {
		log.Fatalf("Error getting content files: %v", err)
	}

	var articles []parse.Article
	var searchIndex []map[string]interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Concurrent Processing
	for _, path := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			article, err := processFile(filePath, settings, templates)
			if err != nil {
				log.Printf("Error processing file %s: %v\n", filePath, err)
				return
			}

			mu.Lock()
			articles = append(articles, article)
			searchIndex = append(searchIndex, map[string]interface{}{
				"title":        article.Title,
				"content":      parse.CleanContent(article.TextContent),
				"description":  article.Description,
				"tags":         article.Tags,
				"url":          article.LinkToSelf,
				"html_content": article.HtmlContent,
			})
			mu.Unlock()
		}(path)
	}
	wg.Wait()

	switch settings.Sort {
	case "date-created":
		sort.Slice(articles, func(i, j int) bool { return articles[i].Created.After(articles[j].Created) })
	case "reverse-date-created":
		sort.Slice(articles, func(i, j int) bool { return articles[i].Created.Before(articles[j].Created) })
	case "date-updated":
		sort.Slice(articles, func(i, j int) bool { return articles[i].Updated.After(articles[j].Updated) })
	case "reverse-date-updated":
		sort.Slice(articles, func(i, j int) bool { return articles[i].Updated.Before(articles[j].Updated) })
	case "title":
		sort.Slice(articles, func(i, j int) bool { return articles[i].Title < articles[j].Title })
	case "reverse-title":
		sort.Slice(articles, func(i, j int) bool { return articles[i].Title > articles[j].Title })
	case "path":
		sort.Slice(articles, func(i, j int) bool { return articles[i].OriginalPath < articles[j].OriginalPath })
	case "reverse-path":
		sort.Slice(articles, func(i, j int) bool { return articles[i].OriginalPath > articles[j].OriginalPath })
	}

	searchIndexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		log.Fatalf("Error marshaling search index to JSON: %v", err)
	}
	searchIndexPath := filepath.Join(settings.OutputDirectory, "search_index.json")
	if err := os.WriteFile(searchIndexPath, searchIndexJSON, 0644); err != nil {
		log.Fatalf("Error saving search index JSON file: %v", err)
	}

	if err := parse.GenerateHtmlIndex(articles, settings, templates.Index, assets); err != nil {
		log.Fatalf("Error generating HTML index page: %v", err)
	}

	if err := parse.GenerateRSS(articles, settings, templates.RSS, assets); err != nil {
		log.Fatalf("Error generating RSS feed: %v", err)
	}

	if settings.PathToCustomCss == "" {
		theme := parse.GetThemeData(settings.Theme)
		parse.ApplyCSSTemplate(theme, settings.OutputDirectory, templates.Style)
	} else {
		if err := copyFile(settings.PathToCustomCss, filepath.Join(settings.OutputDirectory, "style.css")); err != nil {
			log.Fatalf("Error handling custom CSS file: %v", err)
		}
	}

	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomJs, filepath.Join(settings.OutputDirectory, "script.js")); err != nil {
			log.Fatalf("Error handling custom JavaScript file: %v", err)
		}
	}

	if settings.PathToCustomFavicon == "" {
		saveAsset("favicon.ico", "favicon.ico", settings.OutputDirectory)
	} else {
		if err := copyFile(settings.PathToCustomFavicon, filepath.Join(settings.OutputDirectory, "favicon.ico")); err != nil {
			log.Fatalf("Error handling custom favicon file: %v", err)
		}
	}

	saveAsset("search.js", "search.js", settings.OutputDirectory)
	saveAsset("rss.svg", "rss.svg", settings.OutputDirectory)
	saveAsset("copy.svg", "copy.svg", settings.OutputDirectory)

	log.Println("Website generated successfully in:", settings.OutputDirectory)
}

func processFile(filePath string, settings parse.Settings, templates parse.SiteTemplates) (parse.Article, error) {
	var article parse.Article
	var err error
	filePathLower := strings.ToLower(filePath)

	if strings.HasSuffix(filePathLower, ".md") {
		article, err = parse.MarkdownFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing markdown file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources: %w", err)
		}
		if err := parse.FormatMarkdown(&article, settings, templates.Article, assets); err != nil {
			return parse.Article{}, fmt.Errorf("error formatting markdown: %w", err)
		}
	} else if strings.HasSuffix(filePath, ".html") {
		article, err = parse.HTMLFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing HTML file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources: %w", err)
		}
	} else {
		return parse.Article{}, fmt.Errorf("unsupported file type: %s", filePath)
	}

	if err := os.WriteFile(article.LinkToSave, []byte(article.HtmlContent), 0644); err != nil {
		return parse.Article{}, fmt.Errorf("error writing processed file: %w", err)
	}
	return article, nil
}

func saveAsset(assetName string, saveName string, outputDirectory string) {
	file, err := assets.ReadFile("src/assets/" + assetName)
	if err != nil {
		log.Fatalf("Error reading asset '%s': %v", assetName, err)
	}
	pathToSave := filepath.Join(outputDirectory, saveName)
	if err := os.WriteFile(pathToSave, file, 0644); err != nil {
		log.Fatalf("Error saving asset '%s': %v", assetName, err)
	}
}

func copyFile(srcPath string, destPath string) error {
	input, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", srcPath, err)
	}
	err = os.WriteFile(destPath, input, 0644)
	if err != nil {
		return fmt.Errorf("error writing file '%s': %w", destPath, err)
	}
	return nil
}
