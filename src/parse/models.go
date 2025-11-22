package parse

import (
	"html/template"
	"time"
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
	InputPath                 string
	OutputPath                string
	DateFormat                string
	IndexName                 string
	Theme                     string
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

	// AuthorName is used in meta tags and structured data as the article author.
	AuthorName string
	// PublisherName is used in structured data as the publisher name.
	PublisherName string
	// PublisherLogoPath is an optional path (relative to site root) to a logo image used
	// in structured data as publisher.logo.
	PublisherLogoPath string
}

// ShareButton describes a single social or custom share target.
type ShareButton struct {
	Name        string
	Display     string
	UrlTemplate string
}

// Article models a single article or page, regardless of its original format.
type Article struct {
	Title        string
	Description  string
	CoverImage   string
	Created      time.Time
	Updated      time.Time
	Tags         []string
	TextContent  string
	HtmlContent  string
	OriginalPath string
	LinkToSelf   string
	LinkToSave   string
	ShareUrl     string
	CanonicalUrl string
}
