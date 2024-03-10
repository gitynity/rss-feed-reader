package main

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Subscription represents a subscribed RSS feed.
type Subscription struct {
	Name          string
	URL           string
	LastReadIndex int // Index of the last read item
	LastReadTime  string
}

// State represents the overall state of the application.
type State struct {
	Subscriptions []Subscription
}

func saveState(state State) error {
	file, err := os.Create("rss_state.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(state)
}

func generateRandomState() State {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number of subscriptions (between 1 and 5)
	numSubscriptions := rand.Intn(5) + 1

	var subscriptions []Subscription

	for i := 1; i <= numSubscriptions; i++ {
		subscription := Subscription{
			Name:          fmt.Sprintf("Feed %d", i),
			URL:           generateRandomURL(),
			LastReadIndex: rand.Intn(10),
			LastReadTime:  generateRandomTime(),
		}
		subscriptions = append(subscriptions, subscription)
	}

	return State{Subscriptions: subscriptions}
}

func generateRandomURL() string {
	// Valid RSS feed URLs
	rssFeedURLs := []string{
		"http://feeds.bbci.co.uk/news/rss.xml",
		"https://rss.nytimes.com/services/xml/rss/nyt/World.xml",
		"https://www.nasa.gov/rss/dyn/breaking_news.rss",
		"http://www.espn.com/espn/rss/news",
		"https://github.blog/all.atom",
	}

	randIndex := rand.Intn(len(rssFeedURLs))
	return rssFeedURLs[randIndex]
}

func generateRandomTime() string {
	// Generate a random time within the last week
	minTime := time.Now().Add(-7 * 24 * time.Hour)
	maxTime := time.Now()

	randomTime := minTime.Add(time.Duration(rand.Int63n(maxTime.Unix()-minTime.Unix())) * time.Second)

	return randomTime.Format(time.RFC3339)
}

func main() {
	// Example of generating a random state with valid RSS feed URLs
	randomState := generateRandomState()

	// Save the state to rss_state.gob
	err := saveState(randomState)
	if err != nil {
		fmt.Println("Error saving state:", err)
	} else {
		fmt.Println("State saved successfully.")
	}
}
