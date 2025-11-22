package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tesserato/DSBG/src/parse"
)

//go:embed src/assets
var assets embed.FS

// ANSI Color Codes for Help Output
const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cCyan   = "\033[36m"
	cGray   = "\033[90m"
	cWhite  = "\033[97m"
)

// shareButtonsFlag is a custom flag type that collects repeated --share flags.
type shareButtonsFlag []parse.ShareButton

// String returns a human-readable description of the shareButtonsFlag format.
func (s *shareButtonsFlag) String() string {
	return "Share buttons defined by Name|Display|UrlTemplate"
}

// Set parses and appends a value to shareButtonsFlag.
func (s *shareButtonsFlag) Set(value string) error {
	parts := strings.SplitN(value, "|", 3)

	switch len(parts) {
	case 2:
		// Format: Name|UrlTemplate
		*s = append(*s, parse.ShareButton{
			Name:        parts[0],
			Display:     parts[0],
			UrlTemplate: parts[1],
		})
	case 3:
		// Format: Name|Display|UrlTemplate
		*s = append(*s, parse.ShareButton{
			Name:        parts[0],
			Display:     parts[1],
			UrlTemplate: parts[2],
		})
	default:
		return fmt.Errorf("invalid share format. Expected 'Name|URL' or 'Name|Display|URL', got '%s'", value)
	}
	return nil
}

// noFlagsPassed reports whether any flags were set in the provided FlagSet.
func noFlagsPassed(fs *flag.FlagSet) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		found = true
	})
	return !found
}

// printFlagHelp formats and prints a single flag's help with colors.
func printFlagHelp(f *flag.Flag) {
	if f == nil {
		return
	}
	name := fmt.Sprintf("-%s", f.Name)
	def := ""
	if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
		def = fmt.Sprintf(" (default: %s)", f.DefValue)
	}
	// Flag Name in Green, Default in Gray, Usage in Standard
	fmt.Fprintf(os.Stderr, "  %s%-24s%s%s%s\n    %s\n", cGreen, name, cGray, def, cReset, f.Usage)
}

