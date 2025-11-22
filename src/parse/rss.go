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
// It sorts articles by creation date in descending order and writes rss.xml
// into the output directory.
func GenerateRSS(articles []Article, settings Settings, tmpl *texttemplate.Template, assets fs.FS) error {
	// Sort articles by creation date in descending order for the RSS feed.
	slices.SortFunc(articles, func(a, b Article) int {
		return b.Created.Compare(a.Created)
	})

	var tp bytes.Buffer
	err := tmpl.Execute(&tp, struct {
		Articles  []Article
		Settings  Settings
		BuildDate string
	}{
		Articles:  articles,
		Settings:  settings,
		BuildDate: time.Now().Format(time.RFC1123Z),
	})
	if err != nil {
		return fmt.Errorf("error executing RSS template: %w", err)
	}

	filePath := filepath.Join(settings.OutputPath, "rss.xml")
	if err := os.WriteFile(filePath, tp.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing RSS file to '%s': %w", filePath, err)
	}
	return nil
}
