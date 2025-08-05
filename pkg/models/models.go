package models

import "time"

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Highlight struct {
	ID            int        `json:"id"`
	Text          string     `json:"text"`
	Note          string     `json:"note"`
	Location      int        `json:"location"`
	LocationType  string     `json:"location_type"`
	HighlightedAt *time.Time `json:"highlighted_at"`
	URL           string     `json:"url"`
	Color         string     `json:"color"`
	Updated       time.Time  `json:"updated"`
	BookID        int        `json:"book_id"`
	Tags          []Tag      `json:"tags"`
	IsFavorite    bool       `json:"is_favorite"`
	IsDiscard     bool       `json:"is_discard"`
	ReadwiseURL   string     `json:"readwise_url"`
}

type HighlightList struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Results  []Highlight `json:"results"`
}

type Book struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Author          string     `json:"author"`
	Category        string     `json:"category"`
	Source          string     `json:"source"`
	NumHighlights   int        `json:"num_highlights"`
	LastHighlightAt *time.Time `json:"last_highlight_at"`
	Updated         time.Time  `json:"updated"`
	CoverImageURL   string     `json:"cover_image_url"`
	HighlightsURL   string     `json:"highlights_url"`
	SourceURL       string     `json:"source_url"`
	ASIN            string     `json:"asin"`
	Tags            []Tag      `json:"tags"`
	DocumentNote    string     `json:"document_note"`
}

type BookList struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []Book `json:"results"`
}

type HighlightUpdate struct {
	Text     string `json:"text,omitempty"`
	Note     string `json:"note,omitempty"`
	Location int    `json:"location,omitempty"`
	URL      string `json:"url,omitempty"`
	Color    string `json:"color,omitempty"`
}