// main is the entrypoint for DSBG (Dead Simple Blog Generator).
func main() {
	flagSet := flag.NewFlagSet("dsbg", flag.ExitOnError)

	var settings parse.Settings
	var shareButtons shareButtonsFlag

	// Prepare dynamic theme list for help text
	themeDesc := "Selects the built-in color scheme/CSS framework to use."
	if availableThemes, err := parse.GetAvailableThemes(assets); err == nil {
		themeDesc += fmt.Sprintf(" Available: [%s]", strings.Join(availableThemes, ", "))
	}

	// --- General Config ---
	flagSet.StringVar(&settings.Title, "title", "Blog", "The main title of your website. Used in the browser tab, header, and RSS feed.")
	flagSet.StringVar(&settings.BaseUrl, "base-url", "", "The public URL (e.g., https://example.com). Essential for generating correct Canonical URLs, RSS feeds, and Open Graph social meta tags.")
	flagSet.StringVar(&settings.InputPath, "input", "content", "Directory containing your source Markdown (.md) or HTML files.")
	flagSet.StringVar(&settings.OutputPath, "output", "public", "Directory where the generated static site will be saved.")
	flagSet.BoolVar(&settings.ForceOverwrite, "overwrite", false, "Skip the confirmation prompt when the output directory is not empty.")
	flagSet.StringVar(&settings.DescriptionMarkdown, "description", "This is my blog", "A short summary of your site. Supports Markdown links. Appears on the homepage (rendered) and in the HTML <meta name='description'> tag (plain text).")

	// --- Metadata & SEO ---
	flagSet.StringVar(&settings.AuthorName, "author", "", "The default author name. Injected into JSON-LD structured data and <meta name='author'> tags.")
	flagSet.StringVar(&settings.PublisherName, "publisher", "", "Organization/Publisher name for structured data. Defaults to the Blog Title if omitted.")
	flagSet.StringVar(&settings.PublisherLogoPath, "logo", "", "Path to a logo image (relative to current dir). Injected into JSON-LD for Google search result branding.")
	flagSet.StringVar(&settings.DateFormat, "date-format", "2006 01 02", "Go layout string for rendering dates (e.g., 'Jan 02, 2006').")
	flagSet.StringVar(&settings.IndexName, "index-name", "index.html", "The filename to use for directory indexes. Change to 'README.html' if hosting on certain file servers.")

	// --- Theme & Customization ---
	flagSet.StringVar(&settings.Theme, "theme", "default", themeDesc)
	flagSet.StringVar(&settings.PathToCustomCss, "css-path", "", "Path to a local CSS file. If set, this REPLACES the built-in theme entirely.")
	flagSet.StringVar(&settings.PathToCustomJs, "js-path", "", "Path to a local JS file. Appended to the site's default functionality.")
	flagSet.StringVar(&settings.PathToCustomFavicon, "favicon-path", "", "Path to a 'favicon.ico' file to replace the default icon.")
	flagSet.Var(&shareButtons, "share", "Add a custom share button. Format: 'Name|Icon.svg|URL_Template'. Can be used multiple times. See variables below.")

	// --- Injections ---
	pathToAdditionalElementsTop := flagSet.String("elements-top", "", "Path to an HTML snippet to inject at the top of the <head> tag (e.g., Analytics scripts).")
	pathToAdditionalElemensBottom := flagSet.String("elements-bottom", "", "Path to an HTML snippet to inject at the bottom of the <body> tag (e.g., Comment widgets).")

	// --- Behavior Toggles ---
	sortFlag := flagSet.String("sort", "date-created", "Order of articles on the homepage. Options: date-created, date-updated, title, path (prefix with 'reverse-' to flip).")
	flagSet.BoolVar(&settings.DoNotExtractTagsFromPaths, "ignore-tags-from-paths", false, "If true, folder names in the source path (e.g., content/linux/...) are NOT added as tags.")
	flagSet.BoolVar(&settings.DoNotRemoveDateFromPaths, "keep-date-in-paths", false, "If true, date patterns in filenames (2023-01-01-post.md) are preserved in the output URL.")
	flagSet.BoolVar(&settings.DoNotRemoveDateFromTitles, "keep-date-in-titles", false, "If true, date patterns in filenames are preserved in the Article Title string.")
	flagSet.BoolVar(&settings.OpenInNewTab, "open-in-new-tab", false, "If true, clicking articles on the homepage opens them in a new browser tab/window.")

	// --- Dev Server ---
	watch := flagSet.Bool("watch", false, "Watch mode: Starts a local web server and automatically rebuilds the site when source files change.")
	flagSet.StringVar(&settings.Port, "port", "666", "The port to use for the local preview server (used with -watch).")

	// --- Custom Usage Output ---
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "%sDSBG: Dead Simple Blog Generator%s\n", cBold+cCyan, cReset)
		fmt.Fprintln(os.Stderr, "A minimalist, single-binary static site generator.")
		fmt.Fprintln(os.Stderr)

		fmt.Fprintf(os.Stderr, "%sUSAGE:%s\n", cBold+cYellow, cReset)
		fmt.Fprintln(os.Stderr, "  dsbg [flags]")
		fmt.Fprintln(os.Stderr)

		// Helper to print a group of flags
		printGroup := func(title string, flagNames ...string) {
			fmt.Fprintf(os.Stderr, "%s%s:%s\n", cBold+cWhite, title, cReset)
			for _, name := range flagNames {
				printFlagHelp(flagSet.Lookup(name))
			}
			fmt.Fprintln(os.Stderr)
		}

		printGroup("GENERAL CONFIGURATION", "input", "output", "title", "description", "base-url", "overwrite")
		printGroup("METADATA & SEO", "author", "publisher", "logo", "date-format")
		printGroup("THEMING & UI", "theme", "css-path", "js-path", "favicon-path", "share")
		printGroup("INJECTIONS", "elements-top", "elements-bottom")
		printGroup("CONTENT BEHAVIOR", "sort", "ignore-tags-from-paths", "keep-date-in-paths", "keep-date-in-titles", "open-in-new-tab", "index-name")
		printGroup("LOCAL DEVELOPMENT", "watch", "port")

		fmt.Fprintf(os.Stderr, "%sFRONTMATTER METADATA:%s\n", cBold+cYellow, cReset)
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "share_url", "Override the URL shared by buttons (good for link-blogging).")
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "canonical_url", "Set the 'rel=canonical' tag (good for cross-posting/SEO).")
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "cover_image", "Path to an image to display on the index and in social cards.")
		fmt.Fprintln(os.Stderr)

		fmt.Fprintf(os.Stderr, "%sSHARE TEMPLATE VARIABLES:%s\n", cBold+cYellow, cReset)
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "{URL}", "Public URL of the article (or share_url if set)")
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "{TITLE}", "Article title")
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "{DESCRIPTION}", "Article description")
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", "{TARGET_URL}", "The target destination: uses 'share_url' if present, else the first link in text.")
		fmt.Fprintln(os.Stderr)

		fmt.Fprintf(os.Stderr, "%sSHARE EXAMPLES:%s\n", cBold+cYellow, cReset)
		fmt.Fprintf(os.Stderr, "  %s1. Standard Share:%s\n", cWhite, cReset)
		fmt.Fprintf(os.Stderr, "     -share \"X|https://x.com/intent/tweet?text={TITLE}&url={URL}\"\n")
		fmt.Fprintf(os.Stderr, "  %s2. 'HackerNews' Style Submission (uses share_url or first link):%s\n", cWhite, cReset)
		fmt.Fprintf(os.Stderr, "     -share \"HN|https://news.ycombinator.com/submitlink?u={TARGET_URL}&t={TITLE}\"\n")
		fmt.Fprintln(os.Stderr)

		// Dynamic Date for the example
		today := time.Now().Format("2006 01 02")

		fmt.Fprintf(os.Stderr, "%sTEMPLATE EXAMPLE:%s\n", cBold+cYellow, cReset)
		fmt.Fprintln(os.Stderr, "  Copy and paste this frontmatter at the top of your Markdown files:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "%s  ---%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  title: My New Post%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  description: A short summary of the post.%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  created: %s%s\n", cCyan, today, cReset)
		fmt.Fprintf(os.Stderr, "%s  tags: Technology, Go%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  cover_image: image.webp%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  share_url: (optional override)%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  canonical_url: (optional SEO override)%s\n", cCyan, cReset)
		fmt.Fprintf(os.Stderr, "%s  ---%s\n", cCyan, cReset)
		fmt.Fprintln(os.Stderr)
	}

	// Show usage if no arguments are provided
	if len(os.Args) <= 1 {
		flagSet.Usage()
		return
	}

	// Parse flags
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	settings.ShareButtons = shareButtons

	var buf strings.Builder
	if err := parse.Markdown.Convert([]byte(settings.DescriptionMarkdown), &buf); err != nil {
		log.Fatalf("failed to convert description to HTML: %v", err)
	}
	settings.DescriptionHTML = template.HTML(buf.String())

	if _, err := os.Stat(settings.InputPath); os.IsNotExist(err) {
		if noFlagsPassed(flagSet) {
			flagSet.Usage()
			return
		}
		log.Fatalf("Input directory '%s' does not exist.", settings.InputPath)
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

	// Default author / publisher names to blog title if not provided.
	if settings.AuthorName == "" {
		settings.AuthorName = settings.Title
	}
	if settings.PublisherName == "" {
		settings.PublisherName = settings.Title
	}

	// Determine syntax highlight theme automatically from CSS.
	themeType := parse.GetThemeType(assets, settings.Theme)
	if themeType == "light" {
		settings.HighlightTheme = "stackoverflow-light"
	} else {
		settings.HighlightTheme = "github-dark-dimmed"
	}

	// Parse sort order into strongly-typed SortOrder.
	sortOrder, err := parse.ParseSortOrder(*sortFlag)
	if err != nil {
		log.Fatalf("invalid sort order '%s': %v", *sortFlag, err)
	}
	settings.Sort = sortOrder

	// Parse templates once.
	templates, err := parse.LoadTemplates(assets)
	if err != nil {
		log.Fatalf("Error loading templates: %v", err)
	}

	// Perform initial build (clean=true).
	if err := buildWebsite(&settings, templates, true); err != nil {
		log.Fatal(err)
	}

	if *watch {
		// In watch mode, start the server and open the browser ONCE here.
		addr := ":" + settings.Port
		url := fmt.Sprintf("http://localhost%s", addr)

		go serve(settings)

		// Small delay so the server is listening before opening the browser.
		go func() {
			time.Sleep(300 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				log.Printf("Could not open browser: %v\n", err)
			}
		}()

		// Block here to watch for changes and rebuild.
		startWatcher(&settings, templates)
	}
}

