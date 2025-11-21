package parse

import (
	"html/template"
	"time"
)

// Style represents a high-level visual style/theme for the generated site.
type Style int64

// Supported Style values.
const (
	Default Style = iota
	Dark
	Clean
	Colorful
)

// SortOrder is a strongly-typed representation of article sort order.
type SortOrder string

// Supported SortOrder values.
const (
	SortDateCreated        SortOrder = "date-created"
	SortReverseDateCreated SortOrder = "reverse-date-created"
	SortDateUpdated        SortOrder = "date-updated"
	SortReverseDateUpdated SortOrder = "reverse-date-updated"
	SortTitle              SortOrder = "title"
	SortReverseTitle       SortOrder = "reverse-title"
	SortPath               SortOrder = "path"
	SortReversePath        SortOrder = "reverse-path"
)

// Settings holds global configuration for site generation.
type Settings struct {
	Title                     string
	DescriptionMarkdown       string
	DescriptionHTML           template.HTML
	InputDirectory            string
	OutputDirectory           string
	DateFormat                string
	IndexName                 string
	Theme                     Style
	PathToCustomCss           string
	PathToCustomJs            string
	PathToCustomFavicon       string
	AdditionalElementsTop     template.HTML
	AdditionalElemensBottom   template.HTML
	DoNotExtractTagsFromPaths bool
	DoNotRemoveDateFromPaths  bool
	DoNotRemoveDateFromTitles bool
	OpenInNewTab              bool
	BaseUrl                   string
	ShareButtons              []ShareButton
	Sort                      SortOrder
	HighlightTheme            string
	Port                      string
	ForceOverwrite            bool
}

// ShareButton describes a single social or custom share target.
type ShareButton struct {
	Name        string
	Display     string
	UrlTemplate string
}

// TemplateSettings controls generation of a new Markdown template file.
type TemplateSettings struct {
	Title           string
	Description     string
	Created         string
	Updated         string
	CoverImagePath  string
	Tags            string
	OutputDirectory string
	DateFormat      string
}

// Article models a single article or page, regardless of its original format.
type Article struct {
	Title          string
	Description    string
	CoverImagePath string
	Created        time.Time
	Updated        time.Time
	Tags           []string
	TextContent    string
	HtmlContent    string
	OriginalPath   string
	LinkToSelf     string
	LinkToSave     string
	Url            string
}

// Theme describes the visual styling parameters that are applied to the CSS template.
type Theme struct {
	Dark           bool
	HeaderFont     string
	BodyFont       string
	Background     string
	Text           string
	Card           string
	Link           string
	Shadow         string
	Button         string
	FontSize       float64
	HeaderFontSize float64
	BodyFontSize   float64
}
