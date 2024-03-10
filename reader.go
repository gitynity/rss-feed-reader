package main

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

// RSS represents the structure of an RSS feed.
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Channel represents the channel information in an RSS feed.
type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

// Item represents an item in an RSS feed.
type Item struct {
	Title       string  `xml:"title"`
	Description string  `xml:"description"`
	Link        string  `xml:"link"`
	PubDate     string  `xml:"pubDate"`
	Media       []Media `xml:"http://search.yahoo.com/mrss/ content"`
}

// Media represents media content in an RSS feed item.
type Media struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

// Subscription represents a subscribed RSS feed.
type Subscription struct {
	Name          string
	URL           string
	LastReadIndex int // Index of last read item
	LastReadTime  string
}

// State represents the overall state of the application.
type State struct {
	Subscriptions []Subscription
}

const (
	stateFileName = "rss_state.gob"
	downloadDir   = "downloads"
	htmlTemplate  = `
		<!DOCTYPE html>
		<html>
		<head>
			<title>RSS Feeds</title>
		</head>
		<body>
			<h1>RSS Feeds</h1>
			{{range .}}
				<h2>{{.Title}}</h2>
				<p>{{.Description}}</p>
				<ul>
					{{range .Items}}
						<li>
							<h3>{{.Title}}</h3>
							<p>{{.Description}}</p>
							<p>Published: {{.PubDate}}</p>
							<p>Status: {{.Status}}</p>
							{{if .HasMedia}}
								<img src="{{.MediaURL}}" alt="Media">
							{{end}}
						</li>
					{{end}}
				</ul>
			{{end}}
		</body>
		</html>
	`
)

// ItemData represents data for a feed item in the HTML template.
type ItemData struct {
	Title       string
	Description string
	Link        string
	PubDate     string
	Status      string
	HasMedia    bool
	MediaURL    string
}

// FeedData represents data for a feed in the HTML template.
type FeedData struct {
	Title       string
	Description string
	Link        string
	Items       []ItemData
}

func fetchRSS(url string) (*RSS, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return nil, err
	}

	return &rss, nil
}

func subscribe(subscriptions []Subscription, name, url string) []Subscription {
	subscription := Subscription{Name: name, URL: url}
	return append(subscriptions, subscription)
}

func saveState(state State) error {
	file, err := os.Create(stateFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(state)
}

func loadState() (State, error) {
	var state State

	file, err := os.Open(stateFileName)
	if err != nil {
		return state, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&state)

	return state, err
}

func downloadMedia(url, filename string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	return err
}

func renderHTML(feed RSS) string {
	var result strings.Builder
	writer := io.Writer(&result)

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		fmt.Println("Error parsing HTML template:", err)
		return ""
	}

	err = tmpl.Execute(writer, feed)
	if err != nil {
		fmt.Println("Error executing HTML template:", err)
		return ""
	}

	return result.String()
}