// startWatcher monitors input and asset changes and triggers rebuilds.
func startWatcher(settings *parse.Settings, templates parse.SiteTemplates) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(settings.InputPath); err != nil {
		log.Fatal(err)
	}

	err = filepath.WalkDir(settings.InputPath, func(path string, d fs.DirEntry, err error) error {
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
		_ = watcher.Add(settings.PathToCustomCss)
	}
	if settings.PathToCustomJs != "" {
		_ = watcher.Add(settings.PathToCustomJs)
	}
	if settings.PathToCustomFavicon != "" {
		_ = watcher.Add(settings.PathToCustomFavicon)
	}

	log.Printf("\n%s Watching for changes in '%s'...\n", time.Now().Format(time.RFC850), settings.InputPath)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				log.Println("File change detected:", event.Name, "- Rebuilding website...")
				if err := buildWebsite(settings, templates, false); err != nil {
					log.Printf("Rebuild failed: %v\n", err)
				}
				log.Printf("\n%s Watching for changes in '%s'...\n", time.Now().Format(time.RFC850), settings.InputPath)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

// openBrowser tries to open the given URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// serve starts an HTTP file server for the generated output directory.
func serve(settings parse.Settings) {
	addr := ":" + settings.Port
	url := fmt.Sprintf("http://localhost%s", addr)
	fmt.Printf("Serving website from '%s' at %s. Press Ctrl+C to stop.\n", settings.OutputPath, url)
	http.Handle("/", http.FileServer(http.Dir(settings.OutputPath)))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// deleteChildren removes all children of a directory but keeps the directory itself.
func deleteChildren(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// buildWebsite generates the website based on the provided settings and templates.
func buildWebsite(settings *parse.Settings, templates parse.SiteTemplates, clean bool) error {
	if clean {
		// Check output directory safety.
		if !settings.ForceOverwrite {
			f, err := os.Open(settings.OutputPath)
			if err == nil {
				defer f.Close()
				_, err = f.Readdirnames(1)
				if err == nil {
					fmt.Printf("Output directory '%s' is not empty. Overwrite? (y/n): ", settings.OutputPath)
					reader := bufio.NewReader(os.Stdin)
					response, _ := reader.ReadString('\n')
					response = strings.TrimSpace(strings.ToLower(response))
					if response != "y" && response != "yes" {
						return fmt.Errorf("operation cancelled by user")
					}
					settings.ForceOverwrite = true
				}
			}
		}

		// Try to clean children first to avoid root lock issues.
		if err := deleteChildren(settings.OutputPath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: Failed to clean output directory: %v. Trying to proceed...", err)
			}
		}
	}

	if err := os.MkdirAll(settings.OutputPath, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Handle share assets.
	for i, btn := range settings.ShareButtons {
		if strings.HasPrefix(btn.Display, "http://") || strings.HasPrefix(btn.Display, "https://") {
			continue
		}
		if parse.IsImage(btn.Display) {
			src := btn.Display
			destName := filepath.Base(src)
			destPath := filepath.Join(settings.OutputPath, destName)
			if err := copyFile(src, destPath); err != nil {
				log.Printf("Warning: Failed to copy share icon '%s': %v", src, err)
			}
			settings.ShareButtons[i].Display = destName
		}
	}

	// Handle publisher logo asset (relative to command, copied to output root).
	if settings.PublisherLogoPath != "" &&
		!strings.HasPrefix(settings.PublisherLogoPath, "http://") &&
		!strings.HasPrefix(settings.PublisherLogoPath, "https://") {

		src := settings.PublisherLogoPath
		destName := filepath.Base(src)
		destPath := filepath.Join(settings.OutputPath, destName)
		if err := copyFile(src, destPath); err != nil {
			log.Printf("Warning: Failed to copy publisher logo '%s': %v", src, err)
		} else {
			// After copying, make the path relative to the site root for templates/JSON-LD.
			settings.PublisherLogoPath = destName
		}
	}

	files, err := parse.GetPaths(settings.InputPath, []string{".md", ".html"})
	if err != nil {
		return fmt.Errorf("error getting content files: %v", err)
	}

	var articles []parse.Article
	var searchIndex []map[string]interface{}
	var mu sync.Mutex

	// Worker pool for concurrent processing of files.
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}

	pathsCh := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range pathsCh {
				article, err := processFile(filePath, *settings, templates)
				if err != nil {
					log.Printf("Error processing file %s: %v\n", filePath, err)
					continue
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
			}
		}()
	}

	for _, path := range files {
		pathsCh <- path
	}
	close(pathsCh)
	wg.Wait()

	switch settings.Sort {
	case parse.SortDateCreated:
		sort.Slice(articles, func(i, j int) bool { return articles[i].Created.After(articles[j].Created) })
	case parse.SortReverseDateCreated:
		sort.Slice(articles, func(i, j int) bool { return articles[i].Created.Before(articles[j].Created) })
	case parse.SortDateUpdated:
		sort.Slice(articles, func(i, j int) bool { return articles[i].Updated.After(articles[j].Updated) })
	case parse.SortReverseDateUpdated:
		sort.Slice(articles, func(i, j int) bool { return articles[i].Updated.Before(articles[j].Updated) })
	case parse.SortTitle:
		sort.Slice(articles, func(i, j int) bool { return articles[i].Title < articles[j].Title })
	case parse.SortReverseTitle:
		sort.Slice(articles, func(i, j int) bool { return articles[i].Title > articles[j].Title })
	case parse.SortPath:
		sort.Slice(articles, func(i, j int) bool { return articles[i].OriginalPath < articles[j].OriginalPath })
	case parse.SortReversePath:
		sort.Slice(articles, func(i, j int) bool { return articles[i].OriginalPath > articles[j].OriginalPath })
	}

	searchIndexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		return fmt.Errorf("error marshaling search index to JSON: %v", err)
	}
	searchIndexPath := filepath.Join(settings.OutputPath, "search_index.json")
	if err := os.WriteFile(searchIndexPath, searchIndexJSON, 0644); err != nil {
		return fmt.Errorf("error saving search index JSON file: %v", err)
	}

	if err := parse.GenerateHtmlIndex(articles, *settings, templates.Index, assets); err != nil {
		return fmt.Errorf("error generating HTML index page: %v", err)
	}

	if err := parse.GenerateRSS(articles, *settings, templates.RSS, assets); err != nil {
		return fmt.Errorf("error generating RSS feed: %v", err)
	}

	if settings.PathToCustomCss == "" {
		if err := parse.SaveThemeCSS(assets, settings.Theme, settings.OutputPath); err != nil {
			return fmt.Errorf("error processing theme CSS: %v", err)
		}
	} else {
		if err := copyFile(settings.PathToCustomCss, filepath.Join(settings.OutputPath, "style.css")); err != nil {
			return fmt.Errorf("error handling custom CSS file: %v", err)
		}
	}

	if settings.PathToCustomJs == "" {
		saveAsset("script.js", "script.js", settings.OutputPath)
	} else {
		if err := copyFile(settings.PathToCustomJs, filepath.Join(settings.OutputPath, "script.js")); err != nil {
			return fmt.Errorf("error handling custom JavaScript file: %v", err)
		}
	}

	if settings.PathToCustomFavicon == "" {
		saveAsset("favicon.ico", "favicon.ico", settings.OutputPath)
	} else {
		if err := copyFile(settings.PathToCustomFavicon, filepath.Join(settings.OutputPath, "favicon.ico")); err != nil {
			return fmt.Errorf("error handling custom favicon file: %v", err)
		}
	}

	saveAsset("search.js", "search.js", settings.OutputPath)
	saveAsset("rss.svg", "rss.svg", settings.OutputPath)
	saveAsset("copy.svg", "copy.svg", settings.OutputPath)

	log.Println("Website generated successfully in:", settings.OutputPath)
	return nil
}

