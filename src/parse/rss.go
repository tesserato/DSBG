package parse

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	texttemplate "text/template"
	"time"
)

// GenerateRSS creates an RSS feed XML file from the processed articles.
// It ensures proper formatting for RSS elements like title, link, description, and pubDate.
// Returns an error if template parsing or execution fails, or if writing the output file fails.
func GenerateRSS(articles []Article, settings Settings, tmpl *texttemplate.Template, assets fs.FS) error {
	// Sort articles by creation date in descending order for the RSS feed.
	slices.SortFunc(articles, func(a, b Article) int {
		return b.Created.Compare(a.Created)
	})

	// Execute the template with article data and settings.
	var tp bytes.Buffer
	err := tmpl.Execute(&tp, struct {
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