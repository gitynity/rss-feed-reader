package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

func serveHTML(w http.ResponseWriter, r *http.Request) {
	// Load previous state
	state, err := loadState()
	if err != nil {
		http.Error(w, "Error loading state", http.StatusInternalServerError)
		return
	}

	// Fetch and display each subscribed feed
	var feeds []FeedData
	for _, subscription := range state.Subscriptions {
		rss, err := fetchRSS(subscription.URL)
		if err != nil {
			fmt.Printf("Error fetching RSS feed '%s': %v\n", subscription.Name, err)
			continue
		}

		feedData := FeedData{
			Title:       rss.Channel.Title,
			Description: rss.Channel.Description,
			Link:        rss.Channel.Link,
		}

		// Process each feed item
		for _, item := range rss.Channel.Items {
			itemData := ItemData{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				PubDate:     item.PubDate,
				Status:      "Unread",
			}

			// Check if the item was read
			if itemData.PubDate <= subscription.LastReadTime {
				itemData.Status = "Read"
			}

			// Add media information if available
			if len(item.Media) > 0 {
				itemData.HasMedia = true
				itemData.MediaURL = item.Media[0].URL
			}

			// Add item to feedData
			feedData.Items = append(feedData.Items, itemData)
		}

		// Add feedData to feeds
		feeds = append(feeds, feedData)
	}

	// Render HTML and serve to the client
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Error parsing HTML template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, feeds)
	if err != nil {
		http.Error(w, "Error executing HTML template", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Set up a simple web server
	http.HandleFunc("/", serveHTML)
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))

	// Load previous state
	state, err := loadState()
	if err != nil {
		fmt.Println("Error loading state:", err)
	}

	// Subscribe to a feed
	state.Subscriptions = subscribe(state.Subscriptions, "Example Feed", "https://example.com/rss-feed.xml")

	// Fetch and display each subscribed feed
	for i, subscription := range state.Subscriptions {
		rss, err := fetchRSS(subscription.URL)
		if err != nil {
			fmt.Printf("Error fetching RSS feed '%s': %v\n", subscription.Name, err)
			continue
		}

		// Print the feed details
		fmt.Printf("Feed Title: %s\n", rss.Channel.Title)
		fmt.Printf("Feed Description: %s\n", rss.Channel.Description)
		fmt.Printf("Feed Link: %s\n", rss.Channel.Link)

		// Print the items in the feed
		fmt.Println("\nItems:")
		for j, item := range rss.Channel.Items {
			fmt.Printf("Title: %s\n", item.Title)
			fmt.Printf("Description: %s\n", item.Description)
			fmt.Printf("Link: %s\n", item.Link)
			fmt.Printf("Published: %s\n", item.PubDate)

			// Check if the item was read
			if j <= subscription.LastReadIndex {
				fmt.Println("Status: Read")
			} else {
				fmt.Println("Status: Unread")
			}

			// Download media content
			for k, media := range item.Media {
				mediaFilename := filepath.Join(downloadDir, fmt.Sprintf("%s_media_%d%s", item.Title, k, filepath.Ext(media.URL)))
				err := downloadMedia(media.URL, mediaFilename)
				if err != nil {
					fmt.Printf("Error downloading media for '%s': %v\n", item.Title, err)
				} else {
					fmt.Printf("Media downloaded for '%s': %s\n", item.Title, mediaFilename)
				}
			}
		}

		// Update the last read index for the subscription
		state.Subscriptions[i].LastReadIndex = len(rss.Channel.Items) - 1
		state.Subscriptions[i].LastReadTime = time.Now().Format(time.RFC3339)
	}

	// Save the updated state
	err = saveState(state)
	if err != nil {
		fmt.Println("Error saving state:", err)
	}

	// Start the web server
	fmt.Println("Server is running on http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting web server:", err)
	}
}