// processFile parses a single Markdown or HTML file into an Article and writes its output HTML.
func processFile(filePath string, settings parse.Settings, templates parse.SiteTemplates) (parse.Article, error) {
	var article parse.Article
	var resources []string
	var err error
	filePathLower := strings.ToLower(filePath)

	if strings.HasSuffix(filePathLower, ".md") {
		article, resources, err = parse.MarkdownFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing markdown file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article, resources); err != nil {
			return parse.Article{}, fmt.Errorf("error copying resources: %w", err)
		}
		if err := parse.FormatMarkdown(&article, settings, templates.Article, assets); err != nil {
			return parse.Article{}, fmt.Errorf("error formatting markdown: %w", err)
		}
	} else if strings.HasSuffix(filePathLower, ".html") {
		article, resources, err = parse.HTMLFile(filePath)
		if err != nil {
			return parse.Article{}, fmt.Errorf("error parsing HTML file: %w", err)
		}
		if err := parse.CopyHtmlResources(settings, &article, resources); err != nil {
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

// saveAsset copies a named embedded asset from the assets filesystem into the output directory.
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

// copyFile copies a file from srcPath to destPath on the local filesystem.
func copyFile(srcPath string, destPath string) error {
	input, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", srcPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("error creating directory for '%s': %w", destPath, err)
	}
	if err := os.WriteFile(destPath, input, 0644); err != nil {
		return fmt.Errorf("error writing file '%s': %w", destPath, err)
	}
	return nil
}
